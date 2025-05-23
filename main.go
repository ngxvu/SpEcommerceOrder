package main

import (
	"fmt"
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"kimistore/conf"
	repo "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/route"
	"kimistore/internal/utils/app_errors"
	"kimistore/pkg/http/db"
	"kimistore/pkg/http/logger"
	"kimistore/pkg/http/middlewares"
	"net/http"
	"strings"
	"time"
)

const (
	APPNAME = "kimistore"
)

type App struct {
	Config *viper.Viper
	PGRepo repo.PGInterface
}

// initializeApp initializes all application dependencies
func initializeApp() (*App, error) {
	config := conf.GetConfig()

	// Initialize database
	dbBackend, err := db.InitDatabase(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	pgRepo := repo.NewPGRepo(dbBackend)

	return &App{
		PGRepo: pgRepo,
		Config: config,
	}, nil
}

func main() {
	logger.Init(APPNAME)

	// Initialize application
	app, err := initializeApp()
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize application")
		return
	}

	setupLogger()

	// Setup and start server
	router := gin.Default()
	router.Use(limit.MaxAllowed(200))

	configCors, err := middlewares.ConfigCors()
	if err != nil {
		logger.LogError(logger.WithTag("Backend|Main"), err, "failed to initialize cors")
		return
	}

	setupRouter(router, configCors, app)
	startServer(router, app.Config)
}

func setupRouter(router *gin.Engine, configCors cors.Config, app *App) {
	router.Use(cors.New(configCors))
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.RequestLogger(APPNAME))
	router.Use(app_errors.ErrorHandler)
	router.Use(static.Serve("/image-storage/", static.LocalFile("./image-storage", true)))

	route.ApplicationV1Router(
		app.PGRepo,
		router,
		app.Config,
	)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func startServer(router http.Handler, config *viper.Viper) {

	serverPort := fmt.Sprintf(":%s", config.GetString("ServerPort"))

	s := &http.Server{
		Addr:           serverPort,
		Handler:        router,
		ReadTimeout:    18000 * time.Second,
		WriteTimeout:   18000 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		_ = fmt.Errorf("fatal error description: %s", strings.ToLower(err.Error()))
		panic(err)
	}
}

func setupLogger() {
	logger.DefaultLogger.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		FullTimestamp:    true,
		PadLevelText:     true,
		ForceQuote:       true,
		QuoteEmptyFields: true,
	})
}
