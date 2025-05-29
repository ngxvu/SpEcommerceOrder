package handlers

import (
	"github.com/gin-gonic/gin"
	model "kimistore/internal/models"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/services"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/paging"
	"net/http"
)

type PostHandler struct {
	db          pgGorm.PGInterface
	postService services.PostServiceInterface
}

func NewPostHandler(
	pgRepo pgGorm.PGInterface,
	postService services.PostServiceInterface) *PostHandler {
	return &PostHandler{
		db:          pgRepo,
		postService: postService,
	}
}

func (p *PostHandler) CreatePost(ctx *gin.Context) {

	log := logger.WithTag("Backend|PostHandler|CreatePost")

	var post model.CreatePostRequest

	if err := ctx.ShouldBindJSON(&post); err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to bind JSON")
		_ = ctx.Error(err)
		return
	}

	// Call the service to create the product
	createdPost, err := p.postService.CreatePost(ctx, post)
	if err != nil {
		logger.LogError(log, err, "Failed to create post")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, createdPost)
}

func (p *PostHandler) GetDetailPost(ctx *gin.Context) {
	log := logger.WithTag("Backend|PostHandler|GetDetailPost")

	postID := ctx.Param("id")
	if postID == "" {
		err := app_errors.AppError("Post ID is required", app_errors.StatusValidationError)
		logger.LogError(log, err, "Post ID is required")
		_ = ctx.Error(err)
		return
	}

	// Call the service to get the post details
	post, err := p.postService.GetDetailPost(ctx, postID)
	if err != nil {
		logger.LogError(log, err, "Failed to get post details")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (p *PostHandler) GetListPost(ctx *gin.Context) {
	log := logger.WithTag("Backend|PostHandler|GetListPost")

	var req model.PostFilterRequest

	err := ctx.BindQuery(&req)
	if err != nil {
		err = app_errors.AppError("fail to get pagination", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to bind query")
		_ = ctx.Error(err)
		return
	}

	filter := &model.ListPostFilter{
		PostFilterRequest: req,
		Pager:             paging.NewPagerWithGinCtx(ctx),
	}

	rs, err := p.postService.GetListPost(ctx, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, paging.NewBodyPaginated(ctx, rs.Records, rs.Filter.Pager))
}

func (p *PostHandler) UpdatePost(ctx *gin.Context) {
	log := logger.WithTag("Backend|PostHandler|UpdatePost")

	postID := ctx.Param("id")
	if postID == "" {
		err := app_errors.AppError("Post ID is required", app_errors.StatusValidationError)
		logger.LogError(log, err, "Post ID is required")
		_ = ctx.Error(err)
		return
	}

	var postRequest model.UpdatePostRequest

	if err := ctx.ShouldBindJSON(&postRequest); err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to bind JSON")
		_ = ctx.Error(err)
		return
	}

	post, err := p.postService.UpdatePost(ctx, postID, postRequest)
	if err != nil {
		logger.LogError(log, err, "Failed to update post")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (p *PostHandler) DeletePost(ctx *gin.Context) {
	log := logger.WithTag("Backend|PostHandler|DeletePost")

	postID := ctx.Param("id")
	if postID == "" {
		err := app_errors.AppError("Post ID is required", app_errors.StatusValidationError)
		logger.LogError(log, err, "Post ID is required")
		_ = ctx.Error(err)
		return
	}

	response, err := p.postService.DeletePost(ctx, postID)
	if err != nil {
		logger.LogError(log, err, "Failed to delete post")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
