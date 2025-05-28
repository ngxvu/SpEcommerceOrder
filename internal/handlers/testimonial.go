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

	log := logger.WithTag("TestimonialHandler|CreateTestimonial")

	var req model.TestimonialRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, " fail to bind json")
		_ = ctx.Error(err)
		return
	}

	testimonial, err := h.TestimonialService.CreateTestimonial(ctx, req)
	if err != nil {
		logger.LogError(log, err, "failed to create testimonial")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, testimonial)
}

func (h *TestimonialHandler) GetDetailTestimonial(ctx *gin.Context) {
	log := logger.WithTag("TestimonialHandler|GetDetailTestimonial")

	id := ctx.Param("id")
	if id == "" {
		err := app_errors.AppError("Testimonial ID is required", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Testimonial ID is required")
		_ = ctx.Error(err)
		return
	}

	testimonial, err := h.TestimonialService.GetDetailTestimonial(ctx, id)
	if err != nil {
		logger.LogError(log, err, "failed to get testimonial detail")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, testimonial)
}

func (h *TestimonialHandler) GetListTestimonial(ctx *gin.Context) {
	log := logger.WithTag("TestimonialHandler|GetListTestimonial")

	var req paging.Param

	err := ctx.BindQuery(&req)
	if err != nil {
		err = app_errors.AppError("fail to get pagination", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to bind query")
		_ = ctx.Error(err)
		return
	}

	filter := &paging.Filter{
		Param: req,
		Pager: paging.NewPagerWithGinCtx(ctx),
	}

	testimonials, err := h.TestimonialService.GetListTestimonial(ctx, filter)
	if err != nil {
		logger.LogError(log, err, "failed to get list of testimonials")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, testimonials)
}

func (h *TestimonialHandler) UpdateTestimonial(ctx *gin.Context) {
	log := logger.WithTag("TestimonialHandler|UpdateTestimonial")

	id := ctx.Param("id")
	if id == "" {
		err := app_errors.AppError("Testimonial ID is required", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Testimonial ID is required")
		_ = ctx.Error(err)
		return
	}

	var req model.TestimonialRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, "fail to bind json")
		_ = ctx.Error(err)
		return
	}

	testimonial, err := h.TestimonialService.UpdateTestimonial(ctx, id, req)
	if err != nil {
		logger.LogError(log, err, "failed to update testimonial")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, testimonial)
}

func (h *TestimonialHandler) DeleteTestimonial(ctx *gin.Context) {
	log := logger.WithTag("TestimonialHandler|DeleteTestimonial")

	id := ctx.Param("id")
	if id == "" {
		err := app_errors.AppError("Testimonial ID is required", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Testimonial ID is required")
		_ = ctx.Error(err)
		return
	}

	response, err := h.TestimonialService.DeleteTestimonial(ctx, id)
	if err != nil {
		logger.LogError(log, err, "failed to delete testimonial")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
