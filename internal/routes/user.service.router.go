package routes

import (
	"net/http"

	"e-commerce.com/internal/app"
	"e-commerce.com/internal/middleware"
	"e-commerce.com/internal/models"
	"github.com/gin-gonic/gin"
)

func UserServiceRouter(router *gin.RouterGroup, appConfig *app.App) {
	userServiceRoute := router.Group("/user-service")
	userServiceRoute.GET("/user-token-verification", middleware.UserTokenVerification(), func(c *gin.Context) {
		userEmail, ok := c.Get("userEmail")
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "user email not found in context",
			})
			return
		}

		user, err := appConfig.UserRepo.GetUserByEmail(userEmail.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "user not found",
			})
			return
		}

		userData := models.SafeUser{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Role:       user.Role,
			IsVerified: user.IsVerified,
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "user token verified",
			"user":    userData,
		})
	})
	userServiceRoute.GET("/verify-user", appConfig.UserHandler.VerifyUserTokenHandler)
	userServiceRoute.GET("/verify-user-token", appConfig.UserHandler.GetUserFromToken)
	userServiceRoute.GET("/delete-user-session", appConfig.UserHandler.DeleteUserSessionHandler)

	userServiceRoute.POST("/create-user", appConfig.UserHandler.RegisterUserHandler)
	userServiceRoute.POST("/login", appConfig.UserHandler.LoginUserHandler)
}
