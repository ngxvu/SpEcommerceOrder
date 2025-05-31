package handlers

import (
	"github.com/gin-gonic/gin"
	model "kimistore/internal/models"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/services"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
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

	context := ctx.Request.Context()

	var postRequest model.CreatePostRequest

	if err := ctx.ShouldBindJSON(&postRequest); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	allowedPublishValues := []string{utils.PublishDraft, utils.PublishPublished}
	if !utils.ContainsString(*postRequest.Publish, allowedPublishValues) {
		err := app_errors.AppError("Must be 'draft' or 'published'", app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	// Call the service to create the product
	createdPost, err := p.postService.CreatePost(context, postRequest)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, createdPost)
}

func (p *PostHandler) GetDetailPost(ctx *gin.Context) {

	context := ctx.Request.Context()

	postID := ctx.Param("id")

	// Call the service to get the post details
	post, err := p.postService.GetDetailPost(context, postID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (p *PostHandler) GetListPost(ctx *gin.Context) {

	context := ctx.Request.Context()

	var req model.PostFilterRequest

	err := ctx.BindQuery(&req)
	if err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	if req.Publish != nil {
		filterByPublish := *req.Publish

		allowedPublishValues := []string{utils.PublishDraft, utils.PublishPublished}
		if !utils.ContainsString(filterByPublish, allowedPublishValues) {
			err := app_errors.AppError("Must be 'draft' or 'published'", app_errors.StatusBadRequest)
			_ = ctx.Error(err)
			return
		}
	}

	filter := &model.ListPostFilter{
		PostFilterRequest: req,
		Pager:             paging.NewPagerWithGinCtx(ctx),
	}

	rs, err := p.postService.GetListPost(context, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, paging.NewBodyPaginated(ctx, rs.Records, rs.Filter.Pager))
}

func (p *PostHandler) UpdatePost(ctx *gin.Context) {

	context := ctx.Request.Context()

	postID := ctx.Param("id")

	var postRequest model.UpdatePostRequest

	if err := ctx.ShouldBindJSON(&postRequest); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	post, err := p.postService.UpdatePost(context, postID, postRequest)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (p *PostHandler) DeletePost(ctx *gin.Context) {

	context := ctx.Request.Context()

	postID := ctx.Param("id")

	response, err := p.postService.DeletePost(context, postID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
