package models

import (
	"github.com/jinzhu/gorm/dialects/postgres"
	"kimistore/internal/utils"
	"kimistore/pkg/http/paging"
	"time"
)

type Product struct {
	BaseModel
	CoverURL       string          `json:"cover_url" gorm:"type:text"`
	Images         *postgres.Jsonb `json:"images" gorm:"type:jsonb"`
	Publish        string          `json:"publish" gorm:"type:text;default:'draft'"`
	Name           string          `json:"name" gorm:"type:varchar(255);not null"`
	Price          float64         `json:"price" gorm:"type:decimal(10,2);not null"`
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
	Publish        *string   `json:"publish"`
	Name           *string   `json:"name" binding:"required"`
	Price          *float64  `json:"price" binding:"required"`
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

type ListProductResult struct {
	Filter  *paging.Filter
	Records []Product `json:"products"`
}

type ListProductResponse struct {
	Filter  *paging.Filter
	Records []OriginalProduct `json:"products"`
}

type UpdateProductRequest struct {
	CoverURL       *string   `json:"cover_url"`
	Images         []*string `json:"images"`
	Publish        *string   `json:"publish"`
	Name           *string   `json:"name"`
	Price          *float64  `json:"price"`
	Sizes          []*string `json:"sizes"`
	SubDescription *string   `json:"sub_description"`
	Description    *string   `json:"description"`
}
type DeleteProductResponse struct {
	Meta    *utils.MetaData `json:"meta"`
	Message string          `json:"data"`
}
