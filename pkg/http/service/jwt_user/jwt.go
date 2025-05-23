package jwt_user

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	model "kimistore/internal/models"
	"kimistore/internal/utils"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/logger"
	"strconv"
	"time"
)

const (
	UserAccess  = "access"
	UserRefresh = "refresh"
)

// Claims is a struct that contains the claims of the JWT
type Claims struct {
	ID        string `json:"id"`
	TokenType string `json:"type"`
	jwt.RegisteredClaims
}

var UserTokenTypeKeyName = map[string]string{
	UserAccess:  "SecureUser.JWTAccessSecure",
	UserRefresh: "SecureUser.JWTRefreshSecure",
}

var UserCarTokenTypeExpTime = map[string]string{
	UserAccess:  "SecureUser.JWTAccessTimeMinute",
	UserRefresh: "SecureUser.JWTRefreshTimeHour",
}

// GenerateJWTToken generates a JWT token (refresh or access)
func GenerateJWTTokenUser(context context.Context,
	userID string,
	tokenType string,
	config *viper.Viper) (appToken *AppToken, err error) {
	log := logger.WithCtx(context, "GenerateJWTTokenUser")

	JWTSecureKey := config.GetString(UserTokenTypeKeyName[tokenType])
	JWTExpTime := config.GetString(UserCarTokenTypeExpTime[tokenType])

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
		ID:        userID,
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

func DecodeUserJwtToID(ctx *gin.Context, config *viper.Viper) int {
	JWTAccessSecure := config.GetString("SecureUser.JWTAccessSecure")
	authHeader := ctx.GetHeader("Authorization")

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		_ = ctx.Error(errors.New("invalid authorization header"))
		return 0
	}

	tokenString := authHeader[7:]
	signature := []byte(JWTAccessSecure)

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return signature, nil
	})

	if err != nil {
		_ = ctx.Error(err)
		return 0
	}

	id, ok := claims["id"].(float64)
	if !ok {
		_ = ctx.Error(errors.New("invalid token claims"))
		return 0
	}
	return int(id)
}

func GetClaimsUserAndVerifyToken(tokenString string, tokenType string, config *viper.Viper) (claims jwt.MapClaims, err error) {

	JWTRefreshSecure := config.GetString(UserTokenTypeKeyName[tokenType])
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, app_errors.AppError("Fail to Authorized", app_errors.StatusUnauthorized)
		}
		return []byte(JWTRefreshSecure), nil
	})
	if err != nil {
		return nil, app_errors.AppError("Time out", app_errors.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["type"] != tokenType {
			return nil, app_errors.AppError("Fail to Authorized", app_errors.StatusUnauthorized)
		}

		var timeExpire = claims["exp"].(float64)
		if time.Now().Unix() > int64(timeExpire) {
			return nil, app_errors.AppError("Time out", app_errors.StatusUnauthorized)
		}
		return claims, nil
	}
	return nil, app_errors.AppError("Fail to Authorized", app_errors.StatusUnauthorized)
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
