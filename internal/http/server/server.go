package server

import (
	handlers2 "basesource/internal/http/handlers"
	"basesource/internal/repositories"
	pgGorm "basesource/internal/repositories/pg-gorm"
	"basesource/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func ApplicationV1Router(
	newPgRepo pgGorm.PGInterface,
	router *gin.Engine,
) {
	routerV1 := router.Group("/v1")
	{
		// Swagger
		routerV1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// Migrations
		MigrateRoutes(routerV1, handlers2.NewMigrationHandler(newPgRepo))

		// Auth for User
		authUserRepo := repositories.NewAuthUserRepository(newPgRepo)
		authUserService := services.NewAuthUserService(authUserRepo, newPgRepo)
		AuthorizationUserRoutes(routerV1, handlers2.NewAuthUserHandler(newPgRepo, authUserService))
	}
}

func MigrateRoutes(router *gin.RouterGroup, handler *handlers2.MigrationHandler) {
	routerAuth := router.Group("/internal")
	{
		routerAuth.POST("/migrate", handler.Migrate)
	}
}

func AuthorizationUserRoutes(router *gin.RouterGroup, handler *handlers2.AuthUserHandler) {
	routerAuth := router.Group("/auth")
	{
		routerAuth.POST("/login", handler.Login)
		routerAuth.POST("/register", handler.Register)
	}
}
