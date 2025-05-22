package models

type Post struct {
	BaseModel
	Publish     string `json:"publish" gorm:"column:publish;default:draft"`
	Title       string `json:"title" gorm:"column:title"`
	CoverURL    string `json:"cover_url" gorm:"column:cover_url"`
	Description string `json:"description" gorm:"column:description"`
}
