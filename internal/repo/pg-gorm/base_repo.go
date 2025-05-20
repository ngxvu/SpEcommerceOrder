package pg_gorm

import (
	"context"
	"emission/internal/models"
	"gorm.io/gorm"
	"time"
)

//go:generate mockgen -source base_repo.go -destination mocks/base_repo.go

const (
	generalQueryTimeout = 30 * time.Second
)

type RepoPG struct {
	DB    *gorm.DB
	debug bool
}

func (r *RepoPG) GetRepo() *gorm.DB {
	return r.DB
}

func NewRepo(db *gorm.DB) IRepo {
	return &RepoPG{DB: db}
}

type IRepo interface {
	GetRepo() *gorm.DB
	CreateFactory(ctx context.Context, ob *models.Factory) (*models.Factory, error)
	FilterEmission(ctx context.Context, f *models.EmissionFilter) (*models.EmissionFilterResult, error)
	GetEmissionFactorByCountryCode(ctx context.Context, countryCode string) (*models.EmissionFactor, error)
}
