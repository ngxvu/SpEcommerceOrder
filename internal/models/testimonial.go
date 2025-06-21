package models

import (
	"basesource/internal/utils"
	"basesource/pkg/http/paging"
)

type Testimonial struct {
	BaseModel
	Author  string `json:"author" gorm:"type:varchar(255);not null"`
	Content string `json:"content" gorm:"type:text;not null"`
}

func (Testimonial) TableName() string {
	return "testimonials"
}

type TestimonialRequest struct {
	Author  *string `json:"author" binding:"required"`
	Content *string `json:"content" binding:"required"`
}

type GetTestimonialResponse struct {
	Meta *utils.MetaData `json:"meta"`
	Data *Testimonial    `json:"data"`
}

type ListTestimonialResponse struct {
	Filter  *paging.Filter
	Records []Testimonial `json:"products"`
}

type DeleteTestimonialResponse struct {
	Meta *utils.MetaData `json:"meta"`
	Data string          `json:"data"`
}
