package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"e-commerce.com/internal/service"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service service.PaymentService
}

// CartItem represents an item in the cart with seller information
type CartItem struct {
	ID       string `json:"id"`
	SellerID string `json:"sellerId"`
	Quantity int64  `json:"quantity"`
	Price    int64  `json:"price"`
	Name     string `json:"name"`
}

type CartItemsRequest struct {
	CartItems []CartItem `json:"cartItems"`
}

type ProcessSuccessfulPaymentRequest struct {
	TransactionUUID string `json:"transaction_uuid"`
}

func (h *PaymentHandler) InitiatePayment(c *gin.Context) {
	var cartItemsReq CartItemsRequest
	if err := c.ShouldBindJSON(&cartItemsReq); err != nil {
		fmt.Println("error in data : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	// Convert to service CartItem type
	cartItems := make([]service.CartItem, len(cartItemsReq.CartItems))
	for i, item := range cartItemsReq.CartItems {
		cartItems[i] = service.CartItem{
			ID:       item.ID,
			SellerID: item.SellerID,
			Quantity: item.Quantity,
			Price:    item.Price,
			Name:     item.Name,
		}
	}

	paymentUrl, err := h.service.InitiatePayment(c, cartItems)
	if err != nil {
		fmt.Println("failed to intitate paym;ent : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": paymentUrl, "success": true})
}

func (h *PaymentHandler) CheckPaymentStatus(c *gin.Context) {
	// Get query parameters
	transactionUUID := c.Query("transaction_uuid")
	productCode := c.Query("product_code")
	totalAmountStr := c.Query("total_amount")

	// Validate required parameters
	if transactionUUID == "" || productCode == "" || totalAmountStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "transaction_uuid, product_code, and total_amount are required",
			"success": false,
		})
		return
	}

	// Parse total amount
	totalAmount, err := strconv.ParseFloat(totalAmountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid total_amount format",
			"success": false,
		})
		return
	}

	// Check payment status
	statusResponse, err := h.service.CheckPaymentStatus(c, transactionUUID, productCode, strconv.FormatFloat(totalAmount, 'f', -1, 64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    statusResponse,
	})
}

func (h *PaymentHandler) ProcessSuccessfulPayment(c *gin.Context) {
	var req ProcessSuccessfulPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	if req.TransactionUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "transaction_uuid is required",
			"success": false,
		})
		return
	}

	// Process successful payment and create order
	order, err := h.service.ProcessSuccessfulPayment(c, req.TransactionUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"order":   order,
		"message": "Order created successfully",
	})
}

func NewPaymentHandler(service service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}
