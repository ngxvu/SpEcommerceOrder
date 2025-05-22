package route

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"kimistore/internal/handlers"
	repo "kimistore/internal/repo/pg-gorm"
)

func ApplicationV1Router(
	newPgRepo repo.PGInterface,
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

	}
}

func MigrateRoutes(router *gin.RouterGroup, controller *handlers.MigrationHandler) {
	routerAuth := router.Group("/internal")
	{
		routerAuth.POST("/migrate", controller.Migrate)
	}
}

func AuthUserRoutes(router *gin.RouterGroup, controller *handlers.AuthUserHandler) {
	routerAuth := router.Group("/auth")
	{
		routerAuth.POST("/login", controller.Login)
		routerAuth.POST("/register", controller.Register)
	}
}
