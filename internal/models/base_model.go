package models

import (
	"emission/pkg/http/ginext"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

type EmissionListRequest struct {
	FactoryID *uuid.UUID `form:"factory_id"`
	From      *time.Time `form:"from" form:"from"`
	To        *time.Time `form:"to" form:"to"`
}

type EmissionFilter struct {
	EmissionListRequest
	Pager *ginext.Pager
}

type EmissionFilterResult struct {
	Filter  *EmissionFilter
	Records []Emission
}

type EmissionCreateRequest struct {
	FactoryID              *uuid.UUID `json:"factory_id" valid:"Required"`
	ElectricityConsumption *float64   `json:"electricity_consumption" valid:"Required"`
}

type FactoryRequest struct {
	Name        *string `json:"name" valid:"Required"`
	CountryCode *string `json:"country_code" valid:"Required"`
}

type CreateFactoryResponse struct {
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
	ApiKey      string `json:"api_key"`
}

type CreateTokenRequest struct {
	ObjectID string
	NumHour  int
}

type AccessTokenClaims struct {
	jwt.StandardClaims
	ObjectID string
}
