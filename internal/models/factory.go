package models

// Factory struct represents the factory model

type Factory struct {
	BaseModel
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
}
