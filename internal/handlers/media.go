package handlers

import (
	model "basesource/internal/models"
	pgGorm "basesource/internal/repo/pg-gorm"
	"basesource/internal/services"
	"basesource/internal/utils"
	"basesource/internal/utils/app_errors"
	"github.com/gin-gonic/gin"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
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

	form, err := ctx.MultipartForm()
	if err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	files := form.File["upload-images"]
	if len(files) == 0 {
		err = app_errors.AppError("There are no images uploaded", app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	// Process and upload images
	mediaList, err := m.mediaService.ProcessAndUploadImages(ctx, files)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// Return response
	response := model.ListImageSaveResponse{
		Meta: utils.NewMetaData(ctx),
		Data: mediaList,
	}

	ctx.JSON(http.StatusOK, response)
}
