package main

import (
	"fmt"
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
	"kimistore/conf"
	"kimistore/pkg/http/common"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/middlewares"
	"kimistore/pkg/http/service/app_config"
	"kimistore/pkg/http/service/app_router"
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
	if err := s.ListenAndServe(); err != nil {
		_ = fmt.Errorf("failed to start server on port %s: %w", serverPort, err)
		panic(err)
	}
}
