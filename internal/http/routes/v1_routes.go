package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"order/internal/bootstrap"
	"order/internal/http/server"
	"order/pkg/http/middlewares"
	"order/pkg/http/utils"
	"order/pkg/http/utils/app_errors"
)

func NewHTTPServer(router *gin.Engine, configCors cors.Config, app *bootstrap.App) {
	router.Use(cors.New(configCors))
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.RequestLogger(utils.APPNAME))
	router.Use(app_errors.ErrorHandler)

	server.ApplicationV1Router(
		app.PGRepo,
		router,
	)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
