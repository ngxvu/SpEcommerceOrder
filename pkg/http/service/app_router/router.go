package app_router

import (
	"basesource/internal/route"
	"basesource/internal/utils/app_errors"
	"basesource/pkg/http/common"
	"basesource/pkg/http/middlewares"
	"basesource/pkg/http/service/app_config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(router *gin.Engine, configCors cors.Config, app *app_config.App) {
	router.Use(cors.New(configCors))
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.RequestLogger(common.APPNAME))
	router.Use(app_errors.ErrorHandler)
	router.Use(static.Serve("/image-storage/", static.LocalFile("./image-storage", true)))

	route.ApplicationV1Router(
		app.PGRepo,
		router,
	)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
