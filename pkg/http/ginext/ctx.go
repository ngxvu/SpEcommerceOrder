package ginext

import (
	"context"

	"github.com/gin-gonic/gin"
)

// FromGinRequestContext makes a new context from Gin request context, copy x-request-id to if any
// use request context instead of gin context to handle user cancelling
func FromGinRequestContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	if requestID := c.GetString("x-request-id"); requestID != "" {
		ctx = context.WithValue(ctx, "x-request-id", requestID)
	} else if requestID = c.GetHeader("x-request-id"); requestID != "" {
		ctx = context.WithValue(ctx, "x-request-id", requestID)
	}
	return ctx
}
