package services

import (
	"context"
	"emission/conf"
	"emission/internal/models"
	repo "emission/internal/repo/pg-gorm"
	"emission/internal/utils"
	"emission/pkg/http/ginext"
	"emission/pkg/http/logger"
	"github.com/volatiletech/null/v8"
	"net/http"
)

type FactoryService struct {
	repo        repo.IRepo
	authService AuthInterface
}

func NewFactoryService(sqlBoilerRepo repo.IRepo, authService AuthInterface) FactoryInterface {
	return &FactoryService{repo: sqlBoilerRepo, authService: authService}
}

type FactoryInterface interface {
	CreateFactory(ctx context.Context, req models.FactoryRequest) (res *models.CreateFactoryResponse, err error)
}

func (s FactoryService) CreateFactory(ctx context.Context, req models.FactoryRequest) (res *models.CreateFactoryResponse, err error) {
	log := logger.WithCtx(ctx, "FactoryService.GetEmissionFactorByFactory")

	if err := utils.CheckRequireValid(req); err != nil {
		log.WithError(err).Error("Error when check require valid")
		return nil, ginext.NewError(http.StatusBadRequest, err.Error())
	}

	ob := &models.Factory{
		Name:        *req.Name,
		CountryCode: *req.CountryCode,
	}

	// validate country_code exist in emission_factor table
	_, err = s.repo.GetEmissionFactorByCountryCode(ctx, *req.CountryCode)
	if err != nil {
		log.WithError(err).
			WithField("req", req).
			Error("Error when call func GetEmissionFactorByCountryCode")
		return nil, ginext.NewError(http.StatusBadRequest, "Invalid country code, only support country code in emission_factor, default is 'VN','JP','US'")
	}

	rs, err := s.repo.CreateFactory(ctx, ob)
	if err != nil {
		log.WithError(err).
			WithField("req", req).
			Error("Error when call func CreateFactory")
		return nil, ginext.NewError(http.StatusInternalServerError, utils.MessageError()[http.StatusInternalServerError])
	}

	// one time use token,
	apiKey, err := s.authService.CreateAccessToken(ctx, models.CreateTokenRequest{
		ObjectID: rs.ID,
		NumHour:  conf.LoadEnv().JWTExpire,
	})
	if err != nil {
		log.WithError(err).
			WithField("req", req).
			Error("Error when call func CreateAccessToken")
		return nil, ginext.NewError(http.StatusInternalServerError, utils.MessageError()[http.StatusInternalServerError])
	}

	return &models.CreateFactoryResponse{
		Name:        rs.Name.String,
		CountryCode: rs.CountryCode.String,
		ApiKey:      apiKey,
	}, nil
}
