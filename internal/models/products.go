package models

import (
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Product struct {
	BaseModel
	CoverURL       string          `json:"cover_url" gorm:"type:text"`
	Images         *postgres.Jsonb `json:"images" gorm:"type:jsonb"`
	Publish        string          `json:"publish" gorm:"type:text"`
	Name           string          `json:"name" gorm:"type:varchar(255);not null"`
	Price          float64         `json:"price" gorm:"type:decimal(10,2);not null"`
	Sizes          *postgres.Jsonb `json:"sizes" gorm:"type:jsonb"`
	SubDescription string          `json:"sub_description"`
	Description    string          `json:"description"`
}
