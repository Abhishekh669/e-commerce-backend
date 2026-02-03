package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"e-commerce.com/internal/config"
	"e-commerce.com/internal/middleware"
	"e-commerce.com/internal/models"
	"e-commerce.com/internal/repository"
	"e-commerce.com/internal/utils"
	"github.com/gin-gonic/gin"
)

// CartItem represents an item in the cart with seller information
type CartItem struct {
	ID       string `json:"id"`
	SellerID string `json:"sellerId"`
	Quantity int64  `json:"quantity"`
	Price    int64  `json:"price"`
	Name     string `json:"name"`
}

type PaymentData struct {
	Amount                string `json:"amount"`
	TaxAmount             string `json:"tax_amount"`
	ProductServiceCharge  string `json:"product_service_charge"`
	ProductDeliveryCharge string `json:"product_delivery_charge"`
	TotalAmount           string `json:"total_amount"`
	TransactionUUID       string `json:"transaction_uuid"`
	ProductCode           string `json:"product_code"`
	SuccessURL            string `json:"success_url"`
	FailureURL            string `json:"failure_url"`
	SignedFieldNames      string `json:"signed_field_names"`
}

type SignatureData struct {
	TotalAmount     string
	TransactionUUID string
	ProductCode     string
	SecretKey       string
}

type PaymentService interface {
	InitiatePayment(ctx context.Context, cartItems []CartItem) (string, error)
	CheckPaymentStatus(ctx context.Context, transactionUUID, productCode, totalAmount string) (*PaymentStatusResponse, error)
	CreateOrderFromPayment(ctx context.Context, payment *models.Payment) (*models.Order, error)
	ProcessSuccessfulPayment(ctx context.Context, transactionUUID string) (*models.Order, error)
}

type paymentService struct {
	repo        repository.PaymentRepo
	orderRepo   repository.OrderRepo
	productRepo repository.ProductRepo
}

