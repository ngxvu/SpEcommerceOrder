package main

import (
	"context"
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
	"log"
	"order/internal/bootstrap"
	"order/internal/http/routes"
	"order/pkg/core/logger"
	"order/pkg/http/middlewares"
	"order/pkg/http/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger.Init(utils.APPNAME)

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

	go func() {
		// start HTTP server in background so main can continue to start gRPC
		if _, err := bootstrap.StartServer(router, app.Config); err != nil {
			log.Fatalf("failed to start http server: %v", err)
		}
	}()

	grpcSrv, err := bootstrap.StartGRPC(app)
	if err != nil {
		log.Fatalf("failed to start grpc: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down...")

	// Stop gRPC gracefully (safe nil-check)
	if grpcSrv != nil {
		grpcSrv.Stop()
	}

	// example: shutdown other servers with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = shutdownCtx
}
