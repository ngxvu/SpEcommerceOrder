package models

import (
	"basesource/internal/utils"
	"basesource/pkg/http/paging"
	"github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

type Product struct {
	BaseModel
	CoverURL       string          `json:"cover_url" gorm:"type:text"`
	Images         *postgres.Jsonb `json:"images" gorm:"type:jsonb"`
	Publish        string          `json:"publish" gorm:"type:text;default:'draft'"`
	Name           string          `json:"name" gorm:"type:varchar(255);not null"`
	Price          float64         `json:"price" gorm:"type:decimal(10,2);not null"`
	Quantity       int             `json:"quantity" gorm:"type:int;default:0"`
	InventoryType  string          `json:"inventory_type" gorm:"type:text;default:'in stock'"`
	Sizes          *postgres.Jsonb `json:"sizes" gorm:"type:jsonb"`
	SubDescription string          `json:"sub_description;default:null"`
	Description    string          `json:"description;default:null"`
}

type OriginalProduct struct {
	ID      string   `json:"id"`
	Gender  []string `json:"gender"`
	Images  []string `json:"images"`
	Reviews []struct {
		ID          string        `json:"id"`
		Name        string        `json:"name"`
		PostedAt    time.Time     `json:"postedAt"`
		Comment     string        `json:"comment"`
		IsPurchased bool          `json:"isPurchased"`
		Rating      float64       `json:"rating"`
		AvatarURL   string        `json:"avatarUrl"`
		Helpful     int           `json:"helpful"`
		Attachments []interface{} `json:"attachments"`
	} `json:"reviews"`
	Publish string `json:"publish"`
	Ratings []struct {
		Name        string `json:"name"`
		StarCount   int    `json:"starCount"`
		ReviewCount int    `json:"reviewCount"`
	} `json:"ratings"`
	Category      string      `json:"category"`
	Available     int         `json:"available"`
	PriceSale     interface{} `json:"priceSale"`
	Taxes         int         `json:"taxes"`
	Quantity      int         `json:"quantity"`
	InventoryType string      `json:"inventoryType"`
	Tags          []string    `json:"tags"`
	Code          string      `json:"code"`
	Description   string      `json:"description"`
	Sku           string      `json:"sku"`
	CreatedAt     time.Time   `json:"createdAt"`
	Name          string      `json:"name"`
	Price         float64     `json:"price"`
	CoverURL      string      `json:"coverUrl"`
	Colors        []string    `json:"colors"`
	TotalRatings  float64     `json:"totalRatings"`
	TotalSold     int         `json:"totalSold"`
	TotalReviews  int         `json:"totalReviews"`
	NewLabel      struct {
		Enabled bool   `json:"enabled"`
		Content string `json:"content"`
	} `json:"newLabel"`
	SaleLabel struct {
		Enabled bool   `json:"enabled"`
		Content string `json:"content"`
	} `json:"saleLabel"`
	Sizes          []string `json:"sizes"`
	SubDescription string   `json:"subDescription"`
}

type CreateProductRequest struct {
	CoverURL       *string   `json:"cover_url" binding:"required"`
	Images         []*string `json:"images" binding:"required"`
	Publish        *string   `json:"publish" binding:"required"`
	Name           *string   `json:"name" binding:"required"`
	Price          *float64  `json:"price" binding:"required"`
	Quantity       *int      `json:"quantity" binding:"required"`
	InventoryType  *string   `json:"inventory_type" binding:"required"`
	Sizes          []*string `json:"sizes" binding:"required"`
	SubDescription *string   `json:"sub_description"`
	Description    *string   `json:"description"`
}

type GetProductResponseData struct {
	Product OriginalProduct `json:"product"`
}

type GetProductResponse struct {
	Meta *utils.MetaData        `json:"meta"`
	Data GetProductResponseData `json:"data"`
}

type ProductFilterRequest struct {
	DefaultSearch   *string `json:"default_search" form:"default_search"`
	SearchByStock   *string `json:"search_by_stock" form:"search_by_stock"`
	SearchByPrice   *string `json:"search_by_price" form:"search_by_price"`
	SearchByYear    *string `json:"search_by_year" form:"search_by_year"`
	SearchByPublish *string `json:"search_by_publish" form:"search_by_publish"`
	FilterByStock   *string `json:"filter_by_stock" form:"filter_by_stock"`
	FilterByPublish *string `json:"filter_by_publish" form:"filter_by_publish"`
}

type ListProductFilter struct {
	ProductFilterRequest
	Pager *paging.Pager
}

type ListProductResult struct {
	Filter  *ListProductFilter
	Records []Product `json:"products"`
}

type ListProductResponse struct {
	Filter  *ListProductFilter
	Records []OriginalProduct `json:"products"`
}

// ColumnFilterParam defines a complete filter parameter including column, operator, and values
type ColumnFilterParam struct {
	Column   string   `json:"column" binding:"required"`   // Column name to filter on
	Operator string   `json:"operator" binding:"required"` // Filter operator
	Value    string   `json:"value"`                       // Single value
	Values   []string `json:"values"`                      // Multiple values
	Page     int      `json:"page" binding:"required"`
	PageSize int      `json:"page_size" binding:"required"`
}

type UpdateProductRequest struct {
	CoverURL       *string   `json:"cover_url"`
	Images         []*string `json:"images"`
	Publish        *string   `json:"publish"`
	Name           *string   `json:"name"`
	Price          *float64  `json:"price"`
	Sizes          []*string `json:"sizes"`
	Quantity       *int      `json:"quantity"`
	InventoryType  *string   `json:"inventory_type"`
	SubDescription *string   `json:"sub_description"`
	Description    *string   `json:"description"`
}
type DeleteProductResponse struct {
	Meta    *utils.MetaData `json:"meta"`
	Message string          `json:"data"`
}
