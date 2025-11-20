package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"order/pkg/core/configloader"
	"order/pkg/http/utils/errors"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		config := configloader.GetConfig()
		JWTAccessSecure := config.JWTAccessSecure
		authHeader := c.GetHeader("Authorization")
		signature := []byte(JWTAccessSecure)

		if authHeader == "" {
			err := errors.Error("fail to authenticate", errors.StatusValidationError)
			_ = c.Error(err)
			c.Abort()
			return
		}

		authHeaderParts := strings.Split(authHeader, "Bearer ")
		if len(authHeaderParts) != 2 {
			err := errors.Error("fail to authenticate", errors.StatusValidationError)
			_ = c.Error(err)
			c.Abort()
			return
		}

		tokenString := authHeaderParts[1]
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return signature, nil
		})

		if err != nil {
			err := errors.Error("fail to authenticate", errors.StatusValidationError)
			_ = c.Error(err)
			c.Abort()
			return
		}

		claimsRole := claims["role"]
		role := fmt.Sprintf("%v", claimsRole)

		c.Set("role", role)

		if !authorize(role) {
			err := errors.Error("you are not authorized to perform this action", errors.StatusForbidden)
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Next()
	}
}

func authorize(userRole string) bool {
	return userRole == "admin"
}
