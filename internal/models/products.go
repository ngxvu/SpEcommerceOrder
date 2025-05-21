package models

import "time"

type Product struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	CoverURL       string    `json:"coverUrl"`
	Images         []string  `json:"images"`
	Publish        string    `json:"publish"`
	Name           string    `json:"name"`
	Price          float64   `json:"price"`
	Sizes          []string  `json:"sizes"`
	SubDescription string    `json:"subDescription"`
	Description    string    `json:"description"`
}
