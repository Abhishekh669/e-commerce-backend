package routes

import (
	"e-commerce.com/internal/app"
	"e-commerce.com/internal/middleware"
	"github.com/gin-gonic/gin"
)

func ProductServiceRouter(router *gin.RouterGroup, appConfig *app.App) {
	productServiceRoute := router.Group("/product-service")

	productServiceRoute.POST("/create-product", middleware.UserTokenVerification(), appConfig.ProductHandler.CreateProduct)
	productServiceRoute.GET("/get-seller-products", middleware.UserTokenVerification(), appConfig.ProductHandler.GetSellerProducts)
	productServiceRoute.PUT("/update-product/:productId", middleware.UserTokenVerification(), appConfig.ProductHandler.UpdateProduct)
	productServiceRoute.DELETE("/delete-product/:productId", middleware.UserTokenVerification(), appConfig.ProductHandler.DeleteProduct)
	productServiceRoute.GET("/get-product-by-id/:productId", appConfig.ProductHandler.GetProductById)
	productServiceRoute.GET("/get-all-products", appConfig.ProductHandler.GetAllProducts)
}
