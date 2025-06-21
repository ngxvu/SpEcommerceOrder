package models

import "basesource/internal/utils"

// model for storage image link
type Media struct {
	BaseModel
	MediaURL    string  `json:"media_url" gorm:"type:text;not null"`
	MediaFormat string  `json:"media_format" gorm:"type:varchar(50);default:'jpg';not null"` // e.g., JPEG, PNG
	MediaSize   float64 `json:"media_size" gorm:"type:decimal(10,2);not null"`               // size of media
	MediaWidth  int     `json:"media_width" gorm:"type:int;not null"`                        // width of media
	MediaHeight int     `json:"media_height" gorm:"type:int;not null"`                       // height of media
}

func (Media) TableName() string {
	return "media"
}

type ImageSaveResponse struct {
	Meta *utils.MetaData          `json:"meta"`
	Data MediaInformationResponse `json:"data"`
}

type MediaInformationResponse struct {
	ID          string  `json:"id"`
	MediaURL    string  `json:"media_url"`
	MediaFormat string  `json:"media_format"` // e.g., JPEG, PNG
	MediaSize   float64 `json:"media_size"`   // size of media
	MediaWidth  int     `json:"media_width"`  // width of media
	MediaHeight int     `json:"media_height"` // height of media
}

type ListImageSaveResponse struct {
	Meta *utils.MetaData             `json:"meta"`
	Data []*MediaInformationResponse `json:"data"`
}
