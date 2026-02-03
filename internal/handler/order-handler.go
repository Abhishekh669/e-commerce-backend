package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"e-commerce.com/internal/middleware"
	"e-commerce.com/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service service.OrderService
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateProductStockRequest struct {
	Stock int `json:"stock" binding:"required,min=0"`
}

func (h *OrderHandler) FinishOrderHandler(c *gin.Context) {
	// Get seller ID from context
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	orderId := c.Param("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required", "success": false})
		return
	}

	// Accept order
	err := h.service.OrderFinished(c, sellerId, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order deleivered successfully",
	})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	// Get user ID from context
	userId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated", "success": false})
		return
	}

	orders, _, err := h.service.GetUserOrders(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
	})
}

func (h *OrderHandler) GetUserOrderDetails(c *gin.Context) {
	// Get user ID from context
	userId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated", "success": false})
		return
	}

	orderId := c.Param("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required", "success": false})
		return
	}

	// Get order details
	order, err := h.service.GetUserOrderDetails(c, userId, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

func (h *OrderHandler) CancelUserOrder(c *gin.Context) {
	// Get user ID from context
	userId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated", "success": false})
		return
	}

	orderId := c.Param("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required", "success": false})
		return
	}
	fmt.Println("this is order id in cancel user order handler ; ", orderId)
	// Cancel order
	err := h.service.CancelUserOrder(c, userId, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order cancelled successfully",
	})
}

func (h *OrderHandler) GetSellerOrders(c *gin.Context) {
	// Get seller ID from context (sellers are users with a specific role)
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	// Get status filter
	status := c.Query("status")

	// Get all seller orders
	orders, total, err := h.service.GetSellerOrders(c, sellerId, 1, 0, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders": orders,
			"pagination": gin.H{
				"page":  1,
				"limit": total,
				"total": total,
			},
		},
	})
}

func (h *OrderHandler) GetSellerOrderDetails(c *gin.Context) {
	// Get seller ID from context (sellers are users with a specific role)
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	orderId := c.Param("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required", "success": false})
		return
	}

	// Get order details for seller
	order, err := h.service.GetSellerOrderDetails(c, sellerId, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	// Get seller ID from context (sellers are users with a specific role)
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	orderId := c.Param("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required", "success": false})
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	// Update order status
	err := h.service.UpdateOrderStatus(c, sellerId, orderId, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order status updated successfully",
	})
}

func (h *OrderHandler) GetSellerProducts(c *gin.Context) {
	// Get seller ID from context (sellers are users with a specific role)
	fmt.Println("iamhereforsellerproducts")
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	fmt.Println("this is seeler id : ", sellerId)
	if !ok {
		fmt.Println("error in geting data")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	// Get query parameters for pagination and filtering
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get seller products
	products, total, err := h.service.GetSellerProducts(c, sellerId, page, limit)
	if err != nil {
		fmt.Println("this ishte products list : ", products, "or error man : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"products": products,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		},
	})
}

func (h *OrderHandler) UpdateProductStock(c *gin.Context) {
	// Get seller ID from context (sellers are users with a specific role)
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	productId := c.Param("productId")
	if productId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required", "success": false})
		return
	}

	var req UpdateProductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	// Update product stock
	err := h.service.UpdateProductStock(c, sellerId, productId, req.Stock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product stock updated successfully",
	})
}

func (h *OrderHandler) GetSellerOrdersWithDetails(c *gin.Context) {
	// Get seller ID from context
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	// Get all seller orders with details
	orders, total, err := h.service.GetSellerOrdersWithDetails(c, sellerId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders": orders,
			"pagination": gin.H{
				"page":  1,
				"limit": total,
				"total": total,
			},
		},
	})
}

func (h *OrderHandler) AcceptOrder(c *gin.Context) {
	// Get seller ID from context
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	orderId := c.Param("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required", "success": false})
		return
	}

	// Accept order
	err := h.service.AcceptOrder(c, sellerId, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order accepted successfully",
	})
}

func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	// Get seller ID from context
	sellerId, _, _, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not authenticated", "success": false})
		return
	}

	orderId := c.Param("orderId")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required", "success": false})
		return
	}

	// Delete order
	err := h.service.DeleteOrder(c, sellerId, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order deleted successfully",
	})
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}
