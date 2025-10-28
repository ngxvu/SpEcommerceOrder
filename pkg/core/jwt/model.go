package jwt

import (
	"basesource/internal/utils"
	"time"
)

type JWTUserDataResponse struct {
	Meta *utils.MetaData      `json:"meta"`
	Data JWTTokenResponseData `json:"data"`
}

type JWTUserLogoutResponse struct {
	Meta *utils.MetaData `json:"meta"`
	Data string          `json:"data"`
}

type AppToken struct {
	Token          string    `json:"token"`
	TokenType      string    `json:"tokenType"`
	ExpirationTime time.Time `json:"expirationTime"`
}

type JWTTokenResponseData struct {
	JWTAccessToken            string      `json:"jwtAccessToken"`
	JWTRefreshToken           string      `json:"jwtRefreshToken"`
	ExpirationAccessDateTime  time.Time   `json:"expirationAccessDateTime"`
	ExpirationRefreshDateTime time.Time   `json:"expirationRefreshDateTime"`
	Profile                   interface{} `json:"profile"`
}

type DataUserAuthenticated struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
