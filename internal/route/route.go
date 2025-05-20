package route

import (
	handlers2 "emission/internal/handlers"
	"emission/internal/repo/pg-gorm"
	services2 "emission/internal/services"
	"emission/pkg/http/ginext"
	"emission/pkg/http/service"
)

type Service struct {
	*service.BaseApp
}

func NewService() *Service {

	s := &Service{
		service.NewApp("emission Service", "v1.0"),
	}

	pgGormClient := s.GetDB().Debug()
	repo := pg_gorm.NewRepo(pgGormClient)
	authService := services2.NewAuthService()
	factorService := services2.NewFactoryService(repo, authService)
	emissionService := services2.NewEmissionService(repo, natsPublishClient)
	factorHandler := handlers2.NewFactoryHandlers(factorService)
	emissionHandler := handlers2.NewEmissionHandlers(emissionService, authService)
	migrationHandler := handlers2.NewMigrationHandler(pgGormClient)

	// internal routes
	s.Router.POST("/internal/migrate", migrationHandler.Migrate)

	// public routes
	v1Api := s.Router.Group("/api/v1")
	v1Api.POST("/factories", ginext.WrapHandler(factorHandler.CreateFactory))
	v1Api.GET("/emissions", ginext.WrapHandler(emissionHandler.ListEmission))
	v1Api.POST("/emissions", ginext.WrapHandler(emissionHandler.CreateEmission))

	return s
}
