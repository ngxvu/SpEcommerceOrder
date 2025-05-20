package pg_gorm

import (
	"context"
	"emission/internal/models"
)

func (r *RepoPG) CreateFactory(ctx context.Context, ob *models.Factory) (*models.Factory, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RepoPG) FilterEmission(ctx context.Context, f *models.EmissionFilter) (*models.EmissionFilterResult, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RepoPG) GetEmissionFactorByCountryCode(ctx context.Context, countryCode string) (*models.EmissionFactor, error) {
	//TODO implement me
	panic("implement me")
}
