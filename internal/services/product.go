package services

import (
	"context"
	"encoding/json"
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

	description := productRequest.Description
	if description == nil {
		defaultDescription := ""
		description = &defaultDescription
	}

	subDescription := productRequest.SubDescription
	if subDescription == nil {
		defaultSubDescription := ""
		subDescription = &defaultSubDescription
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
		SubDescription: *subDescription,
		Description:    *description,
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

	exist, err := s.repo.ProductExistsByID(tx, id)
	if err != nil {
		logger.LogError(log, err, "Error checking if product exists by ID")
		err := app_errors.AppError(app_errors.StatusInternalServerError, app_errors.StatusInternalServerError)
		return nil, err
	}

	if !exist {
		logger.LogError(log, nil, "Product not found")
		err = app_errors.AppError("Product not found", app_errors.StatusNotFound)
		return nil, err
	}

	// Check if the productRequest already exists
	if productRequest.Name != nil {
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
	}

	currentProduct, err := s.repo.GetDetailProduct(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	orig := currentProduct.Data.Product

	images, err := s.processStringForUpdate(productRequest.Images, orig.Images)
	if err != nil {
		return nil, err
	}

	sizes, err := s.processStringForUpdate(productRequest.Sizes, orig.Sizes)
	if err != nil {
		return nil, err
	}

	product := s.selectOriginForNilValue(productRequest, orig, images, sizes)

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

func (s *ProductService) selectOriginForNilValue(
	productRequest model.UpdateProductRequest,
	orig model.OriginalProduct,
	images *postgres.Jsonb,
	sizes *postgres.Jsonb,
) model.Product {
	return model.Product{
		CoverURL: func() string {
			if productRequest.CoverURL != nil {
				return *productRequest.CoverURL
			}
			return orig.CoverURL
		}(),
		Images: images,
		Publish: func() string {
			if productRequest.Publish != nil {
				return *productRequest.Publish
			}
			return orig.Publish
		}(),
		Name: func() string {
			if productRequest.Name != nil {
				return *productRequest.Name
			}
			return orig.Name
		}(),
		Price: func() float64 {
			if productRequest.Price != nil {
				return *productRequest.Price
			}
			return orig.Price
		}(),
		Quantity: func() int {
			if productRequest.Quantity != nil {
				return *productRequest.Quantity
			}
			return orig.Quantity
		}(),
		InventoryType: func() string {
			if productRequest.InventoryType != nil {
				return *productRequest.InventoryType
			}
			return orig.InventoryType
		}(),
		Sizes: sizes,
		SubDescription: func() string {
			if productRequest.SubDescription != nil {
				return *productRequest.SubDescription
			}
			return orig.SubDescription
		}(),
		Description: func() string {
			if productRequest.Description != nil {
				return *productRequest.Description
			}
			return orig.Description
		}(),
	}
}

// processImagesForUpdate processes the images for update, using requested images if provided or original otherwise
func (s *ProductService) processStringForUpdate(req []*string, orig []string) (*postgres.Jsonb, error) {
	if req != nil {
		return s.transferImagesToJsonB(req)
	}
	var origPtrs []*string
	for _, img := range orig {
		imgCopy := img
		origPtrs = append(origPtrs, &imgCopy)
	}
	return s.transferImagesToJsonB(origPtrs)
}
