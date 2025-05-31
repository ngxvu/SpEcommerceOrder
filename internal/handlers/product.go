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

	context := ctx.Request.Context()

	var productRequest model.CreateProductRequest
	if err := ctx.ShouldBindJSON(&productRequest); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	allowedPublishValues := []string{utils.PublishDraft, utils.PublishPublished}
	if !utils.ContainsString(*productRequest.Publish, allowedPublishValues) {
		err := app_errors.AppError("Must be 'draft' or 'published'", app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	allowedInventoryTypes := []string{utils.InventoryInStock, utils.InventoryOutOfStock, utils.InventoryLowStock}
	if !utils.ContainsString(*productRequest.InventoryType, allowedInventoryTypes) {
		err := app_errors.AppError("Must be 'in stock', 'out of stock', or 'low stock'", app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	// Call the service to create the productRequest
	createdProduct, err := p.productService.CreateProduct(context, productRequest)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, createdProduct)
}

func (p *ProductHandler) GetDetailProduct(ctx *gin.Context) {

	context := ctx.Request.Context()

	productID := ctx.Param("id")

	product, err := p.productService.GetDetailProduct(context, productID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, product)
}

func (p *ProductHandler) GetListProduct(ctx *gin.Context) {

	context := ctx.Request.Context()

	var req model.ProductFilterRequest
	if err := ctx.BindQuery(&req); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	filter := &model.ListProductFilter{
		ProductFilterRequest: req,
		Pager:                paging.NewPagerWithGinCtx(ctx),
	}

	if filter.FilterByStock != nil {
		filterByStock := *filter.FilterByStock

		allowedInventoryTypes := []string{utils.InventoryInStock, utils.InventoryOutOfStock, utils.InventoryLowStock}
		if !utils.ContainsString(filterByStock, allowedInventoryTypes) {
			err := app_errors.AppError("Must be 'in stock', 'out of stock', or 'low stock'", app_errors.StatusBadRequest)
			_ = ctx.Error(err)
			return
		}
	}

	if filter.FilterByPublish != nil {
		filterByPublish := *filter.FilterByPublish

		allowedPublishValues := []string{utils.PublishDraft, utils.PublishPublished}
		if !utils.ContainsString(filterByPublish, allowedPublishValues) {
			err := app_errors.AppError("Must be 'draft' or 'published'", app_errors.StatusBadRequest)
			_ = ctx.Error(err)
			return
		}
	}

	rs, err := p.productService.GetListProduct(context, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, paging.NewBodyPaginated(ctx, rs.Records, rs.Filter.Pager))
}

func (p *ProductHandler) ListProductFilterAdvance(ctx *gin.Context) {

	context := ctx.Request.Context()

	var req model.ColumnFilterParam

	if err := ctx.ShouldBindJSON(&req); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	rs, err := p.productService.ListProductFilterAdvance(context, &req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// Create pager for response
	pager := &paging.Pager{
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalRows: int64(len(rs.Records)),
	}

	ctx.JSON(http.StatusOK, paging.NewBodyPaginated(ctx, rs.Records, pager))
}

func (p *ProductHandler) UpdateProduct(ctx *gin.Context) {

	context := ctx.Request.Context()

	productID := ctx.Param("id")

	var updateProductRequest model.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&updateProductRequest); err != nil {
		err = app_errors.AppError(app_errors.StatusBadRequest, app_errors.StatusBadRequest)
		_ = ctx.Error(err)
		return
	}

	// Call the service to update the product
	updatedProduct, err := p.productService.UpdateProduct(context, productID, updateProductRequest)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, updatedProduct)
}

func (p *ProductHandler) DeleteProduct(ctx *gin.Context) {

	context := ctx.Request.Context()

	productID := ctx.Param("id")

	response, err := p.productService.DeleteProduct(context, productID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
