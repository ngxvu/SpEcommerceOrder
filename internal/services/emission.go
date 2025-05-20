package services

import (
	"context"
	models "emission/internal/models"
	pgsqlboiler "emission/internal/repo/pg-gorm"
	"emission/internal/utils"
	"emission/pkg/http/ginext"
	"emission/pkg/http/logger"
	"encoding/json"
	"net/http"
)

//go:generate mockgen -source emission.go -destination mocks/emission.go
//go:generate mockgen -source factory.go -destination mocks/factory.go
//go:generate mockgen -source auth.go -destination mocks/auth.go

type EmissionService struct {
	repo pgsqlboiler.IRepo
}

func NewEmissionService(sqlBoilerRepo pgsqlboiler.IRepo) EmissionInterface {
	return &EmissionService{repo: sqlBoilerRepo}
}

type EmissionInterface interface {
	GetListEmission(ctx context.Context, req *models.EmissionFilter) (res *models.EmissionFilterResult, err error)
	CreateEmission(ctx context.Context, req models.EmissionCreateRequest) (res *models.Emission, err error)
}

func (s EmissionService) GetListEmission(ctx context.Context, req *models.EmissionFilter) (res *models.EmissionFilterResult, err error) {
	log := logger.WithCtx(ctx, "EmissionService.GetListEmission")
	result, err := s.repo.FilterEmission(ctx, req)
	if err != nil {
		log.WithError(err).
			WithField("req", req).
			Error("Error when call func FilterEmission")
		return nil, ginext.NewError(http.StatusInternalServerError, utils.MessageError()[http.StatusInternalServerError])
	}

	return result, nil
}

func (s EmissionService) CreateEmission(ctx context.Context, req models.EmissionCreateRequest) (res *models.Emission, err error) {
	log := logger.WithCtx(ctx, "EmissionService.CalculateScope2CarbonEmission")

	if err := utils.CheckRequireValid(req); err != nil {
		log.WithError(err).Error("Error when check require valid")
		return nil, ginext.NewError(http.StatusBadRequest, err.Error())
	}

	emissionData, err := json.Marshal(req)
	if err != nil {
		log.WithError(err).Error("Error when marshal request")
		return nil, ginext.NewError(http.StatusInternalServerError, utils.MessageError()[http.StatusInternalServerError])
	}

	return res, nil
}
