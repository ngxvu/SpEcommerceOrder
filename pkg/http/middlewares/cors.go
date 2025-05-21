package middlewares

import (
	"github.com/gin-contrib/cors"
	"sync"
	"time"
)

var (
	configCors cors.Config
	once       sync.Once
)

func ConfigCors() (cors.Config, error) {
	var err error
	once.Do(func() {
		configCors, err = initializeCors()
	})
	return configCors, err
}

func initializeCors() (cors.Config, error) {
	configCors.AllowOrigins = []string{"*"}
	configCors.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	configCors.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	configCors.ExposeHeaders = []string{"Content-TokenType"}
	configCors.AllowCredentials = true
	configCors.MaxAge = 12 * time.Hour
	return configCors, nil
}
