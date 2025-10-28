package jwt

import (
	"basesource/conf"
	model "basesource/internal/models"
	"basesource/internal/utils"
	"basesource/internal/utils/app_errors"
	"basesource/pkg/core/logger"
	"context"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"time"
)

const (
	UserAccess  = "access"
	UserRefresh = "refresh"
)

// Claims is a struct that contains the claims of the JWT
type Claims struct {
	Role      string `json:"role"`
	TokenType string `json:"type"`
	jwt.RegisteredClaims
}

// GenerateJWTToken generates a JWT token (refresh or access)
func GenerateJWTTokenUser(context context.Context,
	userRole string,
	tokenType string) (appToken *AppToken, err error) {

	log := logger.WithCtx(context, "GenerateJWTTokenUser")

	config := conf.GetConfig()

	JWTSecureKey := config.JWTAccessSecure
	JWTExpTime := config.JWTAccessTimeMinute

	tokenTimeConverted, err := strconv.ParseInt(JWTExpTime, 10, 64)
	if err != nil {
		return
	}

	tokenTimeUnix := time.Duration(tokenTimeConverted)
	switch tokenType {
	case UserRefresh:
		tokenTimeUnix *= time.Hour
	case UserAccess:
		tokenTimeUnix *= time.Minute

	default:
		err = app_errors.AppError("Fail to Authorized", app_errors.StatusUnauthorized)
		logger.LogError(log, err, "invalid token type")
	}

	if err != nil {
		return nil, err
	}
	nowTime := time.Now()
	expirationTokenTime := nowTime.Add(tokenTimeUnix)

	tokenClaims := &Claims{
		Role:      userRole,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTokenTime),
		},
	}
	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	// Sign and get the complete encoded token as a string using the secret
	tokenStr, err := tokenWithClaims.SignedString([]byte(JWTSecureKey))
	if err != nil {
		return
	}

	appToken = &AppToken{
		Token:          tokenStr,
		TokenType:      tokenType,
		ExpirationTime: expirationTokenTime,
	}
	return
}

func SecAuthUserMapper(user *model.User,
	accessTokenClaims,
	refreshTokenClaims *AppToken) *JWTTokenResponseData {

	userID := utils.UUIDtoString(user.ID)

	return &JWTTokenResponseData{
		JWTAccessToken:            accessTokenClaims.Token,
		JWTRefreshToken:           refreshTokenClaims.Token,
		ExpirationAccessDateTime:  accessTokenClaims.ExpirationTime,
		ExpirationRefreshDateTime: refreshTokenClaims.ExpirationTime,
		Profile: DataUserAuthenticated{
			ID:    userID,
			Name:  user.Name,
			Email: user.Email,
		},
	}
}
