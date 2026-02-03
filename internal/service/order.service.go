package service

import (
	"context"
	"fmt"

	"e-commerce.com/internal/models"
	"e-commerce.com/internal/repository"
)

type OrderService interface {
	OrderFinished(ctx context.Context, sellerId, orderId string) error
	GetUserOrders(ctx context.Context, userId string) ([]models.OrderWithProductDetails, int64, error)
	GetUserOrderDetails(ctx context.Context, userId, orderId string) (*models.Order, error)
	CancelUserOrder(ctx context.Context, userId, orderId string) error
	GetSellerOrders(ctx context.Context, sellerId string, page, limit int, status string) ([]*models.Order, int64, error)
	GetSellerOrderDetails(ctx context.Context, sellerId, orderId string) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, sellerId, orderId, status string) error
	GetSellerProducts(ctx context.Context, sellerId string, page, limit int) ([]*models.Product, int64, error)
	UpdateProductStock(ctx context.Context, sellerId, productId string, stock int) error
	GetSellerOrdersWithDetails(ctx context.Context, sellerId string) ([]models.OrderWithProductDetails, int64, error)
	AcceptOrder(ctx context.Context, sellerId, orderId string) error
	DeleteOrder(ctx context.Context, sellerId, orderId string) error
}

type orderService struct {
	orderRepo   repository.OrderRepo
	productRepo repository.ProductRepo
}

func NewOrderService(orderRepo repository.OrderRepo, productRepo repository.ProductRepo) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *orderService) OrderFinished(ctx context.Context, sellerId, orderId string) error {
	// Update order status to paid and processing
	err := s.orderRepo.UpdateOrderStatus(ctx, orderId, string(models.OrderStatusDelivered))
	if err != nil {
		return fmt.Errorf("failed to accept order: %v", err)
	}

	return nil
}

func (s *orderService) GetUserOrders(ctx context.Context, userId string) ([]models.OrderWithProductDetails, int64, error) {
	// Get all user orders without pagination
	orders, total, err := s.orderRepo.GetUserOrders(ctx, userId)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user orders: %v", err)
	}

	return orders, total, nil
}

func (s *orderService) GetUserOrderDetails(ctx context.Context, userId, orderId string) (*models.Order, error) {
	// Get order details for specific user
	order, err := s.orderRepo.GetUserOrderDetails(ctx, userId, orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get order details: %v", err)
	}

	return order, nil
}

func (s *orderService) CancelUserOrder(ctx context.Context, userId, orderId string) error {
	// Check if order can be cancelled (only pending orders)
	order, err := s.orderRepo.GetUserOrderDetails(ctx, userId, orderId)
	if err != nil {
		return fmt.Errorf("failed to get order details: %v", err)
	}

	fmt.Println("this is order in cancel user order service ; ", order)

	if order.Status != models.OrderStatusCreated {
		return fmt.Errorf("order cannot be cancelled in current status: %s", order.Status)
	}

	// Update order status to cancelled
	err = s.orderRepo.UpdateOrderStatus(ctx, orderId, string(models.OrderStatusCancelled))
	if err != nil {
		return fmt.Errorf("failed to cancel order: %v", err)
	}

	// Restore product stock
	err = s.restoreProductStock(ctx, order.Products)
	if err != nil {
		fmt.Printf("WARNING: Failed to restore product stock: %v\n", err)
	}

	return nil
}

func (s *orderService) GetSellerOrders(ctx context.Context, sellerId string, page, limit int, status string) ([]*models.Order, int64, error) {
	// Get all seller orders without pagination
	orders, total, err := s.orderRepo.GetSellerOrders(ctx, sellerId, 0, 0, status)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get seller orders: %v", err)
	}

	return orders, total, nil
}

func (s *orderService) GetSellerOrderDetails(ctx context.Context, sellerId, orderId string) (*models.Order, error) {
	// Get order details for seller
	order, err := s.orderRepo.GetSellerOrderDetails(ctx, sellerId, orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get order details: %v", err)
	}

	return order, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, sellerId, orderId, status string) error {
	// Validate status
	validStatuses := []string{
		string(models.OrderStatusCreated),
		string(models.OrderStatusPaidAndProcessing),
		string(models.OrderStatusShipping),
		string(models.OrderStatusDelivered),
		string(models.OrderStatusCancelled),
	}

	isValidStatus := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		return fmt.Errorf("invalid order status: %s", status)
	}

	// Update order status
	err := s.orderRepo.UpdateOrderStatus(ctx, orderId, status)
	if err != nil {
		return fmt.Errorf("failed to update order status: %v", err)
	}

	return nil
}

func (s *orderService) GetSellerProducts(ctx context.Context, sellerId string, page, limit int) ([]*models.Product, int64, error) {
	// Get seller products with pagination
	products, total, _, err := s.productRepo.GetSellerProducts(ctx, sellerId, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get seller products: %v", err)
	}

	return products, total, nil
}

func (s *orderService) UpdateProductStock(ctx context.Context, sellerId, productId string, stock int) error {
	// Update product stock
	err := s.productRepo.UpdateProductStock(ctx, sellerId, productId, stock)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %v", err)
	}

	return nil
}

func (s *orderService) GetSellerOrdersWithDetails(ctx context.Context, sellerId string) ([]models.OrderWithProductDetails, int64, error) {
	// Get all seller orders with details
	orders, total, err := s.orderRepo.GetSellerOrdersWithDetails(ctx, sellerId)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get seller orders with details: %v", err)
	}

	return orders, total, nil
}

func (s *orderService) AcceptOrder(ctx context.Context, sellerId, orderId string) error {
	// Update order status to paid and processing
	err := s.orderRepo.UpdateOrderStatus(ctx, orderId, string(models.OrderStatusPaidAndProcessing))
	if err != nil {
		return fmt.Errorf("failed to accept order: %v", err)
	}

	return nil
}

func (s *orderService) DeleteOrder(ctx context.Context, sellerId, orderId string) error {
	// Delete the order
	err := s.orderRepo.DeleteOrder(ctx, orderId)
	if err != nil {
		return fmt.Errorf("failed to delete order: %v", err)
	}

	return nil
}

// restoreProductStock restores product stock when order is cancelled
func (s *orderService) restoreProductStock(ctx context.Context, products []models.ProductItem) error {
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
