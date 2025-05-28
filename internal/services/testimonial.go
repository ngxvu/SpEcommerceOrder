package services

import (
	"context"
	model "kimistore/internal/models"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/utils/sync_ob"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/paging"
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

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	ob := model.Testimonial{}

	sync_ob.Sync(req, &ob)

	testimonial, err := s.repo.CreateTestimonial(ctx, tx, ob)
	if err != nil {
		logger.LogError(log, err, "failed to create testimonial")
		return nil, err
	}

	tx.Commit()

	return testimonial, nil

}

func (s *TestimonialService) GetDetailTestimonial(ctx context.Context, id string) (*model.GetTestimonialResponse, error) {
	log := logger.WithTag("TestimonialService|GetDetailTestimonial")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	testimonial, err := s.repo.GetDetailTestimonial(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "failed to get testimonial detail")
		return nil, err
	}

	tx.Commit()

	return testimonial, nil
}

func (s *TestimonialService) GetListTestimonial(ctx context.Context, filter *paging.Filter) (*model.ListTestimonialResponse, error) {

	log := logger.WithTag("TestimonialService|GetListTestimonial")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	testimonials, err := s.repo.GetListTestimonial(filter, tx)
	if err != nil {
		logger.LogError(log, err, "failed to get list testimonial")
		return nil, err
	}

	return testimonials, nil
}

func (s *TestimonialService) UpdateTestimonial(ctx context.Context, id string, req model.TestimonialRequest) (*model.GetTestimonialResponse, error) {
	log := logger.WithTag("TestimonialService|UpdateTestimonial")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	ob := model.Testimonial{}

	sync_ob.Sync(req, &ob)

	testimonial, err := s.repo.UpdateTestimonial(ctx, tx, id, ob)
	if err != nil {
		logger.LogError(log, err, "failed to update testimonial")
		return nil, err
	}

	tx.Commit()

	return testimonial, nil
}

func (s *TestimonialService) DeleteTestimonial(ctx context.Context, id string) (*model.DeleteTestimonialResponse, error) {
	log := logger.WithTag("TestimonialService|DeleteTestimonial")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	response, err := s.repo.DeleteTestimonial(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "failed to delete testimonial")
		return nil, err
	}

	tx.Commit()

	return response, nil
}
