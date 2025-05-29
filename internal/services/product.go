package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	model "kimistore/internal/models"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
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

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Check if the productRequest already exists
	existingProduct, err := s.repo.ProductExistsByName(tx, *productRequest.Name)
	if err != nil {
		logger.LogError(log, err, "Error checking if productRequest exists")
		return nil, err
	}
	if existingProduct {
		err := app_errors.AppError("Product already exists", app_errors.StatusConflict)
		logger.LogError(log, err, "Product already exists")
		return nil, err
	}

	images, err := s.transferImagesToJsonB(productRequest.Images)
	if err != nil {
		return nil, err
	}

	sizes, err := s.transferSizesToJsonB(productRequest.Sizes)
	if err != nil {
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
		return nil, err
	}

	tx.Commit()

	return newProduct, nil

}

func (s *ProductService) transferImagesToJsonB(imagesURL []*string) (*postgres.Jsonb, error) {
	return toJsonb(imagesURL)
}

func (s *ProductService) transferSizesToJsonB(sizes []*string) (*postgres.Jsonb, error) {
	return toJsonb(sizes)
}

func toJsonb(data interface{}) (*postgres.Jsonb, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var jsonbData postgres.Jsonb
	if err := jsonbData.UnmarshalJSON(jsonData); err != nil {
		return nil, err
	}
	return &jsonbData, nil
}

func (s *ProductService) GetDetailProduct(ctx context.Context, id string) (*model.GetProductResponse, error) {
	log := logger.WithTag("ProductService|GetProductByID")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	product, err := s.repo.GetDetailProduct(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Error getting product by ID")
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetListProduct(ctx context.Context,
	filter *model.ListProductFilter) (*model.ListProductResponse, error) {
	log := logger.WithTag("ProductService|GetListProduct")

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	result, err := s.repo.GetListProduct(filter, tx)
	if err != nil {
		logger.LogError(log, err, "Error getting list of products")
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
	validOps := s.ValidOperators()
	validOpsMap := make(map[string]bool)
	for _, op := range validOps {
		validOpsMap[op] = true
	}

	// Validate each filter's operator
	if !validOpsMap[filter.Operator] {
		err := app_errors.AppError(fmt.Sprintf("Invalid operator '%s'", filter.Operator), app_errors.StatusValidationError)
		logger.LogError(log, err, "Invalid filter operator")
		return nil, err
	}

	// Create a pager for database pagination
	pager := &paging.Pager{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	tx = s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	result, err := s.repo.FilterColumnProduct(filter, pager, tx)
	if err != nil {
		logger.LogError(log, err, "Error filtering list of products")
		return nil, err
	}

	tx.Commit()
	return result, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, productRequest model.UpdateProductRequest) (*model.GetProductResponse, error) {
	log := logger.WithTag("ProductService|UpdateProduct")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	// Check if the productRequest already exists
	existingProduct, err := s.repo.ProductExistsByName(tx, *productRequest.Name)
	if err != nil {
		logger.LogError(log, err, "Error checking if productRequest exists")
		return nil, err
	}
	if existingProduct {
		err := app_errors.AppError("Product already exists", app_errors.StatusConflict)
		logger.LogError(log, err, "Product already exists")
		return nil, err
	}

	images, err := s.transferImagesToJsonB(productRequest.Images)
	if err != nil {
		return nil, err
	}

	sizes, err := s.transferSizesToJsonB(productRequest.Sizes)
	if err != nil {
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
		return nil, err
	}

	tx.Commit()

	return productResponse, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) (*model.DeleteProductResponse, error) {
	log := logger.WithTag("ProductService|DeleteProduct")

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	response, err := s.repo.DeleteProduct(ctx, tx, id)
	if err != nil {
		logger.LogError(log, err, "Error deleting product")
		return nil, err
	}

	tx.Commit()

	return response, nil
}

func (s *ProductService) ValidOperators() []string {
	return []string{
		"contains", "not_contains", "equals", "not_equals",
		"starts_with", "ends_with", "is_empty", "is_not_empty",
		"is_any_of", "greater_than", "less_than",
		"greater_than_or_equal", "less_than_or_equal",
	}
}
