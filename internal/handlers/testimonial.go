package handlers

import (
	"github.com/gin-gonic/gin"
	model "kimistore/internal/models"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/services"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/paging"
	"net/http"
)

type TestimonialHandler struct {
	db                 pgGorm.PGInterface
	TestimonialService services.TestimonialServiceInterface
}

func NewTestimonialHandler(
	pgRepo pgGorm.PGInterface,
	testimonialService services.TestimonialServiceInterface) *TestimonialHandler {
	return &TestimonialHandler{
		db:                 pgRepo,
		TestimonialService: testimonialService,
	}
}

// CreateTestimonial creates a new testimonial
func (h *TestimonialHandler) CreateTestimonial(ctx *gin.Context) {

	var req model.TestimonialRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	context := ctx.Request.Context()

	testimonial, err := h.TestimonialService.CreateTestimonial(context, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, testimonial)
}

func (h *TestimonialHandler) GetDetailTestimonial(ctx *gin.Context) {

	id := ctx.Param("id")

	context := ctx.Request.Context()

	testimonial, err := h.TestimonialService.GetDetailTestimonial(context, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, testimonial)
}

func (h *TestimonialHandler) GetListTestimonial(ctx *gin.Context) {

	context := ctx.Request.Context()

	var req paging.Param

	err := ctx.BindQuery(&req)
	if err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	filter := &paging.Filter{
		Param: req,
		Pager: paging.NewPagerWithGinCtx(ctx),
	}

	rs, err := h.TestimonialService.GetListTestimonial(context, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, paging.NewBodyPaginated(ctx, rs.Records, rs.Filter.Pager))
}

func (h *TestimonialHandler) UpdateTestimonial(ctx *gin.Context) {

	context := ctx.Request.Context()

	id := ctx.Param("id")

	var req model.TestimonialRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	testimonial, err := h.TestimonialService.UpdateTestimonial(context, id, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, testimonial)
}

func (h *TestimonialHandler) DeleteTestimonial(ctx *gin.Context) {

	context := ctx.Request.Context()

	id := ctx.Param("id")

	response, err := h.TestimonialService.DeleteTestimonial(context, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
