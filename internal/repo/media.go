package repo

import (
	"context"
	"gorm.io/gorm"
	model "kimistore/internal/models"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
)

type MediaRepository struct {
}

func NewMediaRepository() *MediaRepository {
	return &MediaRepository{}
}

type MediaRepositoryInterface interface {
	SaveMedia(media model.Media, tx *gorm.DB, ctx context.Context) (*model.ImageSaveResponse, error)
	IsDuplicate(hash string, tx *gorm.DB) (bool, error)
}

func (r *MediaRepository) SaveMedia(media model.Media, tx *gorm.DB, ctx context.Context) (*model.ImageSaveResponse, error) {

	log := logger.WithTag("MediaRepository|SaveMedia")

	err := tx.Create(&media).Error
	if err != nil {
		err := app_errors.AppError("Error saving image to database", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Error saving image to database")
		return nil, err
	}

	mediaID := utils.UUIDtoString(media.ID)

	mediaInformationResponse := model.MediaInformationResponse{
		ID:          mediaID,
		MediaURL:    media.MediaURL,
		MediaFormat: media.MediaFormat,
		MediaSize:   media.MediaSize,
		MediaWidth:  media.MediaWidth,
		MediaHeight: media.MediaHeight,
	}

	rs := model.ImageSaveResponse{
		Meta: utils.NewMetaData(ctx),
		Data: mediaInformationResponse,
	}

	return &rs, nil
}

func (r *MediaRepository) IsDuplicate(hash string, tx *gorm.DB) (bool, error) {
	var media model.Media
	if err := tx.Where("media_hash = ?", hash).First(&media).Error; err != nil {
		if gorm.ErrRecordNotFound == err {
			return false, nil // No duplicate found
		}
		return false, err // An error occurred
	}
	return true, nil // Duplicate found
}