func NewPaymentService(repo repository.PaymentRepo) PaymentService {
	if repo == nil {
		repo = repository.NewPaymentRepository()
	}
	orderRepo := repository.NewOrderRepository()
	productRepo := repository.NewProductRepository()
	return &paymentService{
		repo:        repo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *paymentService) InitiatePayment(ctx context.Context, cartItems []CartItem) (string, error) {
	// Extract product IDs for availability check
	productIds := make([]string, 0, len(cartItems))
	for _, item := range cartItems {
		productIds = append(productIds, item.ID)
	}

	_, available, err := s.repo.CheckProductAvailability(ctx, productIds)
	if err != nil {
		return "", err
	}
	if !available {
		return "", errors.New("some products are not available")
	}

	var amount int64
	for _, item := range cartItems {
		// Calculate total amount for this item (price * quantity)
		itemTotal := item.Price * item.Quantity
		amount += itemTotal
	}

	// Calculate tax and charges (you can modify these based on your business logic)
	taxAmount := int64(0)
	serviceCharge := int64(0)
	deliveryCharge := int64(0)
	totalAmount := amount + taxAmount + serviceCharge + deliveryCharge

	// Ensure amount is at least 1 (eSewa requirement)
	if totalAmount < 1 {
		totalAmount = 1
	}

	paymentData := PaymentData{
		Amount:                strconv.FormatInt(amount, 10),
		TaxAmount:             strconv.FormatInt(taxAmount, 10),
		ProductServiceCharge:  strconv.FormatInt(serviceCharge, 10),
		ProductDeliveryCharge: strconv.FormatInt(deliveryCharge, 10),
		TotalAmount:           strconv.FormatInt(totalAmount, 10),
		TransactionUUID:       utils.GenerateEsewaTransactionUUID(),
		ProductCode:           config.AppConfig.EsewaMerchantCode,
		SuccessURL:            config.AppConfig.ESewaSuccessURL,
		FailureURL:            config.AppConfig.ESewaFailedURL,
		SignedFieldNames:      "total_amount,transaction_uuid,product_code",
	}

	// Generate signature - order matters: total_amount,transaction_uuid,product_code
	signature := utils.GenerateEsewaSignature(paymentData.TotalAmount, paymentData.TransactionUUID, paymentData.ProductCode, config.AppConfig.EsewaSecretKey)
	if signature == "" {
		return "", errors.New("failed to generate signature")
	}

	// Debug: Print payment data for troubleshooting (remove in production)
	fmt.Printf("DEBUG: Payment Data:\n")
	fmt.Printf("  Amount: %s\n", paymentData.Amount)
	fmt.Printf("  Tax Amount: %s\n", paymentData.TaxAmount)
	fmt.Printf("  Service Charge: %s\n", paymentData.ProductServiceCharge)
	fmt.Printf("  Delivery Charge: %s\n", paymentData.ProductDeliveryCharge)
	fmt.Printf("  Total Amount: %s\n", paymentData.TotalAmount)
	fmt.Printf("  Transaction UUID: %s\n", paymentData.TransactionUUID)
	fmt.Printf("  Product Code: %s\n", paymentData.ProductCode)
	fmt.Printf("  Success URL: %s\n", paymentData.SuccessURL)
	fmt.Printf("  Failure URL: %s\n", paymentData.FailureURL)
	fmt.Printf("  Signed Field Names: %s\n", paymentData.SignedFieldNames)
	fmt.Printf("  Signature: %s\n", signature)

	// Create payment record first
	userId, _, _, ok := middleware.GetUserFromContext(ctx.(*gin.Context))
	if !ok {
		return "", errors.New("user not found")
	}

	paymentRecord := models.Payment{
		ID:              utils.GenerateRandomUUID(),
		Amount:          totalAmount,
		UserId:          userId,
		TransactionUuid: paymentData.TransactionUUID,
		ProductIDs:      productIds,
		Status:          models.PaymentStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = s.repo.CreatePayment(ctx, &paymentRecord)
	if err != nil {
		return "", err
	}

	// For eSewa, we need to submit a form to their endpoint
	// The response will be a redirect, so we need to handle it properly
	// Order matters for eSewa - follow their exact specification
	form := url.Values{}

	// Add parameters in the exact order specified by eSewa
	form.Set("amount", paymentData.Amount)
	form.Set("tax_amount", paymentData.TaxAmount)
	form.Set("product_service_charge", paymentData.ProductServiceCharge)
	form.Set("product_delivery_charge", paymentData.ProductDeliveryCharge)
	form.Set("total_amount", paymentData.TotalAmount)
	form.Set("transaction_uuid", paymentData.TransactionUUID)
	form.Set("product_code", paymentData.ProductCode)
	form.Set("success_url", paymentData.SuccessURL)
	form.Set("failure_url", paymentData.FailureURL)
	form.Set("signed_field_names", paymentData.SignedFieldNames)
	form.Set("signature", signature)

	// Debug: Print the form data being sent (remove in production)
	fmt.Printf("DEBUG: Form data being sent:\n")
	for key, values := range form {
		fmt.Printf("  %s: %v\n", key, values)
	}

	// Make POST request to eSewa
	response, err := http.Post(config.AppConfig.ESewaPaymentURL, "application/x-www-form-urlencoded", bytes.NewBufferString(form.Encode()))
	fmt.Printf("DEBUG: Response: %v\n", response)
	fmt.Printf("DEBUG: Error: %v\n", err)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Read response body to check for any error messages
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if response contains error
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("eSewa returned status %d: %s", response.StatusCode, string(body))
	}

	// For eSewa, the response is typically a redirect or HTML form
	// We need to extract the payment URL from the response
	// If it's a redirect, get the Location header
	if location := response.Header.Get("Location"); location != "" {
		return location, nil
	}

	// If no redirect, check if response contains a payment URL
	// This is a fallback - eSewa usually redirects
	if response.Request != nil && response.Request.URL != nil {
		return response.Request.URL.String(), nil
	}

	// If we can't get a direct URL, return the eSewa payment URL with parameters
	// The frontend can use this to submit the form
	paymentURL := fmt.Sprintf("%s?amount=%s&tax_amount=%s&product_service_charge=%s&product_delivery_charge=%s&total_amount=%s&transaction_uuid=%s&product_code=%s&success_url=%s&failure_url=%s&signed_field_names=%s&signature=%s",
		config.AppConfig.ESewaPaymentURL,
		paymentData.Amount,
		paymentData.TaxAmount,
		paymentData.ProductServiceCharge,
		paymentData.ProductDeliveryCharge,
		paymentData.TotalAmount,
		paymentData.TransactionUUID,
		paymentData.ProductCode,
		paymentData.SuccessURL,
		paymentData.FailureURL,
		paymentData.SignedFieldNames,
		signature,
	)

	return paymentURL, nil
}

// CheckPaymentStatus checks the status of a payment using eSewa's status check API
func (s *paymentService) CheckPaymentStatus(ctx context.Context, transactionUUID, productCode, totalAmount string) (*PaymentStatusResponse, error) {
	// Build status check URL
	statusURL := fmt.Sprintf("%s?product_code=%s&total_amount=%s&transaction_uuid=%s",
		config.AppConfig.EsewaPaymentStatusCheckURL,
		productCode,
		totalAmount,
		transactionUUID,
	)

	// Make GET request to status check API
	response, err := http.Get(statusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check payment status: %v", err)
	}
	defer response.Body.Close()

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse JSON response
	var statusResponse PaymentStatusResponse
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %v", err)
	}

	return &statusResponse, nil
}

