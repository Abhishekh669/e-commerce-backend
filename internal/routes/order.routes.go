package routes

import (
	"e-commerce.com/internal/app"
	"e-commerce.com/internal/middleware"
	"github.com/gin-gonic/gin"
)

func OrderRouter(router *gin.RouterGroup, appConfig *app.App) {
	orderRoute := router.Group("/orders")

	// User order routes
	orderRoute.GET("/user", middleware.UserTokenVerification(), appConfig.OrderHandler.GetUserOrders)
	orderRoute.GET("/user/:orderId", middleware.UserTokenVerification(), appConfig.OrderHandler.GetUserOrderDetails)
	orderRoute.PUT("/user/:orderId/cancel", middleware.UserTokenVerification(), appConfig.OrderHandler.CancelUserOrder)

	// Seller order management routes
	orderRoute.GET("/seller", middleware.UserTokenVerification(), appConfig.OrderHandler.GetSellerOrders)
	orderRoute.GET("/seller/details", middleware.UserTokenVerification(), appConfig.OrderHandler.GetSellerOrdersWithDetails)
	orderRoute.GET("/seller/:orderId", middleware.UserTokenVerification(), appConfig.OrderHandler.GetSellerOrderDetails)
	orderRoute.PUT("/seller/:orderId/status", middleware.UserTokenVerification(), appConfig.OrderHandler.UpdateOrderStatus)
	orderRoute.PUT("/seller/:orderId/accept", middleware.UserTokenVerification(), appConfig.OrderHandler.AcceptOrder)
	orderRoute.PUT("/seller/:orderId/delivered", middleware.UserTokenVerification(), appConfig.OrderHandler.FinishOrderHandler)
	orderRoute.DELETE("/seller/:orderId", middleware.UserTokenVerification(), appConfig.OrderHandler.DeleteOrder)
	orderRoute.GET("/seller/products", middleware.UserTokenVerification(), appConfig.OrderHandler.GetSellerProducts)
	orderRoute.PUT("/seller/products/:productId/stock", middleware.UserTokenVerification(), appConfig.OrderHandler.UpdateProductStock)
}
