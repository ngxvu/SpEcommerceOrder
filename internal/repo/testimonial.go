package repo

import (
	model "basesource/internal/models"
	pgGorm "basesource/internal/repo/pg-gorm"
	"basesource/internal/utils"
	"basesource/pkg/http/paging"
	"context"
	"gorm.io/gorm"
)

type TestimonialRepository struct {
	db pgGorm.PGInterface
}

func NewTestimonialRepository(newPgRepo pgGorm.PGInterface) *TestimonialRepository {
	return &TestimonialRepository{db: newPgRepo}
}

type TestimonialRepositoryInterface interface {
	CheckTestimonialExists(ctx context.Context, tx *gorm.DB, id string) (bool, error)
	CreateTestimonial(ctx context.Context, tx *gorm.DB, testimonial model.Testimonial) (*model.GetTestimonialResponse, error)
	GetDetailTestimonial(ctx context.Context, tx *gorm.DB, id string) (*model.GetTestimonialResponse, error)
	GetListTestimonial(ctx context.Context, filter *paging.Filter, tx *gorm.DB) (*model.ListTestimonialResponse, error)
	UpdateTestimonial(ctx context.Context, tx *gorm.DB, id string, testimonialUpdate model.Testimonial) (*model.GetTestimonialResponse, error)
	DeleteTestimonial(ctx context.Context, tx *gorm.DB, testimonial *model.Testimonial) (*model.DeleteTestimonialResponse, error)
}

func (r *TestimonialRepository) CheckTestimonialExists(ctx context.Context, tx *gorm.DB, id string) (bool, error) {
	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = r.db.DBWithTimeout(ctx)
		defer cancel()
	}

	var count int64
	if err := tx.WithContext(ctx).Model(&model.Testimonial{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *TestimonialRepository) CreateTestimonial(ctx context.Context,
	tx *gorm.DB, testimonial model.Testimonial) (*model.GetTestimonialResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = r.db.DBWithTimeout(ctx)
		defer cancel()
	}

	if err := tx.WithContext(ctx).Create(&testimonial).Error; err != nil {
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

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = r.db.DBWithTimeout(ctx)
		defer cancel()
	}

	var testimonial model.Testimonial
	if err := tx.Where("id = ?", id).First(&testimonial).Error; err != nil {
		return nil, err
	}

	response := &model.GetTestimonialResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &testimonial,
	}

	return response, nil
}

func (r *TestimonialRepository) GetListTestimonial(ctx context.Context, filter *paging.Filter, tx *gorm.DB) (*model.ListTestimonialResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = r.db.DBWithTimeout(ctx)
		defer cancel()
	}

	query := tx.Model(&model.Testimonial{})

	result := &model.ListTestimonialResponse{
		Filter:  filter,
		Records: []model.Testimonial{},
	}

	filter.Pager.SortableFields = []string{"author"}

	pager := filter.Pager

	err := pager.DoQuery(&result.Records, query).Error
	if err != nil {
		return nil, err
	}

	response := &model.ListTestimonialResponse{
		Filter:  filter,
		Records: result.Records,
	}

	return response, nil
}

func (r *TestimonialRepository) UpdateTestimonial(ctx context.Context, tx *gorm.DB, id string, testimonialUpdate model.Testimonial) (*model.GetTestimonialResponse, error) {
	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = r.db.DBWithTimeout(ctx)
		defer cancel()
	}

	if err := tx.Model(&model.Testimonial{}).WithContext(ctx).Where("id = ?", id).Updates(testimonialUpdate).Error; err != nil {
		return nil, err
	}

	detailTestimonial, err := r.GetDetailTestimonial(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return detailTestimonial, nil
}

func (r *TestimonialRepository) DeleteTestimonial(ctx context.Context, tx *gorm.DB, testimonial *model.Testimonial) (*model.DeleteTestimonialResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = r.db.DBWithTimeout(ctx)
		defer cancel()
	}

	if err := tx.WithContext(ctx).Delete(testimonial).Error; err != nil {
		return nil, err
	}

	response := &model.DeleteTestimonialResponse{
		Meta: utils.NewMetaData(ctx),
		Data: "Testimonial deleted successfully",
	}

	return response, nil
}
