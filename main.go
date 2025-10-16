package main

import (
	"basesource/conf"
	"basesource/pkg/http/common"
	"basesource/pkg/http/logger"
	"basesource/pkg/http/middlewares"
	"basesource/pkg/http/service/app_config"
	"basesource/pkg/http/service/app_router"
	"fmt"
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	logger.Init(common.APPNAME)

	// Initialize application
	app, err := app_config.InitializeApp()
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize application")
		return
	}

	logger.SetupLogger()

	// Setup and start server
	router := gin.Default()
	router.Use(limit.MaxAllowed(200))

	configCors, err := middlewares.ConfigCors()
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize cors")
		return
	}

	app_router.SetupRouter(router, configCors, app)
	startServer(router, app.Config)
}

func startServer(router http.Handler, config *conf.Config) {

	serverPort := fmt.Sprintf(":%s", config.ServerPort)
	s := &http.Server{
		Addr:    serverPort,
		Handler: router,
	}
	log.Println("Server started on port", serverPort)
	if err := s.ListenAndServe(); err != nil {
		_ = fmt.Errorf("failed to start server on port %s: %w", serverPort, err)
		panic(err)
	}
}
