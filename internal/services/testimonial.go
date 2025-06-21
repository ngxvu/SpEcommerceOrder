package services

import (
	model "basesource/internal/models"
	"basesource/internal/repo"
	pgGorm "basesource/internal/repo/pg-gorm"
	"basesource/internal/utils/app_errors"
	"basesource/internal/utils/sync_ob"
	"basesource/pkg/http/logger"
	"basesource/pkg/http/paging"
	"context"
)

type TestimonialService struct {
	repo      repo.TestimonialRepositoryInterface
	newPgRepo pgGorm.PGInterface
}

type TestimonialServiceInterface interface {
	CreateTestimonial(ctx context.Context, req model.TestimonialRequest) (*model.GetTestimonialResponse, error)
	GetDetailTestimonial(ctx context.Context, id string) (*model.GetTestimonialResponse, error)
	GetListTestimonial(ctx context.Context, filter *paging.Filter) (*model.ListTestimonialResponse, error)
	UpdateTestimonial(ctx context.Context, id string, req model.TestimonialRequest) (*model.GetTestimonialResponse, error)
	DeleteTestimonial(ctx context.Context, id string) (*model.DeleteTestimonialResponse, error)
}

func NewTestimonialService(repo repo.TestimonialRepositoryInterface, newRepo pgGorm.PGInterface) *TestimonialService {
	return &TestimonialService{
		repo:      repo,
		newPgRepo: newRepo,
	}
}

func (s *TestimonialService) CreateTestimonial(ctx context.Context,
	req model.TestimonialRequest) (*model.GetTestimonialResponse, error) {

	log := logger.WithTag("TestimonialService|CreateTestimonial")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	ob := model.Testimonial{}

	sync_ob.Sync(req, &ob)

	testimonial, err := s.repo.CreateTestimonial(ctx, tx, ob)
	if err != nil {
		logger.LogError(log, err, "failed to create testimonial")
		err = app_errors.AppError("Fail to create testimonial", app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return testimonial, nil

}

func (s *TestimonialService) GetDetailTestimonial(ctx context.Context, id string) (*model.GetTestimonialResponse, error) {
	log := logger.WithTag("TestimonialService|GetDetailTestimonial")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	exists, err := s.repo.CheckTestimonialExists(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "failed to check testimonial existence")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !exists {
		logger.LogError(log, nil, "testimonial not found")
		err = app_errors.AppError("Testimonial not found", app_errors.StatusNotFound)
		return nil, err
	}

	testimonial, err := s.repo.GetDetailTestimonial(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "failed to get testimonial detail")
		err = app_errors.AppError("Fail to get testimonial detail", app_errors.StatusInternalServerError)
		return nil, err
	}

	return testimonial, nil
}

func (s *TestimonialService) GetListTestimonial(ctx context.Context, filter *paging.Filter) (*model.ListTestimonialResponse, error) {

	log := logger.WithTag("TestimonialService|GetListTestimonial")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	testimonials, err := s.repo.GetListTestimonial(ctx, filter, tx)
	if err != nil {
		logger.LogError(log, err, "failed to get list testimonial")
		err = app_errors.AppError("Fail to get list testimonial", app_errors.StatusInternalServerError)
		return nil, err
	}

	return testimonials, nil
}

func (s *TestimonialService) UpdateTestimonial(ctx context.Context, id string, req model.TestimonialRequest) (*model.GetTestimonialResponse, error) {
	log := logger.WithTag("TestimonialService|UpdateTestimonial")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	exists, err := s.repo.CheckTestimonialExists(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "failed to check testimonial existence")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !exists {
		logger.LogError(log, nil, "testimonial not found")
		err = app_errors.AppError("Testimonial not found", app_errors.StatusNotFound)
		return nil, err
	}

	ob := model.Testimonial{}

	sync_ob.Sync(req, &ob)

	testimonial, err := s.repo.UpdateTestimonial(ctx, tx, id, ob)
	if err != nil {
		logger.LogError(log, err, "failed to update testimonial")
		err = app_errors.AppError("Fail to update testimonial", app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return testimonial, nil
}

func (s *TestimonialService) DeleteTestimonial(ctx context.Context, id string) (*model.DeleteTestimonialResponse, error) {
	log := logger.WithTag("TestimonialService|DeleteTestimonial")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	exists, err := s.repo.CheckTestimonialExists(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "failed to check testimonial existence")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !exists {
		logger.LogError(log, nil, "testimonial not found")
		err = app_errors.AppError("Testimonial not found", app_errors.StatusNotFound)
		return nil, err
	}

	getTestimonial, err := s.repo.GetDetailTestimonial(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "failed to get testimonial detail")
		err = app_errors.AppError("Fail to get testimonial", app_errors.StatusInternalServerError)
		return nil, err
	}

	response, err := s.repo.DeleteTestimonial(ctx, tx, getTestimonial.Data)
	if err != nil {
		logger.LogError(log, err, "failed to delete testimonial")
		err := app_errors.AppError("Fail to delete testimonial", app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return response, nil
}
