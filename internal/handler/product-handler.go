package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"e-commerce.com/internal/middleware"
	"e-commerce.com/internal/models"
	"e-commerce.com/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	maxLimit      = 4
	defaultOffset = 0
)

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	userId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "success": false})
		return
	}
	var product models.CreateProductRequest
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	fmt.Println("this is product : ", product)
	err := h.service.CreateProduct(c, &product, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product created successfully", "success": true})
}

func (h *ProductHandler) GetSellerProducts(c *gin.Context) {
	userId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "success": false})
		return
	}
	page, limit := c.Query("page"), c.Query("limit")
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "100"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page", "success": false})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit", "success": false})
		return
	}
	products, total, hasMore, err := h.service.GetSellerProducts(c, userId, pageInt, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products, "success": true, "page": pageInt, "limit": limitInt, "total": total, "hasMore": hasMore})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {

	productId := c.Param("productId")
	if productId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required", "success": false})
		return
	}
	var product models.UpdateProductRequest
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	err := h.service.UpdateProduct(c, productId, &product)
	fmt.Println("this is product : ", product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully", "success": true})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productId := c.Param("productId")
	if productId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required", "success": false})
		return
	}
	err := h.service.DeleteProduct(c, productId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully", "success": true})
}

func (h *ProductHandler) GetProductById(c *gin.Context) {
	productId := c.Param("productId")
	if productId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required", "success": false})
		return
	}
	product, reviews, err := h.service.GetProductById(c, productId)
	if err != nil {
		fmt.Println("htisi s error man what ot do now ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	fmt.Println("htisis revies in handler: ", reviews)
	c.JSON(http.StatusOK, gin.H{"product": product, "reviews": reviews, "success": true})
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	search := c.Query("search")
	limit := c.DefaultQuery("limit", "4")
	offset := c.DefaultQuery("offset", "0")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 4
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		offsetInt = defaultOffset
	}

	if limitInt > 4 {
		limitInt = 4
	}

	fmt.Println("htis are ooffset and limit : ", offsetInt, limitInt)

	products, err := h.service.GetAllProducts(c, &search, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products, "success": true})
}
