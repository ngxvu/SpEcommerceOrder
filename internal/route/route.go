package route

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"kimistore/internal/handlers"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/services"
)

func ApplicationV1Router(
	newPgRepo pgGorm.PGInterface,
	router *gin.Engine,
	config *viper.Viper,
) {
	routerV1 := router.Group("/v1")
	{
		// Swagger
		routerV1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// Migrations
		MigrateRoutes(routerV1, handlers.NewMigrationHandler(newPgRepo))

		// Auth for User
		AuthUserRoutes(routerV1, handlers.NewAuthUserHandler(newPgRepo, config))

		// Media
		mediaRepo := repo.NewMediaRepository()
		mediaService := services.NewMediaService(mediaRepo, newPgRepo)
		MediaRoutes(routerV1, handlers.NewMediaHandler(newPgRepo, mediaService, config))

	}
}

func MigrateRoutes(router *gin.RouterGroup, handler *handlers.MigrationHandler) {
	routerAuth := router.Group("/internal")
	{
		routerAuth.POST("/migrate", handler.Migrate)
	}
}

func AuthUserRoutes(router *gin.RouterGroup, handler *handlers.AuthUserHandler) {
	routerAuth := router.Group("/auth")
	{
		routerAuth.POST("/login", handler.Login)
		routerAuth.POST("/register", handler.Register)
	}
}

func MediaRoutes(router *gin.RouterGroup, handler *handlers.MediaHandler) {
	routerAuth := router.Group("/media")
	{
		//routerAuth.POST("/upload-media", handler.UploadMedia)
		routerAuth.POST("/upload-images", handler.UploadListImage)
	}
}
