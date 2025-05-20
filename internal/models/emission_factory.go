package models

// EmissionFactor represents the emission factor of a country
type EmissionFactor struct {
	BaseModel
	CountryCode string  `json:"country_code"`
	Value       float64 `json:"value"`
}
