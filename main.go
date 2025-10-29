package main

import (
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
	"order/internal/bootstrap"
	"order/internal/http/routes"
	"order/pkg/core/logger"
	"order/pkg/http/middlewares"
	"order/pkg/http/utils"
)

func main() {
	logger.Init(utils.APPNAME)

	// Initialize application
	app, err := bootstrap.InitializeApp()
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize application")
		return
	}

	// Initialize Kafka
	//kafkaApp := bootstrap.InitializeKafka()
	//defer kafkaApp.Producer.Writer.Close()

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
	bootstrap.StartServer(router, app.Config)
	bootstrap.StartGRPC(app)
}
