package models

import "time"

type Post struct {
	ID          string    `json:"id"`
	Publish     string    `json:"publish"`
	CreatedAt   time.Time `json:"createdAt"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"coverUrl"`
	Description string    `json:"description"`
}
