package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	model "kimistore/internal/models"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/services"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/service/s3_storage"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type MediaHandler struct {
	db           pgGorm.PGInterface
	mediaService services.MediaServiceInterface
}

func NewMediaHandler(
	pgRepo pgGorm.PGInterface,
	mediaService services.MediaServiceInterface) *MediaHandler {
	return &MediaHandler{
		db:           pgRepo,
		mediaService: mediaService,
	}
}

func (m *MediaHandler) UploadListImage(ctx *gin.Context) {
	log := logger.WithTag("MediaHandler|UploadListImage")

	// Parse and validate form
	files, err := m.parseAndValidateForm(ctx, log)
	if err != nil {
		return
	}

	// Prepare upload directory
	uploadDir, err := m.prepareUploadDirectory(ctx, log)
	if err != nil {
		return
	}

	// Check for duplicates
	duplicateFiles, err := m.checkForDuplicates(ctx, files, uploadDir, log)
	if err != nil {
		return
	}

	// Handle duplicate files if any
	if len(duplicateFiles) > 0 {
		m.handleDuplicateFiles(ctx, duplicateFiles, log)
		return
	}

	// Process and upload images
	mediaList, err := m.processAndUploadImages(ctx, files, uploadDir, log)
	if err != nil {
		return
	}

	// Return response
	response := model.ListImageSaveResponse{
		Meta: utils.NewMetaData(ctx),
		Data: mediaList,
	}

	ctx.JSON(http.StatusOK, response)
}

// parseAndValidateForm validates the multipart form and returns the uploaded files
func (m *MediaHandler) parseAndValidateForm(ctx *gin.Context, log *logrus.Entry) ([]*multipart.FileHeader, error) {
	form, err := ctx.MultipartForm()
	if err != nil {
		err = app_errors.AppError("Unknown error", app_errors.StatusValidationError)
		logger.LogError(log, err, "Error parsing multipart form")
		_ = ctx.Error(err)
		return nil, err
	}

	files := form.File["upload-images"]
	if len(files) == 0 {
		err = app_errors.AppError("There are no images uploaded", app_errors.StatusValidationError)
		logger.LogError(log, err, "No images uploaded")
		_ = ctx.Error(err)
		return nil, err
	}

	return files, nil
}

// prepareUploadDirectory creates the temporary directory for file processing
func (m *MediaHandler) prepareUploadDirectory(ctx *gin.Context, log *logrus.Entry) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		err = app_errors.AppError("Unknown error", app_errors.StatusValidationError)
		logger.LogError(log, err, "Error getting current directory")
		_ = ctx.Error(err)
		return "", err
	}

	uploadDir := filepath.Join(dir, "image-storage")
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			logger.LogError(log, err, "Failed to create upload directory")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return "", err
		}
	}

	return uploadDir, nil
}

// checkForDuplicates checks if any of the files already exist in the database
func (m *MediaHandler) checkForDuplicates(ctx *gin.Context, files []*multipart.FileHeader, uploadDir string, log *logrus.Entry) ([]int, error) {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(files))
	duplicateFiles := make([]int, 0)
	mu := &sync.Mutex{}

	for index, file := range files {
		wg.Add(1)
		go func(index int, file *multipart.FileHeader) {
			defer wg.Done()

			// Create temp file path for hash checking
			fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
			path := filepath.Join(uploadDir, fileName)

			// Save file temporarily
			if err := ctx.SaveUploadedFile(file, path); err != nil {
				errorChan <- app_errors.AppError("Failed to save uploaded file", app_errors.StatusValidationError)
				logger.LogError(log, err, "Error saving uploaded file")
				return
			}
			defer os.Remove(path) // Clean up temp file after check

			// Generate hash for duplicate check
			fileHash, err := m.mediaService.GenerateFileHash(path)
			if err != nil {
				errorChan <- app_errors.AppError("Failed to generate file hash", app_errors.StatusValidationError)
				logger.LogError(log, err, "Error generating file hash")
				return
			}

			// Check for duplicates
			isDuplicate, err := m.mediaService.IsDuplicate(ctx, fileHash)
			if err != nil {
				errorChan <- err
				logger.LogError(log, err, "Error checking for duplicates")
				return
			}

			if isDuplicate {
				mu.Lock()
				duplicateFiles = append(duplicateFiles, index+1)
				mu.Unlock()
			}
		}(index, file)
	}

	wg.Wait()
	close(errorChan)

	// Handle errors from duplicate check phase
	if len(errorChan) > 0 {
		for err := range errorChan {
			logger.LogError(log, err, "Error processing images")
			_ = ctx.Error(err)
			return nil, err
		}
	}

	return duplicateFiles, nil
}

