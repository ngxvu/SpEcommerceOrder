package main

import (
	"context"
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
	"log"
	"order/internal/bootstrap"
	"order/internal/http/routes"
	"order/internal/telemetry"
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
	logger.SetupLogger()

	// Initialize application
	app, err := bootstrap.InitializeAppConfiguration()
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize application")
		return
	}

	// Initialize tracer early and keep cleanup for shutdown
	ctx := context.Background()
	cleanup, err := telemetry.InializerTracer(ctx, utils.APPNAME, app.AppConfig)
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize tracer")
		return
	}
	// ensure tracer provider shutdown on process exit
	defer func() { _ = cleanup(context.Background()) }()

	// Setup and start server
	router := gin.Default()
	router.Use(limit.MaxAllowed(200))

	configCors, err := middlewares.ConfigCors()
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize cors")
		return
	}

	routes.NewHTTPServerSetup(router, configCors, app)

	httpSrv, httpErrCh := bootstrap.StartServer(router, app.AppConfig)
	go func() {
		if err := <-httpErrCh; err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Start gRPC server
	grpcSrv, stopKafka, err := bootstrap.StartGRPC(app)
	if err != nil {
		log.Fatalf("failed to start grpc: %v", err)
	}
	defer stopKafka()

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-sigCtx.Done()

	log.Println("shutting down...")

	// Stop gRPC gracefully (safe nil-check)
	if grpcSrv != nil {
		grpcSrv.Stop()
	}

	// Shutdown HTTP server with timeout
	if httpSrv != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown error: %v", err)
		}
	}
}
