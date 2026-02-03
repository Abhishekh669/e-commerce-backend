package routes

import (
	"e-commerce.com/internal/app"
	"e-commerce.com/internal/middleware"
	"github.com/gin-gonic/gin"
)

func PaymentServiceRouter(router *gin.RouterGroup, appConfig *app.App) {
	paymentServiceRoute := router.Group("/payment-service")

	paymentServiceRoute.POST("/initiate-payment", middleware.UserTokenVerification(), appConfig.PaymentHandler.InitiatePayment)
	paymentServiceRoute.GET("/check-status", middleware.UserTokenVerification(), appConfig.PaymentHandler.CheckPaymentStatus)
	paymentServiceRoute.POST("/process-successful-payment", middleware.UserTokenVerification(), appConfig.PaymentHandler.ProcessSuccessfulPayment)
}