// handleDuplicateFiles returns an error response for duplicate files
func (m *MediaHandler) handleDuplicateFiles(ctx *gin.Context, duplicateFiles []int, log *logrus.Entry) {
	sort.Ints(duplicateFiles)
	duplicateIndices := make([]string, len(duplicateFiles))
	for i, v := range duplicateFiles {
		duplicateIndices[i] = fmt.Sprintf("%d", v)
	}
	duplicateMessage := fmt.Sprintf("Duplicate image: %s", strings.Join(duplicateIndices, ","))
	err := app_errors.AppError(duplicateMessage, app_errors.StatusValidationError)
	logger.LogError(log, err, "Duplicate images found")
	_ = ctx.Error(err)
}

// processAndUploadImages processes and uploads images to S3 and saves to database
func (m *MediaHandler) processAndUploadImages(ctx *gin.Context, files []*multipart.FileHeader, uploadDir string, log *logrus.Entry) ([]*model.MediaInformationResponse, error) {
	var wg sync.WaitGroup
	resultChan := make(chan *model.MediaInformationResponse, len(files))
	errorChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()

			// Process single image
			result, err := m.processSingleImage(ctx, file, uploadDir, log)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- result
		}(file)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Handle any errors from upload phase
	if len(errorChan) > 0 {
		for err := range errorChan {
			logger.LogError(log, err, "Error processing images")
			_ = ctx.Error(err)
			return nil, err
		}
	}

	// Collect successful uploads
	var mediaList []*model.MediaInformationResponse
	for res := range resultChan {
		mediaList = append(mediaList, res)
	}

	return mediaList, nil
}

// processSingleImage processes a single image file
func (m *MediaHandler) processSingleImage(ctx *gin.Context, file *multipart.FileHeader, uploadDir string, log *logrus.Entry) (*model.MediaInformationResponse, error) {
	// Get file size
	fileSizeBytes := file.Size
	fileSizeMB := float64(fileSizeBytes) / (1024 * 1024)

	// Extract image metadata
	img, format, err := m.extractImageMetadata(file, log)
	if err != nil {
		return nil, err
	}

	// Upload to S3
	url, err := m.uploadToS3(file, log)
	if err != nil {
		return nil, err
	}

	// Generate hash for database storage
	mediaHash, err := m.generateMediaHash(ctx, file, uploadDir, log)
	if err != nil {
		return nil, err
	}

	mediaData := model.Media{
		MediaURL:    url,
		MediaFormat: format,
		MediaSize:   fileSizeMB,
		MediaWidth:  img.Width,
		MediaHeight: img.Height,
		MediaHash:   mediaHash,
	}

	// Create and save media record
	return m.saveMediaRecord(ctx, mediaData, log)
}

// extractImageMetadata extracts width, height and format from image
func (m *MediaHandler) extractImageMetadata(file *multipart.FileHeader, log *logrus.Entry) (image.Config, string, error) {
	imgFile, err := file.Open()
	if err != nil {
		err = app_errors.AppError("Error opening image file", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to open image file")
		return image.Config{}, "", err
	}
	defer imgFile.Close()

	// Read image configuration (dimensions, format)
	img, format, err := image.DecodeConfig(imgFile)
	if err != nil {
		err = app_errors.AppError("Error decoding image", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to decode image")
		return image.Config{}, "", err
	}

	return img, format, nil
}

// uploadToS3 uploads file to S3 storage
func (m *MediaHandler) uploadToS3(file *multipart.FileHeader, log *logrus.Entry) (string, error) {
	imgFile, err := file.Open()
	if err != nil {
		err = app_errors.AppError("Error reopening image file", app_errors.StatusValidationError)
		return "", err
	}
	defer imgFile.Close()

	// Create S3 uploader
	uploader, err := s3_storage.NewS3Uploader()
	if err != nil {
		err = app_errors.AppError("Error creating S3 uploader", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to create S3 uploader")
		return "", err
	}

	// Upload file to S3
	url, err := uploader.UploadFile(imgFile, file)
	if err != nil {
		err = app_errors.AppError("Error uploading to S3", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to upload file to S3")
		return "", err
	}

	return url, nil
}

// generateMediaHash creates a hash for the media file
func (m *MediaHandler) generateMediaHash(ctx *gin.Context, file *multipart.FileHeader, uploadDir string, log *logrus.Entry) (string, error) {
	tempPath := filepath.Join(uploadDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename))
	if err := ctx.SaveUploadedFile(file, tempPath); err != nil {
		return "", app_errors.AppError("Error saving temporary file", app_errors.StatusValidationError)
	}
	defer os.Remove(tempPath)

	// Generate hash for database storage
	mediaHash, err := m.mediaService.GenerateFileHash(tempPath)
	if err != nil {
		err = app_errors.AppError("Error generating file hash", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to generate file hash")
		return "", err
	}

	return mediaHash, nil
}

// saveMediaRecord saves the media record to the database
func (m *MediaHandler) saveMediaRecord(ctx *gin.Context, media model.Media, log *logrus.Entry) (*model.MediaInformationResponse, error) {

	// Save media record to database
	rs, err := m.mediaService.SaveMedia(ctx, media)
	if err != nil {
		logger.LogError(log, err, "Failed to save media record")
		return nil, err
	}

	return &rs.Data, nil
}
