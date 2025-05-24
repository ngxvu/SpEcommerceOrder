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

type ProductHandler struct {
	db             pgGorm.PGInterface
	productService services.ProductServiceInterface
}

func NewProductHandler(
	pgRepo pgGorm.PGInterface,
	productService services.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{
		db:             pgRepo,
		productService: productService,
	}
}

func (p *ProductHandler) CreateProduct(ctx *gin.Context) {

	log := logger.WithTag("Backend|ProductHandler|CreateProduct")

	var product model.CreateProductRequest
	if err := ctx.ShouldBindJSON(&product); err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to bind JSON")
		_ = ctx.Error(err)
		return
	}

	// Call the service to create the product
	createdProduct, err := p.productService.CreateProduct(ctx, product)
	if err != nil {
		err = app_errors.AppError("Failed to create product", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Failed to create product")
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, createdProduct)
}

func (p *ProductHandler) GetDetailProduct(ctx *gin.Context) {

	log := logger.WithTag("Backend|ProductHandler|GetDetailProduct")

	productID := ctx.Param("id")
	if productID == "" {
		err := app_errors.AppError("Product ID is required", app_errors.StatusValidationError)
		logger.LogError(log, err, "Product ID is required")
		_ = ctx.Error(err)
		return
	}

	product, err := p.productService.GetDetailProduct(ctx, productID)
	if err != nil {
		err = app_errors.AppError("Failed to get product details", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Failed to get product details")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, product)
}

func (p *ProductHandler) GetListProduct(ctx *gin.Context) {

	log := logger.WithTag("Backend|ProductHandler|GetListProduct")

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

	rs, err := p.productService.GetListProduct(ctx, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, paging.NewBodyPaginated(ctx, rs.Records, rs.Filter.Pager))
}

func (p *ProductHandler) UpdateProduct(ctx *gin.Context) {

	log := logger.WithTag("Backend|ProductHandler|UpdateProduct")

	productID := ctx.Param("id")
	if productID == "" {
		err := app_errors.AppError("Product ID is required", app_errors.StatusValidationError)
		logger.LogError(log, err, "Product ID is required")
		_ = ctx.Error(err)
		return
	}

	var updateProductRequest model.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&updateProductRequest); err != nil {
		err = app_errors.AppError("Invalid request", app_errors.StatusValidationError)
		logger.LogError(log, err, "Failed to bind JSON")
		_ = ctx.Error(err)
		return
	}

	// Call the service to update the product
	updatedProduct, err := p.productService.UpdateProduct(ctx, productID, updateProductRequest)
	if err != nil {
		err = app_errors.AppError("Failed to update product", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Failed to update product")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, updatedProduct)
}

func (p *ProductHandler) DeleteProduct(ctx *gin.Context) {

	log := logger.WithTag("Backend|ProductHandler|DeleteProduct")

	productID := ctx.Param("id")
	if productID == "" {
		err := app_errors.AppError("Product ID is required", app_errors.StatusValidationError)
		logger.LogError(log, err, "Product ID is required")
		_ = ctx.Error(err)
		return
	}

	response, err := p.productService.DeleteProduct(ctx, productID)
	if err != nil {
		err = app_errors.AppError("Failed to delete product", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Failed to delete product")
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
