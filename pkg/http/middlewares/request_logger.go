package middlewares

import (
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"time"
)

func RequestLogger(appName string) gin.HandlerFunc {
	writerInfo, err := rotatelogs.New(
		"./log_files/info/"+appName+"_%Y%m%d_info.log",
		rotatelogs.WithMaxAge(30*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Fatalf("Failed to create rotatelogs: %s", err)
	}

	writerError, err := rotatelogs.New(
		"./log_files/app_errors/"+appName+"_%Y%m%d_error.log",
		rotatelogs.WithMaxAge(30*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Fatalf("Failed to create rotatelogs: %s", err)
	}

	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(startTime)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		requestID := c.GetString("x-request-id")

		logEntry := log.New(writerInfo, "", log.LstdFlags)
		if status >= 400 {
			logEntry = log.New(writerError, "", log.LstdFlags)
		}

		logEntry.Printf("| %3d | %13v | %15s | %-7s %s | %s",
			status, latency, clientIP, method, path, requestID)
	}
}
