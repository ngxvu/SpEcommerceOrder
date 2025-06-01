package services

import (
	"context"
	model "kimistore/internal/models"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/utils/app_errors"
	"kimistore/internal/utils/sync_ob"
	"kimistore/pkg/http/logger"
)

type PostService struct {
	repo      repo.PostRepositoryInterface
	newPgRepo pgGorm.PGInterface
}

type PostServiceInterface interface {
	CreatePost(ctx context.Context, postRequest model.CreatePostRequest) (*model.GetPostResponse, error)
	GetDetailPost(ctx context.Context, id string) (*model.GetPostResponse, error)
	GetListPost(ctx context.Context, filter *model.ListPostFilter) (*model.ListPostResponse, error)
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

	postExistsByName, err := s.repo.PostExistsByName(ctx, tx, *postRequest.Title)
	if err != nil {
		logger.LogError(log, err, "Error checking if post exists")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	if postExistsByName {
		err = app_errors.AppError("Post already exists", app_errors.StatusConflict)
		return nil, err
	}

	ob := model.Post{}
	sync_ob.Sync(postRequest, &ob)

	// Call the repository to create the post
	createdPost, err := s.repo.CreatePost(ctx, tx, ob)
	if err != nil {
		logger.LogError(log, err, "Failed to create post")
		err = app_errors.AppError("Failed to create post", app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return createdPost, nil
}

func (s *PostService) GetDetailPost(ctx context.Context, id string) (*model.GetPostResponse, error) {
	log := logger.WithTag("PostService|GetDetailPost")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Check if the post exists
	exists, err := s.repo.CheckPostExistById(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to check if post exists")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !exists {
		logger.LogError(log, nil, "Post not found")
		err = app_errors.AppError("Post not found", app_errors.StatusNotFound)
		return nil, err
	}

	// Call the repository to get the post details
	post, err := s.repo.GetDetailPost(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to get post details")
		err = app_errors.AppError("Failed to get post details", app_errors.StatusInternalServerError)
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetListPost(ctx context.Context, filter *model.ListPostFilter) (*model.ListPostResponse, error) {
	log := logger.WithTag("PostService|GetListPost")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	result, err := s.repo.GetListPost(ctx, filter, tx)
	if err != nil {
		logger.LogError(log, err, "Error getting list of Posts")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	return result, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id string, postRequest model.UpdatePostRequest) (*model.GetPostResponse, error) {
	log := logger.WithTag("PostService|UpdatePost")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	// Check if the post exists
	exists, err := s.repo.CheckPostExistById(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to check if post exists")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !exists {
		err = app_errors.AppError("Post not found", app_errors.StatusNotFound)
		return nil, err
	}

	if postRequest.Title != nil {
		postExistsByName, err := s.repo.PostExistsByName(ctx, tx, *postRequest.Title)
		if err != nil {
			logger.LogError(log, err, "Error checking if post exists by name")
			err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
			return nil, err
		}
		if postExistsByName {
			logger.LogError(log, nil, "Post already exists with the same title")
			err = app_errors.AppError("Post already exists with the same title", app_errors.StatusConflict)
			return nil, err
		}
	}

	currentPost, err := s.repo.GetDetailPost(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to get current post details")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	orig := currentPost.Data.Post
	// Select the original values for nil fields in the request
	postToUpdate := s.selectOriginForNilValue(postRequest, orig)

	// Call the repository to update the post
	updatedPost, err := s.repo.UpdatePost(ctx, tx, postToUpdate)
	if err != nil {
		logger.LogError(log, err, "Failed to update post")
		err = app_errors.AppError("Failed to update post", app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return updatedPost, nil
}

func (s *PostService) DeletePost(ctx context.Context, id string) (*model.DeletePostResponse, error) {
	log := logger.WithTag("PostService|DeletePost")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	// Check if the post exists
	exists, err := s.repo.CheckPostExistById(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to check if post exists")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !exists {
		logger.LogError(log, nil, "Post not found")
		err = app_errors.AppError("Post not found", app_errors.StatusNotFound)
		return nil, err
	}

	getPost, err := s.repo.GetPostV1(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Failed to get post details")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	// Call the repository to delete the post
	rs, err := s.repo.DeletePost(ctx, tx, *getPost)
	if err != nil {
		logger.LogError(log, err, "Failed to delete post")
		err = app_errors.AppError("Failed to delete post", app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return rs, nil
}

// selectOriginForNilValue selects the original values from the original post
func (s *PostService) selectOriginForNilValue(
	postRequest model.UpdatePostRequest,
	orig model.OriginalPost,
) model.Post {
	return model.Post{
		Description: func() string {
			if postRequest.Description != nil {
				return *postRequest.Description
			}
			return orig.Description
		}(),
		Content: func() string {
			if postRequest.Content != nil {
				return *postRequest.Content
			}
			return orig.Content
		}(),
		Publish: func() string {
			if postRequest.Publish != nil {
				return *postRequest.Publish
			}
			return orig.Publish
		}(),
		Title: func() string {
			if postRequest.Title != nil {
				return *postRequest.Title
			}
			return orig.Title
		}(),
		CoverURL: func() string {
			if postRequest.CoverURL != nil {
				return *postRequest.CoverURL
			}
			return orig.CoverURL
		}(),
	}
}
