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
	GetListProduct(filter *paging.Filter, tx *gorm.DB) (*model.ListProductResponse, error)
	UpdateProduct(ctx context.Context, tx *gorm.DB, id string, product model.Product) (*model.GetProductResponse, error)
	DeleteProduct(ctx context.Context, tx *gorm.DB, id string) error
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
			("deleted_at","cover_url","images","publish","name","price","sizes","sub_description","description") 
		VALUES 
			(NULL, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	imagesJSON, _ := json.Marshal(product.Images)
	sizesJSON, _ := json.Marshal(product.Sizes)

	var productID string
	err := tx.Raw(rawSQL,
		product.CoverURL,
		string(imagesJSON),
		product.Publish,
		product.Name,
		product.Price,
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

func (r *ProductRepository) GetListProduct(filter *paging.Filter, pgRepo *gorm.DB) (*model.ListProductResponse, error) {

	log := logger.WithTag("ProductRepository|GetListProduct")

	tx := pgRepo.Model(&model.Product{})

	result := &model.ListProductResult{
		Filter:  filter,
		Records: []model.Product{},
	}

	filter.Pager.SortableFields = []string{"name"}

	pager := filter.Pager

	err := pager.DoQuery(&result.Records, tx).Error
	if err != nil {
		err := app_errors.AppError("Error getting list product", app_errors.StatusNotFound)
		logger.LogError(log, err, "Error when getting list internal price group")
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

func (r *ProductRepository) DeleteProduct(ctx context.Context, tx *gorm.DB, id string) error {
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
		return err
	}

	return nil
}
