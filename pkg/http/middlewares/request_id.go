package middlewares

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("x-request-id")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("x-request-id", requestID)
		}
		c.Set("x-request-id", requestID)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "x-request-id", requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
