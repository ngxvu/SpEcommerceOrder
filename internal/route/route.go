package route

import (
	"basesource/internal/handlers"
	"basesource/internal/repo"
	pgGorm "basesource/internal/repo/pg-gorm"
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
		MigrateRoutes(routerV1, handlers.NewMigrationHandler(newPgRepo))

		// Auth for User
		authUserRepo := repo.NewAuthUserRepository(newPgRepo)
		authUserService := services.NewAuthUserService(authUserRepo, newPgRepo)
		AuthorizationUserRoutes(routerV1, handlers.NewAuthUserHandler(newPgRepo, authUserService))
	}
}

func MigrateRoutes(router *gin.RouterGroup, handler *handlers.MigrationHandler) {
	routerAuth := router.Group("/internal")
	{
		routerAuth.POST("/migrate", handler.Migrate)
	}
}

func AuthorizationUserRoutes(router *gin.RouterGroup, handler *handlers.AuthUserHandler) {
	routerAuth := router.Group("/auth")
	{
		routerAuth.POST("/login", handler.Login)
		routerAuth.POST("/register", handler.Register)
	}
}
