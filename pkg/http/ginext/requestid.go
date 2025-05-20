package ginext

import (
	. "emission/pkg/http/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware(c *gin.Context) {
	requestid := c.GetHeader(HeaderXRequestID)
	if requestid == "" {
		requestid = uuid.New().String()
		c.Request.Header.Set(HeaderXRequestID, requestid)
	}
	// set to context
	c.Set(HeaderXRequestID, requestid)

	// set to response header as well
	c.Header(HeaderXRequestID, requestid)

	c.Next()
}
