package repo

import (
	"context"
	"encoding/json"
	"gorm.io/gorm"
	model "kimistore/internal/models"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/paging"
	"time"
)

type ProductRepository struct {
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

type ProductRepositoryInterface interface {
	ProductExistsByName(tx *gorm.DB, name string) (bool, error)
	CreateProduct(ctx context.Context, tx *gorm.DB, product model.Product) (*model.GetProductResponse, error)
	GetDetailProduct(ctx context.Context, tx *gorm.DB, id string) (*model.GetProductResponse, error)
	GetListProduct(filter *model.ListProductFilter, tx *gorm.DB) (*model.ListProductResponse, error)
	FilterColumnProduct(filter *model.ColumnFilterParam, pager *paging.Pager, tx *gorm.DB) (*model.ListProductResponse, error)
	UpdateProduct(ctx context.Context, tx *gorm.DB, id string, product model.Product) (*model.GetProductResponse, error)
	DeleteProduct(ctx context.Context, tx *gorm.DB, id string) (*model.DeleteProductResponse, error)
}

func (r *ProductRepository) ProductExistsByName(tx *gorm.DB, name string) (bool, error) {
	var count int64
	if err := tx.Model(&model.Product{}).Where("name = ?", name).Count(&count).Error; err != nil {
		err := app_errors.AppError("Error checking product existence by name", app_errors.StatusBadRequest)
		logger.LogError(logger.WithTag("ProductRepository|ProductExistsByName"), err, "Error checking product existence by name")
		return false, err
	}
	return count > 0, nil
}

func (r *ProductRepository) CreateProduct(ctx context.Context, tx *gorm.DB, product model.Product) (*model.GetProductResponse, error) {

	log := logger.WithTag("ProductRepository|CreateProduct")

	rawSQL := `
	INSERT INTO "products" 
		("deleted_at","cover_url","images","publish","name","price","quantity","inventory_type","sizes","sub_description","description") 
	VALUES 
		(NULL, ?, ?::jsonb, ?::publish, ?, ?, ?, ?::inventory_status, ?::jsonb, ?, ?)
	RETURNING id
`

	imagesJSON, _ := json.Marshal(product.Images)
	sizesJSON, _ := json.Marshal(product.Sizes)

	// Validate publish value
	if product.Publish != "draft" && product.Publish != "published" {
		err := app_errors.AppError("Invalid publish value. Must be 'draft' or 'published'", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Invalid publish value")
		return nil, err
	}

	// Validate inventory_type value
	if product.InventoryType != "in stock" && product.InventoryType != "out of stock" && product.InventoryType != "low stock" {
		err := app_errors.AppError("Invalid inventory_type value. Must be 'in stock', 'out of stock', or 'low stock'", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Invalid inventory_type value")
		return nil, err
	}

	var productID string
	err := tx.Raw(rawSQL,
		product.CoverURL,
		string(imagesJSON),
		product.Publish,
		product.Name,
		product.Price,
		product.Quantity,
		product.InventoryType,
		string(sizesJSON),
		product.SubDescription,
		product.Description,
	).Row().Scan(&productID)

	if err != nil {
		err := app_errors.AppError("Error creating product", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Error creating product")
		return nil, err
	}

	// Set the ID on the product before mapping
	getProductResponse, err := r.GetDetailProduct(ctx, tx, productID)
	if err != nil {
		err := app_errors.AppError("Error getting product by ID", app_errors.StatusNotFound)
		logger.LogError(log, err, "Error getting product by ID")
		return nil, err
	}

	return &model.GetProductResponse{
		Meta: utils.NewMetaData(ctx),
		Data: getProductResponse.Data,
	}, nil
}

func (r *ProductRepository) GetDetailProduct(ctx context.Context, tx *gorm.DB, id string) (*model.GetProductResponse, error) {
	log := logger.WithTag("ProductRepository|GetProductByID")

	var product model.Product
	if err := tx.Where("id = ?", id).First(&product).Error; err != nil {
		err := app_errors.AppError("Error getting product by ID", app_errors.StatusNotFound)
		logger.LogError(log, err, "Error getting product by ID")
		return nil, err
	}

	mapper := r.mapProductToResponseData(product)

	return &model.GetProductResponse{
		Meta: utils.NewMetaData(ctx),
		Data: mapper,
	}, nil
}

func (r *ProductRepository) GetListProduct(filter *model.ListProductFilter, pgRepo *gorm.DB) (*model.ListProductResponse, error) {
	log := logger.WithTag("ProductRepository|GetListProduct")

	tx := pgRepo.Model(&model.Product{})

	result := &model.ListProductResult{
		Filter:  filter,
		Records: []model.Product{},
	}

	if filter.DefaultSearch != nil {
		searchTerm := "%" + *filter.DefaultSearch + "%"
		tx = tx.Where("name ILIKE ?", searchTerm)
	}

	if filter.SearchByStock != nil {
		searchStock := "%" + *filter.SearchByStock + "%"
		tx = tx.Where("inventory_type ILIKE ?", searchStock)
	}

	if filter.SearchByPrice != nil {
		searchPrice := *filter.SearchByPrice
		tx = tx.Where("price::text LIKE ?", "%"+searchPrice+"%")
	}

	if filter.SearchByPublish != nil {
		searchPublish := *filter.SearchByPublish
		tx = tx.Where("publish LIKE ?", "%"+searchPublish+"%")
	}

	if filter.SearchByYear != nil {
		searchYear := *filter.SearchByYear
		tx = tx.Where("EXTRACT(YEAR FROM created_at) = ?", searchYear)
	}

	if filter.FilterByStock != nil {
		filterByStock := *filter.FilterByStock
		if filterByStock == "in_stock" || filterByStock == "out_of_stock" || filterByStock == "low_stock" {
			tx = tx.Where("inventory_type = ?", filterByStock)
		} else {
			err := app_errors.AppError("Must be 'in_stock', 'out_of_stock', or 'low_stock'", app_errors.StatusBadRequest)
			logger.LogError(log, err, "Invalid filter_by_stock value")
			return nil, err
		}
	}

	if filter.FilterByPublish != nil {
		filterByPublish := *filter.FilterByPublish
		if filterByPublish == "draft" || filterByPublish == "published" {
			tx = tx.Where("publish = ?", filterByPublish)
		} else {
			err := app_errors.AppError("Must be 'draft' or 'published'", app_errors.StatusBadRequest)
			logger.LogError(log, err, "Invalid filter_by_publish value")
			return nil, err
		}
	}

	pager := filter.Pager

	err := pager.DoQuery(&result.Records, tx).Error
	if err != nil {
		err := app_errors.AppError("Error getting list product", app_errors.StatusNotFound)
		logger.LogError(log, err, "Error when getting list products")
		return nil, err
	}

	mapper := model.GetProductResponseData{}

	var mapperList []model.OriginalProduct

	for i := 0; i < len(result.Records); i++ {
		mapper = r.mapProductToResponseData(result.Records[i])
		mapperList = append(mapperList, mapper.Product)
	}

	response := &model.ListProductResponse{
		Filter:  filter,
		Records: mapperList,
	}

	return response, nil
}

// Add this method to your ProductRepository struct in product.go

func (r *ProductRepository) FilterColumnProduct(
	filter *model.ColumnFilterParam,
	pager *paging.Pager,
	tx *gorm.DB) (*model.ListProductResponse, error) {

	log := logger.WithTag("ProductRepository|FilterColumnProduct")

	// Start with the base query
	query := tx.Model(&model.Product{})

	// Apply the column filter
	query = r.applyColumnFilter(query, filter.Column, filter.Operator, filter.Value, filter.Values)

	// Execute the query with pagination
	var products []model.Product
	if err := pager.DoQuery(&products, query).Error; err != nil {
		err := app_errors.AppError("Error filtering products", app_errors.StatusInternalServerError)
		logger.LogError(log, err, "Error executing filter query")
		return nil, err
	}

	// Map products to response format
	var origProducts []model.OriginalProduct
	for _, p := range products {
		mapper := r.mapProductToResponseData(p)
		origProducts = append(origProducts, mapper.Product)
	}

	// Create a response structure
	response := &model.ListProductResponse{
		Records: origProducts,
	}

	return response, nil
}

func (r *ProductRepository) mapProductToResponseData(product model.Product) model.GetProductResponseData {
	// Parse Images and Sizes from postgres.Jsonb to string slices
	var images []string
	var sizes []string

	if product.Images != nil {
		if err := json.Unmarshal(product.Images.RawMessage, &images); err != nil {
			images = []string{}
		}
	}

	if product.Sizes != nil {
		if err := json.Unmarshal(product.Sizes.RawMessage, &sizes); err != nil {
			sizes = []string{}
		}
	}

	productID := utils.UUIDtoString(product.ID)

	responseData := model.GetProductResponseData{}
	responseData.Product.ID = productID
	responseData.Product.Name = product.Name
	responseData.Product.Price = product.Price
	responseData.Product.Quantity = product.Quantity
	responseData.Product.InventoryType = product.InventoryType
	responseData.Product.CoverURL = product.CoverURL
	responseData.Product.Publish = product.Publish
	responseData.Product.Description = product.Description
	responseData.Product.SubDescription = product.SubDescription
	responseData.Product.CreatedAt = product.CreatedAt
	responseData.Product.Images = images
	responseData.Product.Sizes = sizes

	// Set default values for other fields that don't exist in Product
	responseData.Product.Gender = []string{}
	responseData.Product.Reviews = []struct {
		ID          string        `json:"id"`
		Name        string        `json:"name"`
		PostedAt    time.Time     `json:"postedAt"`
		Comment     string        `json:"comment"`
		IsPurchased bool          `json:"isPurchased"`
		Rating      float64       `json:"rating"`
		AvatarURL   string        `json:"avatarUrl"`
		Helpful     int           `json:"helpful"`
		Attachments []interface{} `json:"attachments"`
	}{}
	responseData.Product.Ratings = []struct {
		Name        string `json:"name"`
		StarCount   int    `json:"starCount"`
		ReviewCount int    `json:"reviewCount"`
	}{}
	responseData.Product.Colors = []string{}
	responseData.Product.Tags = []string{}

	return responseData
}

func (r *ProductRepository) applyColumnFilter(tx *gorm.DB, column, operator, value string, values []string) *gorm.DB {
	if column == "created_at" {
		switch operator {
		case "contains":
			// Extract year from timestamp for "contains" year search
			return tx.Where("EXTRACT(YEAR FROM created_at)::text LIKE ?", "%"+value+"%")
		case "does_not_contains":
			// Exclude timestamps where year contains the value
			return tx.Where("EXTRACT(YEAR FROM created_at)::text NOT LIKE ?", "%"+value+"%")
		case "equals":
			return tx.Where(column+" = ?", value)
		case "does_not_equals":
			return tx.Where(column+" != ?", value)
		case "starts_with":
			// Handle starts_with for date (e.g., starts with 2023)
			return tx.Where("TO_CHAR(created_at, 'YYYY-MM-DD') LIKE ?", value+"%")
		case "ends_with":
			// Handle ends_with for date (e.g., ends with -31 for last day of month)
			return tx.Where("TO_CHAR(created_at, 'YYYY-MM-DD') LIKE ?", "%"+value)
		case "is_empty":
			return tx.Where(column + " IS NULL")
		case "is_not_empty":
			return tx.Where(column + " IS NOT NULL")
		case "is_any_of":
			if len(values) > 0 {
				return tx.Where(column+" IN (?)", values)
			}
		}
	}

	switch operator {
	case "contains":
		return tx.Where(column+" ILIKE ?", "%"+value+"%")
	case "does_not_contains":
		return tx.Where(column+" NOT ILIKE ?", "%"+value+"%")
	case "equals":
		return tx.Where(column+" = ?", value)
	case "does_not_equals":
		return tx.Where(column+" != ?", value)
	case "starts_with":
		return tx.Where(column+" ILIKE ?", value+"%")
	case "ends_with":
		return tx.Where(column+" ILIKE ?", "%"+value)
	case "is_empty":
		return tx.Where(column + " = '' OR " + column + " IS NULL")
	case "is_not_empty":
		return tx.Where(column + " != '' AND " + column + " IS NOT NULL")
	case "is_any_of":
		if len(values) > 0 {
			return tx.Where(column+" IN (?)", values)
		}
	}
	return tx
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, tx *gorm.DB, id string, product model.Product) (*model.GetProductResponse, error) {
	log := logger.WithTag("ProductRepository|UpdateProduct")

	// Update the product in the database using raw SQL
	rawSQL := `
    UPDATE "products"
    SET cover_url = ?, images = ?, publish = ?, name = ?, price = ?, sizes = ?, sub_description = ?, description = ?
    WHERE id = ?
`

	imagesJSON, _ := json.Marshal(product.Images)
	sizesJSON, _ := json.Marshal(product.Sizes)

	result := tx.Exec(rawSQL,
		product.CoverURL,
		string(imagesJSON),
		product.Publish,
		product.Name,
		product.Price,
		string(sizesJSON),
		product.SubDescription,
		product.Description,
		id,
	)
	if result.Error != nil {
		err := app_errors.AppError("Error updating product", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Error updating product")
		return nil, err
	}

	detailProduct, err := r.GetDetailProduct(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return &model.GetProductResponse{
		Meta: utils.NewMetaData(ctx),
		Data: detailProduct.Data,
	}, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, tx *gorm.DB, id string) (*model.DeleteProductResponse, error) {
	log := logger.WithTag("ProductRepository|DeleteProduct")

	// Delete the product from the database using raw SQL
	rawSQL := `
	DELETE FROM "products"
	WHERE id = ?
`

	result := tx.Exec(rawSQL, id)
	if result.Error != nil {
		err := app_errors.AppError("Error deleting product", app_errors.StatusBadRequest)
		logger.LogError(log, err, "Error deleting product")
		return nil, err
	}

	return &model.DeleteProductResponse{
		Meta:    utils.NewMetaData(ctx),
		Message: "Product deleted successfully",
	}, nil
}
