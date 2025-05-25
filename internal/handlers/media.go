package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
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

	// Parse multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		err = app_errors.AppError("Unknown error", app_errors.StatusValidationError)
		logger.LogError(log, err, "Error parsing multipart form")
		_ = ctx.Error(err)
		return
	}

	// Check if any files were uploaded
	files := form.File["upload-images"]
	if len(files) == 0 {
		err = app_errors.AppError("There are no images uploaded", app_errors.StatusValidationError)
		logger.LogError(log, err, "No images uploaded")
		_ = ctx.Error(err)
		return
	}

	// Create working directory if it doesn't exist
	dir, err := os.Getwd()
	if err != nil {
		err = app_errors.AppError("Unknown error", app_errors.StatusValidationError)
		logger.LogError(log, err, "Error getting current directory")
		_ = ctx.Error(err)
		return
	}
	uploadDir := filepath.Join(dir, "image-storage")
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			logger.LogError(log, err, "Failed to create upload directory")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}
	}

	// Setup concurrency primitives
	var wg sync.WaitGroup
	errorChan := make(chan error, len(files))
	duplicateFiles := make([]int, 0)
	mu := &sync.Mutex{} // Mutex for thread-safe append operations

	// Step 1: Check for duplicates
	for index, file := range files {
		wg.Add(1)
		go func(index int, file *multipart.FileHeader) {
			defer wg.Done()

			// Create temp file path for hash checking
			fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
			path := filepath.Join(uploadDir, fileName)

			// Save file temporarily
			if err = ctx.SaveUploadedFile(file, path); err != nil {
				errorChan <- app_errors.AppError("Failed to save uploaded file", app_errors.StatusValidationError)
				logger.LogError(log, err, "Error saving uploaded file")
				return
			}
			defer os.Remove(path) // Clean up temp file after check

			// Generate hash for duplicate check
			fileHash, err := m.mediaService.GenerateFileHash(path)
			if err != nil {
				err = app_errors.AppError("Failed to generate file hash", app_errors.StatusValidationError)
				logger.LogError(log, err, "Error generating file hash")
				errorChan <- err
				return
			}

			// Check for duplicates
			isDuplicate, err := m.mediaService.IsDuplicate(ctx, fileHash)
			if err != nil {
				errorChan <- err
				return
			}

			if isDuplicate {
				mu.Lock()
				duplicateFiles = append(duplicateFiles, index+1)
				mu.Unlock()
			}
		}(index, file)
	}

	// Wait for all duplicate checks to complete
	wg.Wait()
	close(errorChan)

	// Handle errors from duplicate check phase
	if len(errorChan) > 0 {
		for err := range errorChan {
			logger.LogError(log, err, "Error processing images")
			_ = ctx.Error(err)
			return
		}
	}

	// If duplicates found, return error
	if len(duplicateFiles) > 0 {
		sort.Ints(duplicateFiles)
		duplicateIndices := make([]string, len(duplicateFiles))
		for i, v := range duplicateFiles {
			duplicateIndices[i] = fmt.Sprintf("%d", v)
		}
		duplicateMessage := fmt.Sprintf("Duplicate image: %s", strings.Join(duplicateIndices, ","))
		err = app_errors.AppError(duplicateMessage, app_errors.StatusValidationError)
		logger.LogError(log, err, "Duplicate images found")
		_ = ctx.Error(err)
		return
	}

	// Step 2: Process and upload images if no duplicates
	resultChan := make(chan *model.MediaInformationResponse, len(files))
	errorChan = make(chan error, len(files)) // Reset error channel

	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()

			// Get file size
			fileSizeBytes := file.Size
			fileSizeMB := float64(fileSizeBytes) / (1024 * 1024)

			// Open file to read image metadata
			imgFile, err := file.Open()
			if err != nil {
				err = app_errors.AppError("Error opening image file", app_errors.StatusValidationError)
				logger.LogError(log, err, "Failed to open image file")
				errorChan <- err
				return
			}
			defer imgFile.Close()

			// Read image configuration (dimensions, format)
			img, format, err := image.DecodeConfig(imgFile)
			if err != nil {
				err = app_errors.AppError("Error decoding image", app_errors.StatusValidationError)
				logger.LogError(log, err, "Failed to decode image")
				errorChan <- err
				return
			}

			// Reset file pointer for upload
			imgFile.Close()
			imgFile, err = file.Open()
			if err != nil {
				err = app_errors.AppError("Error reopening image file", app_errors.StatusValidationError)
				errorChan <- err
				return
			}
			defer imgFile.Close()

			// Create S3 uploader
			uploader, err := s3_storage.NewS3Uploader()
			if err != nil {
				err = app_errors.AppError("Error creating S3 uploader", app_errors.StatusValidationError)
				logger.LogError(log, err, "Failed to create S3 uploader")
				errorChan <- err
				return
			}

			// Upload file to S3
			url, err := uploader.UploadFile(imgFile, file)
			if err != nil {
				err = app_errors.AppError("Error uploading to S3", app_errors.StatusValidationError)
				logger.LogError(log, err, "Failed to upload file to S3")
				errorChan <- err
				return
			}

			// Create temporary file for hash generation
			tempPath := filepath.Join(uploadDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename))
			if err = ctx.SaveUploadedFile(file, tempPath); err != nil {
				errorChan <- app_errors.AppError("Error saving temporary file", app_errors.StatusValidationError)
				return
			}

			// Generate hash for database storage
			mediaHash, err := m.mediaService.GenerateFileHash(tempPath)
			if err != nil {
				err = app_errors.AppError("Error generating file hash", app_errors.StatusValidationError)
				logger.LogError(log, err, "Failed to generate file hash")
				errorChan <- err
				os.Remove(tempPath) // Clean up temp file
				return
			}
			os.Remove(tempPath) // Clean up temp file

			// Create media database record
			mediaData := model.Media{
				MediaURL:    url,
				MediaFormat: format,
				MediaSize:   fileSizeMB,
				MediaWidth:  img.Width,
				MediaHeight: img.Height,
				MediaHash:   mediaHash,
			}

			// Save media record to database
			rs, err := m.mediaService.SaveMedia(ctx, mediaData)
			if err != nil {
				errorChan <- err
				return
			}

			resultChan <- &rs.Data
		}(file)
	}

	// Wait for all uploads to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Handle any errors from upload phase
	if len(errorChan) > 0 {
		for err := range errorChan {
			logger.LogError(log, err, "Error processing images")
			_ = ctx.Error(err)
			return
		}
	}

	// Collect successful uploads
	var mediaList []*model.MediaInformationResponse
	for res := range resultChan {
		mediaList = append(mediaList, res)
	}

	// Return response
	response := model.ListImageSaveResponse{
		Meta: utils.NewMetaData(ctx),
		Data: mediaList,
	}

	ctx.JSON(http.StatusOK, response)
}
