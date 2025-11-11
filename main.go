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

	httpSrv, httpErrCh := bootstrap.StartServer(router, app.Config)

	go func() {
		if err := <-httpErrCh; err != nil {
			log.Fatalf("http server error: %v", err)
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

	// Shutdown HTTP server with timeout
	if httpSrv != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown error: %v", err)
		}
	}
}
