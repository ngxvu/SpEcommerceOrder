package services

import (
	"context"
	model "kimistore/internal/models"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/paging"
)

type PostService struct {
	repo      repo.PostRepositoryInterface
	newPgRepo pgGorm.PGInterface
}

type PostServiceInterface interface {
	CreatePost(ctx context.Context, postRequest model.CreatePostRequest) (*model.GetPostResponse, error)
	GetDetailPost(ctx context.Context, id string) (*model.GetPostResponse, error)
	GetListPost(ctx context.Context, filter *paging.Filter) (*model.ListPostResponse, error)
	UpdatePost(ctx context.Context, id string, postRequest model.UpdatePostRequest) (*model.GetPostResponse, error)
	DeletePost(ctx context.Context, id string) (*model.DeletePostResponse, error)
}

func NewPostService(repo repo.PostRepositoryInterface, newRepo pgGorm.PGInterface) *PostService {
	return &PostService{
		repo:      repo,
		newPgRepo: newRepo,
	}
}

func (s *PostService) CreatePost(ctx context.Context, postRequest model.CreatePostRequest) (*model.GetPostResponse, error) {

	log := logger.WithTag("PostService|CreatePost")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Call the repository to create the post
	createdPost, err := s.repo.CreatePost(ctx, tx, postRequest)
	if err != nil {
		logger.LogError(log, err, "Failed to create post")
		return nil, err
	}

	return createdPost, nil
}

func (s *PostService) GetDetailPost(ctx context.Context, id string) (*model.GetPostResponse, error) {
	log := logger.WithTag("PostService|GetDetailPost")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Call the repository to get the post details
	post, err := s.repo.GetDetailPost(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to get post details")
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetListPost(ctx context.Context, filter *paging.Filter) (*model.ListPostResponse, error) {
	log := logger.WithTag("PostService|GetListPost")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	result, err := s.repo.GetListPost(filter, tx)
	if err != nil {
		logger.LogError(log, err, "Error getting list of Posts")
		return nil, err
	}
	return result, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id string, postRequest model.UpdatePostRequest) (*model.GetPostResponse, error) {
	log := logger.WithTag("PostService|UpdatePost")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Call the repository to update the post
	updatedPost, err := s.repo.UpdatePost(ctx, tx, id, postRequest)
	if err != nil {
		logger.LogError(log, err, "Failed to update post")
		return nil, err
	}

	return updatedPost, nil
}

func (s *PostService) DeletePost(ctx context.Context, id string) (*model.DeletePostResponse, error) {
	log := logger.WithTag("PostService|DeletePost")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Call the repository to delete the post
	rs, err := s.repo.DeletePost(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to delete post")
		return nil, err
	}

	return rs, nil
}
