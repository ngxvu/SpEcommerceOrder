package repo

import (
	"context"
	"gorm.io/gorm"
	model "kimistore/internal/models"
	"kimistore/internal/utils"
)

type MediaRepository struct {
}

func NewMediaRepository() *MediaRepository {
	return &MediaRepository{}
}

type MediaRepositoryInterface interface {
	SaveMedia(media model.Media, tx *gorm.DB, ctx context.Context) (*model.ImageSaveResponse, error)
}

func (r *MediaRepository) SaveMedia(media model.Media, tx *gorm.DB, ctx context.Context) (*model.ImageSaveResponse, error) {

	err := tx.Create(&media).Error
	if err != nil {
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
