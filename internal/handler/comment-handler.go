package handler

import (
	"fmt"
	"net/http"

	"e-commerce.com/internal/models"
	"e-commerce.com/internal/service"
	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	service service.CommentService
}

func (h *CommentHandler) CreateNewComment(c *gin.Context) {
	var userData models.ProductReviewFromClient

	if err := c.ShouldBindJSON(&userData); err != nil {
		fmt.Println("this is the error in data binding: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.service.CreateNewComment(c.Request.Context(), &userData)
	if err != nil {
		fmt.Println("thisis error in hanlder : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "comment created successfully", "success": true})
}

func NewCommentHandler(service service.CommentService) *CommentHandler {
	return &CommentHandler{
		service: service,
	}
}
