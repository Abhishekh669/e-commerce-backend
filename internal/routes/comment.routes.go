package routes

import (
	"e-commerce.com/internal/app"
	"e-commerce.com/internal/middleware"
	"github.com/gin-gonic/gin"
)

func CommentROuter(router *gin.RouterGroup, appConfig *app.App) {
	commentRoute := router.Group("/comment-service")
	commentRoute.POST("/create-comment", middleware.UserTokenVerification(), appConfig.CommentHandler.CreateNewComment)

}
