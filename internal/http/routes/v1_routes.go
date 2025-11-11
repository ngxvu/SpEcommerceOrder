package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"order_service/internal/bootstrap"
	"order_service/internal/http/server"
	"order_service/pkg/http/middlewares"
	"order_service/pkg/http/utils"
	"order_service/pkg/http/utils/app_errors"
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
