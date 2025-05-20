package ginext

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestContextExtractWithRequestID(t *testing.T) {
	t.Run("RequestIDFromValue", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("x-request-id", "test-request-id")

		ctx := FromGinRequestContext(c)
		assert.Equal(t, "test-request-id", ctx.Value("x-request-id").(string))
	})

	t.Run("RequestIDFromHeader", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("x-request-id", "test-request-2")
		ctx := FromGinRequestContext(c)
		assert.Equal(t, "test-request-2", ctx.Value("x-request-id").(string))
	})
}
