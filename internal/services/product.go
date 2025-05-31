package services

import (
	"context"
	"fmt"
	model "kimistore/internal/models"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/paging"
)

type ProductService struct {
	repo      repo.ProductRepositoryInterface
	newPgRepo pgGorm.PGInterface
}

type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, productRequest model.CreateProductRequest) (*model.GetProductResponse, error)
	GetDetailProduct(ctx context.Context, id string) (*model.GetProductResponse, error)
	GetListProduct(ctx context.Context, filter *model.ListProductFilter) (*model.ListProductResponse, error)
	ListProductFilterAdvance(ctx context.Context, req *model.ColumnFilterParam) (*model.ListProductResponse, error)
	UpdateProduct(ctx context.Context, id string, productRequest model.UpdateProductRequest) (*model.GetProductResponse, error)
	DeleteProduct(ctx context.Context, id string) (*model.DeleteProductResponse, error)
}

func NewProductService(repo repo.ProductRepositoryInterface, newRepo pgGorm.PGInterface) *ProductService {
	return &ProductService{
		repo:      repo,
		newPgRepo: newRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, productRequest model.CreateProductRequest) (*model.GetProductResponse, error) {

	log := logger.WithTag("ProductService|CreateProduct")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	// Check if the productRequest already exists
	existingProduct, err := s.repo.ProductExistsByName(tx, *productRequest.Name)
	if err != nil {
		logger.LogError(log, err, "Error checking if productRequest exists")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if existingProduct {
		logger.LogError(log, fmt.Errorf("product already exists"), "Product already exists")
		err = app_errors.AppError("Product already exists", app_errors.StatusConflict)
		return nil, err
	}

	images, err := utils.TransferDataToJsonB(productRequest.Images)
	if err != nil {
		logger.LogError(log, err, "Error transferring images to JSONB")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	sizes, err := utils.TransferDataToJsonB(productRequest.Sizes)
	if err != nil {
		logger.LogError(log, err, "Error transferring sizes to JSONB")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	product := model.Product{
		CoverURL:       *productRequest.CoverURL,
		Images:         images,
		Publish:        *productRequest.Publish,
		Name:           *productRequest.Name,
		Price:          *productRequest.Price,
		Quantity:       *productRequest.Quantity,
		InventoryType:  *productRequest.InventoryType,
		Sizes:          sizes,
		SubDescription: *productRequest.SubDescription,
		Description:    *productRequest.Description,
	}

	// Create the productRequest in the database
	newProduct, err := s.repo.CreateProduct(ctx, tx, product)
	if err != nil {
		logger.LogError(log, err, "Error creating productRequest")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return newProduct, nil

}

func (s *ProductService) GetDetailProduct(ctx context.Context, id string) (*model.GetProductResponse, error) {

	log := logger.WithTag("ProductService|GetDetailProduct")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Check if the product exists
	existingProduct, err := s.repo.ProductExistsByID(tx, id)
	if err != nil {
		logger.LogError(log, err, "Error checking if product exists")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !existingProduct {
		logger.LogError(log, fmt.Errorf("product with id %s does not exist", id), "Product not found")
		err = app_errors.AppError("Product not found", app_errors.StatusNotFound)
		return nil, err
	}

	product, err := s.repo.GetDetailProduct(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Error getting product details")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetListProduct(ctx context.Context,
	filter *model.ListProductFilter) (*model.ListProductResponse, error) {

	log := logger.WithTag("ProductService|GetListProduct")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	result, err := s.repo.GetListProduct(filter, tx)
	if err != nil {
		logger.LogError(log, err, "Error getting list of products")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	return result, nil
}

func (s *ProductService) ListProductFilterAdvance(ctx context.Context, filter *model.ColumnFilterParam) (*model.ListProductResponse, error) {
	log := logger.WithTag("ProductService|ListProductFilterAdvance")

	// Validate pagination parameters
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	// Validate operators in filters
	validOps := utils.ValidOperators()
	validOpsMap := make(map[string]bool)
	for _, op := range validOps {
		validOpsMap[op] = true
	}

	// Validate each filter's operator
	if !validOpsMap[filter.Operator] {
		logger.LogError(log, fmt.Errorf("invalid operator '%s'", filter.Operator), "Invalid filter operator")
		err := app_errors.AppError(app_errors.StatusValidationError, app_errors.StatusValidationError)
		return nil, err
	}

	// Create a pager for database pagination
	pager := &paging.Pager{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	result, err := s.repo.FilterColumnProduct(filter, pager, tx)
	if err != nil {
		logger.LogError(log, err, "Error filtering list of products")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	return result, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, productRequest model.UpdateProductRequest) (*model.GetProductResponse, error) {
	log := logger.WithTag("ProductService|UpdateProduct")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	// Check if the productRequest already exists
	existingProduct, err := s.repo.ProductExistsByName(tx, *productRequest.Name)
	if err != nil {
		logger.LogError(log, err, "Error checking if productRequest exists")
		err := app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if existingProduct {
		logger.LogError(log, err, "Product already exists")
		err := app_errors.AppError("Product already exists", app_errors.StatusConflict)
		return nil, err
	}

	images, err := utils.TransferDataToJsonB(productRequest.Images)
	if err != nil {
		logger.LogError(log, err, "Error transferring images to JSONB")
		err := app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	sizes, err := utils.TransferDataToJsonB(productRequest.Sizes)
	if err != nil {
		logger.LogError(log, err, "Error transferring sizes to JSONB")
		err := app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	product := model.Product{
		CoverURL:       *productRequest.CoverURL,
		Images:         images,
		Publish:        *productRequest.Publish,
		Name:           *productRequest.Name,
		Price:          *productRequest.Price,
		Quantity:       *productRequest.Quantity,
		InventoryType:  *productRequest.InventoryType,
		Sizes:          sizes,
		SubDescription: *productRequest.SubDescription,
		Description:    *productRequest.Description,
	}

	productResponse, err := s.repo.UpdateProduct(ctx, tx, id, product)
	if err != nil {
		logger.LogError(log, err, "Error updating product")
		err = app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return productResponse, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) (*model.DeleteProductResponse, error) {
	log := logger.WithTag("ProductService|DeleteProduct")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	// Check if the product exists
	existingProduct, err := s.repo.ProductExistsByID(tx, id)
	if err != nil {
		logger.LogError(log, err, "Error checking if product exists")
		err := app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}
	if !existingProduct {
		logger.LogError(log, fmt.Errorf("product with id %s does not exist", id), "Product not found")
		err := app_errors.AppError("Product not found", app_errors.StatusNotFound)
		return nil, err
	}

	response, err := s.repo.DeleteProduct(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Error deleting product")
		err := app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	tx.Commit()

	return response, nil
}
