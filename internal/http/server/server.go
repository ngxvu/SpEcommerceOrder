package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	handlers2 "order_service/internal/http/handlers"
	pgGorm "order_service/internal/repositories/pg-gorm"
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

		//orderRepo := repo.NewOrderRepository(newPgRepo)
		//orderService := services.NewOrderService(orderRepo, newPgRepo)
		//OrderRoutes(routerV1, handlers2.NewOrderHandler(newPgRepo, orderService))
	}
}

func MigrateRoutes(router *gin.RouterGroup, handler *handlers2.MigrationHandler) {
	routerAuth := router.Group("/internal")
	{
		routerAuth.POST("/migrate", handler.Migrate)
	}
}

func OrderRoutes(router *gin.RouterGroup, handler *handlers2.OrderHandler) {
	routerOrder := router.Group("/orders")
	{
		routerOrder.POST("/", handler.CreateOrder)
	}
}
