package main

import (
	"basesource/internal/bootstrap"
	"basesource/internal/http/routes"
	"basesource/pkg/core/configloader"
	"basesource/pkg/core/logger"
	"basesource/pkg/http/common"
	"basesource/pkg/http/middlewares"
	"fmt"
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	logger.Init(common.APPNAME)

	// Initialize application
	app, err := bootstrap.InitializeApp()
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

	routes.NewHTTPServer(router, configCors, app)
	startServer(router, app.Config)
}

func startServer(router http.Handler, config *configloader.Config) {

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
