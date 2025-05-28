package repo

import (
	"context"
	"gorm.io/gorm"
	model "kimistore/internal/models"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/paging"
)

type TestimonialRepository struct {
}

func NewTestimonialRepository() *TestimonialRepository {
	return &TestimonialRepository{}
}

type TestimonialRepositoryInterface interface {
	CreateTestimonial(ctx context.Context, tx *gorm.DB, testimonial model.Testimonial) (*model.GetTestimonialResponse, error)
	GetDetailTestimonial(ctx context.Context, tx *gorm.DB, id string) (*model.GetTestimonialResponse, error)
	GetListTestimonial(filter *paging.Filter, tx *gorm.DB) (*model.ListTestimonialResponse, error)
	UpdateTestimonial(ctx context.Context, tx *gorm.DB, id string, Testimonial model.Testimonial) (*model.GetTestimonialResponse, error)
	DeleteTestimonial(ctx context.Context, tx *gorm.DB, id string) (*model.DeleteTestimonialResponse, error)
}

func (r *TestimonialRepository) CreateTestimonial(ctx context.Context,
	tx *gorm.DB, testimonial model.Testimonial) (*model.GetTestimonialResponse, error) {
	log := logger.WithTag("TestimonialRepository|CreateTestimonial")

	if err := tx.Create(&testimonial).Error; err != nil {
		err = app_errors.AppError("Failed to create testimonial", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Failed to create testimonial")
		return nil, err
	}

	response := &model.GetTestimonialResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &testimonial,
	}

	return response, nil
}

func (r *TestimonialRepository) GetDetailTestimonial(ctx context.Context,
	tx *gorm.DB, id string) (*model.GetTestimonialResponse, error) {
	log := logger.WithTag("TestimonialRepository|GetDetailTestimonial")

	var testimonial model.Testimonial
	if err := tx.Where("id = ?", id).First(&testimonial).Error; err != nil {
		err = app_errors.AppError("Testimonial not found", app_errors.StatusNotFound)
		logger.LogError(log, err, "Testimonial not found")
		return nil, err
	}

	response := &model.GetTestimonialResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &testimonial,
	}

	return response, nil
}

func (r *TestimonialRepository) GetListTestimonial(filter *paging.Filter, pgRepo *gorm.DB) (*model.ListTestimonialResponse, error) {
	log := logger.WithTag("TestimonialRepository|GetListTestimonial")

	tx := pgRepo.Model(&model.Testimonial{})

	result := &model.ListTestimonialResponse{
		Filter:  filter,
		Records: []model.Testimonial{},
	}

	filter.Pager.SortableFields = []string{"author"}

	pager := filter.Pager

	err := pager.DoQuery(&result.Records, tx).Error
	if err != nil {
		err := app_errors.AppError("Error getting list testimonial", app_errors.StatusNotFound)
		logger.LogError(log, err, "Error when getting list internal price group")
		return nil, err

	}

	response := &model.ListTestimonialResponse{
		Filter:  filter,
		Records: result.Records,
	}

	return response, nil
}

func (r *TestimonialRepository) UpdateTestimonial(ctx context.Context, tx *gorm.DB, id string, testimonial model.Testimonial) (*model.GetTestimonialResponse, error) {
	log := logger.WithTag("TestimonialRepository|UpdateTestimonial")

	if err := tx.Model(&model.Testimonial{}).Where("id = ?", id).Updates(testimonial).Error; err != nil {
		err = app_errors.AppError("Failed to update testimonial", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Failed to update testimonial")
		return nil, err
	}

	response := &model.GetTestimonialResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &testimonial,
	}

	return response, nil
}

func (r *TestimonialRepository) DeleteTestimonial(ctx context.Context, tx *gorm.DB, id string) (*model.DeleteTestimonialResponse, error) {
	log := logger.WithTag("TestimonialRepository|DeleteTestimonial")

	if err := tx.Where("id = ?", id).Delete(&model.Testimonial{}).Error; err != nil {
		err = app_errors.AppError("Failed to delete testimonial", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Failed to delete testimonial")
		return nil, err
	}

	response := &model.DeleteTestimonialResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &id,
	}

	return response, nil
}
