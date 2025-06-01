package repo

import (
	"context"
	"gorm.io/gorm"
	model "kimistore/internal/models"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/utils"
)

type MediaRepository struct {
	db pgGorm.PGInterface
}

func NewMediaRepository(newPgRepo pgGorm.PGInterface) *MediaRepository {
	return &MediaRepository{
		db: newPgRepo,
	}
}

type MediaRepositoryInterface interface {
	SaveMedia(ctx context.Context, tx *gorm.DB, media model.Media) (*model.ImageSaveResponse, error)
}

func (r *MediaRepository) SaveMedia(ctx context.Context, tx *gorm.DB, media model.Media) (*model.ImageSaveResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = r.db.DBWithTimeout(ctx)
		defer cancel()
	}

	err := tx.WithContext(ctx).Create(&media).Error
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