// CreateOrderFromPayment creates an order from a successful payment
func (s *paymentService) CreateOrderFromPayment(ctx context.Context, payment *models.Payment) (*models.Order, error) {
	// Validate payment data
	if payment.ProductIDs == nil || len(payment.ProductIDs) == 0 {
		return nil, fmt.Errorf("payment has no product IDs")
	}

	// Get product details for each product ID
	orderItems := make([]models.ProductItem, 0, len(payment.ProductIDs))
	for _, productID := range payment.ProductIDs {
		product, _, err := s.productRepo.GetProductByID(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product details: %v", err)
		}

		orderItem := models.ProductItem{
			ProductID: product.ID,
			SellerID:  product.SellerID, // Include seller ID in order item
			Quantity:  1,                // Default quantity, you can modify this based on your cart logic
			Price:     int64(product.Price),
		}
		orderItems = append(orderItems, orderItem)
	}

	// Create the order
	order := &models.Order{
		ID:            utils.GenerateRandomUUID(),
		User:          payment.UserId,
		Amount:        payment.Amount,
		Products:      orderItems,
		TransactionID: payment.TransactionUuid,
		Status:        models.OrderStatusCreated,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save the order to database
	err := s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	// Decrease product stock for each product in the order
	err = s.decreaseProductStock(ctx, orderItems)
	if err != nil {
		fmt.Printf("WARNING: Failed to decrease product stock: %v\n", err)
		// Don't fail the order creation if stock update fails
	} else {
		fmt.Printf("DEBUG: Successfully decreased product stock for order\n")
	}

	fmt.Printf("DEBUG: Order created successfully with ID: %s\n", order.ID)
	return order, nil
}

// ProcessSuccessfulPayment handles successful payment by updating payment status and creating order
func (s *paymentService) ProcessSuccessfulPayment(ctx context.Context, transactionUUID string) (*models.Order, error) {
	// Get the payment record
	payment, err := s.repo.GetPaymentByTransactionUUID(ctx, transactionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %v", err)
	}

	// Update payment status to success
	err = s.repo.UpdatePaymentStatus(ctx, payment.ID, models.PaymentStatusSuccess)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment status: %v", err)
	}

	// Create order from successful payment
	order, err := s.CreateOrderFromPayment(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	// Clear the user's cart after successful order creation
	err = s.repo.ClearUserCart(ctx, payment.UserId)
	if err != nil {
		// Log the error but don't fail the order creation
		fmt.Printf("WARNING: Failed to clear user cart for user %s: %v\n", payment.UserId, err)
	} else {
		fmt.Printf("DEBUG: Successfully cleared cart for user %s\n", payment.UserId)
	}

	fmt.Printf("DEBUG: Payment processed successfully and order created: %s\n", order.ID)
	return order, nil
}

// PaymentStatusResponse represents the response from eSewa status check API
type PaymentStatusResponse struct {
	ProductCode     string  `json:"product_code"`
	TransactionUUID string  `json:"transaction_uuid"`
	TotalAmount     float64 `json:"total_amount"`
	Status          string  `json:"status"`
	RefID           *string `json:"ref_id"`
}

// restoreProductStock restores product stock when order is cancelled
func (s *paymentService) restoreProductStock(ctx context.Context, products []models.ProductItem) error {
	for _, item := range products {
		// Get current product stock
		product, _, err := s.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			fmt.Printf("WARNING: Failed to get product %s: %v\n", item.ProductID, err)
			continue
		}

		// Restore stock
		newStock := product.Stock + int(item.Quantity)
		err = s.productRepo.UpdateProductStock(ctx, product.SellerID, item.ProductID, newStock)
		if err != nil {
			fmt.Printf("WARNING: Failed to restore stock for product %s: %v\n", item.ProductID, err)
		}
	}

	return nil
}

// decreaseProductStock decreases product stock when order is created
func (s *paymentService) decreaseProductStock(ctx context.Context, orderItems []models.ProductItem) error {
	for _, item := range orderItems {
		// Get current product stock
		product, _, err := s.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			fmt.Printf("WARNING: Failed to get product %s: %v\n", item.ProductID, err)
			continue
		}

		// Check if enough stock is available
		if product.Stock < int(item.Quantity) {
			fmt.Printf("WARNING: Insufficient stock for product %s. Available: %d, Required: %d\n",
				item.ProductID, product.Stock, item.Quantity)
			continue
		}

		// Decrease stock
		newStock := product.Stock - int(item.Quantity)
		err = s.productRepo.UpdateProductStock(ctx, product.SellerID, item.ProductID, newStock)
		if err != nil {
			fmt.Printf("WARNING: Failed to decrease stock for product %s: %v\n", item.ProductID, err)
		} else {
			fmt.Printf("DEBUG: Decreased stock for product %s from %d to %d\n",
				item.ProductID, product.Stock, newStock)
		}
	}

	return nil
}
