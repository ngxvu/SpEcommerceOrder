package route

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"kimistore/internal/handlers"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"kimistore/internal/services"
	"kimistore/pkg/http/middlewares"
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
		AuthUserRoutes(routerV1, handlers.NewAuthUserHandler(newPgRepo))

		// Media
		mediaRepo := repo.NewMediaRepository()
		mediaService := services.NewMediaService(mediaRepo, newPgRepo)
		MediaRoutes(routerV1, handlers.NewMediaHandler(newPgRepo, mediaService))

		// Product
		productRepo := repo.NewProductRepository()
		productService := services.NewProductService(productRepo, newPgRepo)
		ProductRoutes(routerV1, handlers.NewProductHandler(newPgRepo, productService))

		// Post
		postRepo := repo.NewPostRepository()
		postService := services.NewPostService(postRepo, newPgRepo)
		PostRoutes(routerV1, handlers.NewPostHandler(newPgRepo, postService))

		// Testimonial
		testimonialRepo := repo.NewTestimonialRepository()
		testimonialService := services.NewTestimonialService(testimonialRepo, newPgRepo)
		TestimonialRoutes(routerV1, handlers.NewTestimonialHandler(newPgRepo, testimonialService))

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

func ProductRoutes(router *gin.RouterGroup, handler *handlers.ProductHandler) {
	routerProduct := router.Group("/product", middlewares.AuthMiddleware())
	{
		routerProduct.POST("/create", handler.CreateProduct)
		routerProduct.POST("/detail/:id", handler.GetDetailProduct)
		routerProduct.POST("/list", handler.GetListProduct)
		routerProduct.PUT("/update/:id", handler.UpdateProduct)
		routerProduct.DELETE("/delete/:id", handler.DeleteProduct)
	}
}

func PostRoutes(router *gin.RouterGroup, handler *handlers.PostHandler) {
	routerPost := router.Group("/post", middlewares.AuthMiddleware())
	{
		routerPost.POST("/create", handler.CreatePost)
		routerPost.POST("/detail/:id", handler.GetDetailPost)
		routerPost.POST("/list", handler.GetListPost)
		routerPost.PUT("/update/:id", handler.UpdatePost)
		routerPost.DELETE("/delete/:id", handler.DeletePost)
	}
}

func TestimonialRoutes(router *gin.RouterGroup, handler *handlers.TestimonialHandler) {
	routerTestimonial := router.Group("/testimonial", middlewares.AuthMiddleware())
	{
		routerTestimonial.POST("/create", handler.CreateTestimonial)
		routerTestimonial.POST("/detail/:id", handler.GetDetailTestimonial)
		routerTestimonial.POST("/list", handler.GetListTestimonial)
		routerTestimonial.PUT("/update/:id", handler.UpdateTestimonial)
		routerTestimonial.DELETE("/delete/:id", handler.DeleteTestimonial)
	}
}
