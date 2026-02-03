package routes

import (
	"e-commerce.com/internal/app"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(app *gin.Engine, appConfig *app.App) {
	apiGroup := app.Group("/api/v1")
	UserServiceRouter(apiGroup, appConfig)
	ProductServiceRouter(apiGroup, appConfig)
	PaymentServiceRouter(apiGroup, appConfig)
	OrderRouter(apiGroup, appConfig)
	CommentROuter(apiGroup, appConfig)
}
