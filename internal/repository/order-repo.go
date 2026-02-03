package repository

import (
	"context"
	"fmt"
	"time"

	"e-commerce.com/internal/db"
	"e-commerce.com/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByTransactionID(ctx context.Context, transactionID string) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
	GetUserOrders(ctx context.Context, userId string) ([]models.OrderWithProductDetails, int64, error)
	GetUserOrderDetails(ctx context.Context, userId, orderId string) (*models.Order, error)
	GetSellerOrders(ctx context.Context, sellerId string, skip, limit int, status string) ([]*models.Order, int64, error)
	GetSellerOrderDetails(ctx context.Context, sellerId, orderId string) (*models.Order, error)
	GetSellerOrdersWithDetails(ctx context.Context, sellerId string) ([]models.OrderWithProductDetails, int64, error)
	AcceptOrder(ctx context.Context, orderId string) error
	DeleteOrder(ctx context.Context, orderId string) error
}

type orderRepo struct {
	pool        *pgxpool.Pool
	mongoClient *mongo.Client
	redisClient *redis.Client
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *models.Order) error {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")
	_, err := collection.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (r *orderRepo) GetOrderByTransactionID(ctx context.Context, transactionID string) (*models.Order, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")
	var order models.Order
	err := collection.FindOne(ctx, bson.M{"transactionId": transactionID}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepo) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now()}},
	)
	return err
}

func (r *orderRepo) GetUserOrders(ctx context.Context, userId string) ([]models.OrderWithProductDetails, int64, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")

	// Build filter
	filter := bson.M{"userId": userId}

	// Set sort options (newest first)
	opts := options.Find().SetSort(bson.M{"createdAt": -1})

	// Find all orders for the user
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []models.OrderWithProductDetails
	for cursor.Next(ctx) {
		var order models.Order
		var orderWithProductDetails models.OrderWithProductDetails
		if err := cursor.Decode(&order); err != nil {
			return nil, 0, err
		}

		// Get product details with quantities
		productsWithQuantity, err := r.getProductDetailsWithQuantity(ctx, order.Products)
		if err != nil {
			return nil, 0, err
		}
		orderWithProductDetails.Products = productsWithQuantity
		orderWithProductDetails.ID = order.ID
		orderWithProductDetails.User = order.User
		orderWithProductDetails.Amount = order.Amount
		orderWithProductDetails.TransactionID = order.TransactionID
		orderWithProductDetails.Status = order.Status
		orderWithProductDetails.CreatedAt = order.CreatedAt
		orderWithProductDetails.UpdatedAt = order.UpdatedAt

		orders = append(orders, orderWithProductDetails)
	}

	return orders, int64(len(orders)), nil
}

func (r *orderRepo) getProductDetailsWithQuantity(ctx context.Context, productData []models.ProductItem) ([]models.ProductWithQuantity, error) {
	product_collection := r.mongoClient.Database("ecommerce").Collection("products")
	var productsWithQuantity []models.ProductWithQuantity

	for _, productItem := range productData {
		var product models.Product
		err := product_collection.FindOne(ctx, bson.M{"_id": productItem.ProductID}).Decode(&product)
		if err != nil {
			return nil, err
		}

		productWithQuantity := models.ProductWithQuantity{
			ID:             product.ID,
			Name:           product.Name,
			Description:    product.Description,
			Price:          product.Price,
			Quantity:       product.Quantity,
			Discount:       product.Discount,
			SellerID:       product.SellerID,
			Category:       product.Category,
			Images:         product.Images,
			Stock:          product.Stock,
			CreatedAt:      product.CreatedAt,
			UpdatedAt:      product.UpdatedAt,
			BoughtQuantity: productItem.Quantity,
		}
		productsWithQuantity = append(productsWithQuantity, productWithQuantity)
	}
	return productsWithQuantity, nil
}

func (r *orderRepo) GetUserOrderDetails(ctx context.Context, userId, orderId string) (*models.Order, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")
	var order models.Order
	err := collection.FindOne(ctx, bson.M{"_id": orderId, "userId": userId}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepo) GetSellerOrders(ctx context.Context, sellerId string, skip, limit int, status string) ([]*models.Order, int64, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")

	// Build filter: find orders that contain products with the given sellerId
	filter := bson.M{
		"products.sellerId": sellerId, // Match nested product sellerId
	}
	if status != "" {
		filter["status"] = status
	}

	// Count total orders before pagination
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		fmt.Println("eroror is here so fix me ")
		return nil, 0, err
	}

	// Set sort and pagination options
	opts := options.Find().
		SetSort(bson.M{"createdAt": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	// Query orders
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		fmt.Println("thierie not data : ':", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sellerOrders []*models.Order
	if err = cursor.All(ctx, &sellerOrders); err != nil {
		fmt.Println("erro in the seting all teh data :',", err)
		return nil, 0, err
	}

	return sellerOrders, total, nil
}

func (r *orderRepo) GetSellerOrderDetails(ctx context.Context, sellerId, orderId string) (*models.Order, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")
	var order models.Order
	err := collection.FindOne(ctx, bson.M{"_id": orderId}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepo) GetSellerOrdersWithDetails(ctx context.Context, sellerId string) ([]models.OrderWithProductDetails, int64, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")

	// Build filter for seller orders (orders containing seller's products)
	filter := bson.M{"products.sellerId": sellerId}

	// Set sort options (newest first)
	opts := options.Find().SetSort(bson.M{"createdAt": -1})

	// Find all orders
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var allOrders []*models.Order
	err = cursor.All(ctx, &allOrders)
	if err != nil {
		return nil, 0, err
	}

	// Filter orders that contain seller's products
	var sellerOrdersWithDetails []models.OrderWithProductDetails
	for _, order := range allOrders {
		var orderWithDetails models.OrderWithProductDetails
		orderWithDetails.ID = order.ID
		orderWithDetails.User = order.User
		orderWithDetails.Amount = order.Amount
		orderWithDetails.TransactionID = order.TransactionID
		orderWithDetails.Status = order.Status
		orderWithDetails.CreatedAt = order.CreatedAt
		orderWithDetails.UpdatedAt = order.UpdatedAt

		// Get product details with quantities
		productsWithQuantity, err := r.getProductDetailsWithQuantity(ctx, order.Products)
		if err != nil {
			return nil, 0, err
		}
		orderWithDetails.Products = productsWithQuantity
		sellerOrdersWithDetails = append(sellerOrdersWithDetails, orderWithDetails)
	}

	return sellerOrdersWithDetails, int64(len(sellerOrdersWithDetails)), nil
}

func (r *orderRepo) AcceptOrder(ctx context.Context, orderId string) error {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": orderId},
		bson.M{"$set": bson.M{"status": "accepted", "updatedAt": time.Now()}},
	)
	return err
}

func (r *orderRepo) DeleteOrder(ctx context.Context, orderId string) error {
	collection := r.mongoClient.Database("ecommerce").Collection("orders")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": orderId})
	return err
}

func NewOrderRepository() OrderRepo {
	mongoClient, err := db.GetMongoClient()
	if err != nil {
		return nil
	}
	redisClient, err := db.GetRedisClient()
	if err != nil {
		return nil
	}
	pool, err := db.GetPostgresPool()
	if err != nil {
		return nil
	}
	return &orderRepo{
		pool:        pool,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}
