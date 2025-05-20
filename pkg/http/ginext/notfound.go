package ginext

import (
	"emission/pkg/http/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NotFoundHandler(c *gin.Context) {
	log := logger.WithCtx(c, "notfound")
	log.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
	})

	c.Status(http.StatusNotFound)
	c.Header("content-type", "application/json")
	_, _ = c.Writer.WriteString(`{"error": {"route": "not found"}}`)
}
