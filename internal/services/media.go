package services

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"image"
	model "kimistore/internal/models"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/service/s3_storage"
	"mime/multipart"
	"sync"
)

type MediaService struct {
	repo      repo.MediaRepositoryInterface
	newPgRepo pgGorm.PGInterface
}

type MediaServiceInterface interface {
	ProcessAndUploadImages(ctx *gin.Context, files []*multipart.FileHeader) ([]*model.MediaInformationResponse, error)
}

func NewMediaService(repo repo.MediaRepositoryInterface, newRepo pgGorm.PGInterface) *MediaService {
	return &MediaService{
		repo:      repo,
		newPgRepo: newRepo,
	}
}

// processAndUploadImages processes and uploads images to S3 and saves to database
func (s *MediaService) ProcessAndUploadImages(ctx *gin.Context, files []*multipart.FileHeader) ([]*model.MediaInformationResponse, error) {

	getContext := ctx.Request.Context()

	log := logger.WithTag("MediaService|ProcessAndUploadImages")

	var wg sync.WaitGroup
	resultChan := make(chan *model.MediaInformationResponse, len(files))
	errorChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()

			// Process single image
			result, err := s.processSingleImage(getContext, file)
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
func (s *MediaService) processSingleImage(ctx context.Context, file *multipart.FileHeader) (*model.MediaInformationResponse, error) {

	log := logger.WithTag("MediaService|ProcessSingleImage")

	fileSizeBytes := file.Size
	fileSizeMB := float64(fileSizeBytes) / (1024 * 1024)

	// Extract image metadata
	img, format, err := s.extractImageMetadata(file)
	if err != nil {
		return nil, err
	}

	// Upload to S3
	url, err := s.uploadToS3(file, log)
	if err != nil {
		return nil, err
	}

	media := model.Media{
		MediaURL:    *url,
		MediaFormat: *format,
		MediaSize:   fileSizeMB,
		MediaWidth:  img.Width,
		MediaHeight: img.Height,
	}

	saveMedia, err := s.saveMedia(ctx, media)
	if err != nil {
		return nil, err
	}

	// Create and save media record
	return &saveMedia.Data, nil
}

func (s *MediaService) saveMedia(ctx context.Context, media model.Media) (*model.ImageSaveResponse, error) {

	log := logger.WithTag("MediaService|SaveMedia")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	savedMedia, err := s.repo.SaveMedia(media, tx, ctx)
	if err != nil {
		logger.LogError(log, err, "Failed to save media")
		err = app_errors.AppError("Failed to save images", app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return savedMedia, nil
}

// extractImageMetadata extracts width, height and format from image
func (s *MediaService) extractImageMetadata(file *multipart.FileHeader) (*image.Config, *string, error) {

	log := logger.WithTag("MediaService|ExtractImageMetadata")

	imgFile, err := file.Open()
	if err != nil {
		logger.LogError(log, err, "Failed to open image file")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, nil, err
	}
	defer imgFile.Close()

	// Read image configuration (dimensions, format)
	img, format, err := image.DecodeConfig(imgFile)
	if err != nil {
		logger.LogError(log, err, "Failed to decode image configuration")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, nil, err
	}

	return &img, &format, nil
}

// uploadToS3 uploads file to S3 storage
func (s *MediaService) uploadToS3(file *multipart.FileHeader, log *logrus.Entry) (*string, error) {

	log = logger.WithTag("MediaService|UploadToS3")

	imgFile, err := file.Open()
	if err != nil {
		logger.LogError(log, err, "Failed to open image file for upload")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	defer imgFile.Close()

	// Create S3 uploader
	uploader, err := s3_storage.NewS3Uploader()
	if err != nil {
		logger.LogError(log, err, "Failed to create S3 uploader")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	// Upload file to S3
	url, err := uploader.UploadFile(imgFile, file)
	if err != nil {
		logger.LogError(log, err, "Failed to upload image")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	return &url, nil
}
