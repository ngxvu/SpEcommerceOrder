package models

// model for storage image link
type Media struct {
	BaseModel
	URL  string `json:"url" gorm:"column:url;type:text"`
	Type string `json:"type" gorm:"column:type;type:text"`
	Size int64  `json:"size" gorm:"column:size;type:bigint"`
	Name string `json:"name" gorm:"column:name;type:text"`
}
