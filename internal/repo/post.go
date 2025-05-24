package repo

import (
	"context"
	"gorm.io/gorm"
	model "kimistore/internal/models"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/paging"
	"time"
)

type PostRepository struct {
}

func NewPostRepository() *PostRepository {
	return &PostRepository{}
}

type PostRepositoryInterface interface {
	CreatePost(ctx context.Context, tx *gorm.DB, postRequest model.CreatePostRequest) (*model.GetPostResponse, error)
	GetDetailPost(ctx context.Context, tx *gorm.DB, id string) (*model.GetPostResponse, error)
	GetListPost(filter *paging.Filter, pgRepo *gorm.DB) (*model.ListPostResponse, error)
	UpdatePost(ctx context.Context, tx *gorm.DB, id string, postRequest model.UpdatePostRequest) (*model.GetPostResponse, error)
	DeletePost(ctx context.Context, tx *gorm.DB, id string) (*model.DeletePostResponse, error)
}

func (p *PostRepository) CreatePost(ctx context.Context, tx *gorm.DB, postRequest model.CreatePostRequest) (*model.GetPostResponse, error) {

	log := logger.WithTag("PostRepository|CreatePost")

	// Create a new post object
	post := model.Post{
		Title:       *postRequest.Title,
		Description: *postRequest.Description,
		Content:     *postRequest.Content,
		CoverURL:    *postRequest.CoverURL,
		Publish:     *postRequest.Publish,
	}

	// Save the post to the database
	if err := tx.Create(&post).Error; err != nil {
		logger.LogError(log, err, "Failed to create post")
		err := app_errors.AppError("Failed to create post", app_errors.StatusInternalServerError)
		return nil, err
	}

	// Map the post to the response format
	responseData := p.mapperPostsToResponse(post)

	response := &model.GetPostResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &responseData,
	}

	return response, nil

}

func (p *PostRepository) mapperPostsToResponse(post model.Post) model.GetPostResponseData {
	postID := utils.UUIDtoString(post.ID)

	responseData := &model.GetPostResponseData{}
	responseData.Post.ID = postID
	responseData.Post.Title = post.Title
	responseData.Post.Publish = post.Publish
	responseData.Post.Content = post.Content
	responseData.Post.CoverURL = post.CoverURL
	responseData.Post.Description = post.Description
	responseData.Post.CreatedAt = post.CreatedAt

	// Set default values for fields not in Post model
	responseData.Post.Comments = []struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		AvatarURL string    `json:"avatarUrl"`
		Message   string    `json:"message"`
		PostedAt  time.Time `json:"postedAt"`
		Users     []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatarUrl"`
		} `json:"users"`
		ReplyComment []struct {
			ID       string    `json:"id"`
			UserID   string    `json:"userId"`
			Message  string    `json:"message"`
			PostedAt time.Time `json:"postedAt"`
			TagUser  string    `json:"tagUser,omitempty"`
		} `json:"replyComment"`
	}{}
	responseData.Post.MetaKeywords = []string{}
	responseData.Post.Tags = []string{}
	responseData.Post.MetaTitle = ""
	responseData.Post.TotalViews = 0
	responseData.Post.TotalShares = 0
	responseData.Post.TotalComments = 0
	responseData.Post.TotalFavorites = 0
	responseData.Post.MetaDescription = ""
	responseData.Post.Author = struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
	}{}
	responseData.Post.FavoritePerson = []struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
	}{}

	return *responseData
}

func (p *PostRepository) GetDetailPost(ctx context.Context, tx *gorm.DB, id string) (*model.GetPostResponse, error) {
	log := logger.WithTag("PostRepository|GetDetailPost")

	var post model.Post
	if err := tx.Where("id = ?", id).First(&post).Error; err != nil {
		logger.LogError(log, err, "Failed to get post by ID")
		err := app_errors.AppError("Failed to get post", app_errors.StatusNotFound)
		return nil, err
	}

	responseData := p.mapperPostsToResponse(post)

	response := &model.GetPostResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &responseData,
	}

	return response, nil
}

func (r *PostRepository) GetListPost(filter *paging.Filter, pgRepo *gorm.DB) (*model.ListPostResponse, error) {

	log := logger.WithTag("PostRepository|GetListPost")

	tx := pgRepo.Model(&model.Post{})

	result := &model.ListPostResult{
		Filter:  filter,
		Records: []model.Post{},
	}

	pager := filter.Pager

	err := pager.DoQuery(&result.Records, tx).Error
	if err != nil {
		err := app_errors.AppError("Error getting list Post", app_errors.StatusNotFound)
		logger.LogError(log, err, "Error when getting list internal price group")
		return nil, err

	}

	mapper := model.GetPostResponseData{}

	var mapperList []model.OriginalPost

	for i := 0; i < len(result.Records); i++ {
		mapper = r.mapperPostsToResponse(result.Records[i])
		mapperList = append(mapperList, mapper.Post)
	}

	response := &model.ListPostResponse{
		Filter:  filter,
		Records: mapperList,
	}

	return response, nil
}

func (r *PostRepository) UpdatePost(ctx context.Context, tx *gorm.DB, id string, postRequest model.UpdatePostRequest) (*model.GetPostResponse, error) {
	log := logger.WithTag("PostRepository|UpdatePost")

	// Find the post by ID
	var post model.Post
	if err := tx.Where("id = ?", id).First(&post).Error; err != nil {
		logger.LogError(log, err, "Failed to find post by ID")
		err := app_errors.AppError("Failed to find post", app_errors.StatusNotFound)
		return nil, err
	}

	// Update the post fields
	post.Title = *postRequest.Title
	post.Description = *postRequest.Description
	post.Content = *postRequest.Content
	post.CoverURL = *postRequest.CoverURL
	post.Publish = *postRequest.Publish

	// Save the updated post to the database
	if err := tx.Save(&post).Error; err != nil {
		logger.LogError(log, err, "Failed to update post")
		err := app_errors.AppError("Failed to update post", app_errors.StatusInternalServerError)
		return nil, err
	}

	responseData := r.mapperPostsToResponse(post)

	response := &model.GetPostResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &responseData,
	}

	return response, nil
}

func (r *PostRepository) DeletePost(ctx context.Context, tx *gorm.DB, id string) (*model.DeletePostResponse, error) {

	log := logger.WithTag("PostRepository|DeletePost")

	// Find the post by ID
	var post model.Post
	if err := tx.Where("id = ?", id).First(&post).Error; err != nil {
		logger.LogError(log, err, "Failed to find post by ID")
		err = app_errors.AppError("Failed to find post", app_errors.StatusNotFound)
		return nil, err
	}

	// Delete the post from the database
	if err := tx.Delete(&post).Error; err != nil {
		logger.LogError(log, err, "Failed to delete post")
		err = app_errors.AppError("Failed to delete post", app_errors.StatusInternalServerError)
		return nil, err
	}

	return &model.DeletePostResponse{
		Meta:    utils.NewMetaData(ctx),
		Message: "Post deleted successfully",
	}, nil
}
