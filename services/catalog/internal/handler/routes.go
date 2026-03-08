package handler

import (
	"ecommerce/services/catalog/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ecommerce/services/catalog/docs"
)

func RegisterRoutes(
	router *gin.Engine,
	categoryHandler *CategoryHandler,
	productHandler *ProductHandler,
	sellerHandler *SellerHandler,
	variantHandler *VariantHandler,
	sellerService service.SellerService,
) {
	router.GET("/api/v1/catalog/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1 := router.Group("/api/v1/catalog")

	public := v1.Group("/")
	{

		public.GET("/categories", categoryHandler.GetAllCategories)
		public.GET("/categories/:id/breadcrumbs", categoryHandler.GetBreadcrumbs)

		public.GET("/products", productHandler.ListProducts)
		public.GET("/products/:id", productHandler.GetProductByID)

		public.GET("/variants/:sku", variantHandler.GetVariantByID)
		public.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}

	protected := v1.Group("/")
	protected.Use(RequireUser)
	{
		protected.POST("/sellers", sellerHandler.CreateSeller)
		protected.GET("/sellers/me", sellerHandler.GetMyProfile)
	}

	sellerRoutes := v1.Group("/seller")
	sellerRoutes.Use(RequireUser, RequireSeller(sellerService))
	{
		sellerRoutes.POST("/products", productHandler.CreateProduct)
		sellerRoutes.PUT("/products/:id", productHandler.UpdateProduct)
		sellerRoutes.DELETE("/products/:id", productHandler.DeleteProduct)

		sellerRoutes.POST("/products/:product_id/variants", variantHandler.CreateVariant)
		sellerRoutes.PUT("/variants/:id", variantHandler.UpdateVariant)
		sellerRoutes.DELETE("/variants/:id", variantHandler.DeleteVariant)
	}
}
