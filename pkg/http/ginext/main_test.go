package ginext

import (
	"kimistore/pkg/http/logger"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	logger.Init("ginext.test")
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
