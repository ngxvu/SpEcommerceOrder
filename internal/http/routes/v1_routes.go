package routes

import (
	"basesource/internal/bootstrap"
	"basesource/internal/http/server"
	"basesource/pkg/http/common"
	"basesource/pkg/http/middlewares"
	"basesource/pkg/http/utils/app_errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHTTPServer(router *gin.Engine, configCors cors.Config, app *bootstrap.App) {
	router.Use(cors.New(configCors))
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.RequestLogger(common.APPNAME))
	router.Use(app_errors.ErrorHandler)

	server.ApplicationV1Router(
		app.PGRepo,
		router,
	)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
