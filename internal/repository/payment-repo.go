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
)

type PaymentRepo interface {
	CreatePayment(ctx context.Context, payment *models.Payment) error
	CheckProductAvailability(ctx context.Context, productIds []string) ([]*models.Product, bool, error)
	GetPaymentByTransactionUUID(ctx context.Context, transactionUUID string) (*models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error
	ClearUserCart(ctx context.Context, userID string) error
}

type paymentRepo struct {
	pool        *pgxpool.Pool
	mongoClient *mongo.Client
	redisClient *redis.Client
}

func (r *paymentRepo) CheckProductAvailability(ctx context.Context, productIds []string) ([]*models.Product, bool, error) {
	// Validate input
	if productIds == nil || len(productIds) == 0 {
		return nil, false, fmt.Errorf("product IDs cannot be nil or empty")
	}

	fmt.Printf("DEBUG: Checking availability for product IDs: %v\n", productIds)

	collection := r.mongoClient.Database("ecommerce").Collection("products")

	// Create the query with proper validation
	query := bson.M{"_id": bson.M{"$in": productIds}}
	fmt.Printf("DEBUG: MongoDB query: %+v\n", query)

	cursor, err := collection.Find(ctx, query)
	if err != nil {
		return nil, false, fmt.Errorf("MongoDB query failed: %v", err)
	}
	defer cursor.Close(ctx)

	products := make([]*models.Product, 0)
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, false, fmt.Errorf("failed to decode products: %v", err)
	}

	fmt.Printf("DEBUG: Found %d products in database\n", len(products))
	return products, true, nil
}
func (r *paymentRepo) CreatePayment(ctx context.Context, payment *models.Payment) error {
	collection := r.mongoClient.Database("ecommerce").Collection("payments")
	_, err := collection.InsertOne(ctx, payment)
	if err != nil {
		return err
	}
	return nil
}

func (r *paymentRepo) GetPaymentByTransactionUUID(ctx context.Context, transactionUUID string) (*models.Payment, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("payments")
	var payment models.Payment

	// Use the correct field name from the model: TransactionUuid
	err := collection.FindOne(ctx, bson.M{"transactionUuid": transactionUUID}).Decode(&payment)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment with transaction UUID %s: %v", transactionUUID, err)
	}

	fmt.Printf("DEBUG: Found payment with ID: %s, ProductIDs: %v\n", payment.ID, payment.ProductIDs)
	return &payment, nil
}

func (r *paymentRepo) UpdatePaymentStatus(ctx context.Context, paymentID string, status models.PaymentStatus) error {
	collection := r.mongoClient.Database("ecommerce").Collection("payments")
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": paymentID},
		bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now()}},
	)
	return err
}

func (r *paymentRepo) ClearUserCart(ctx context.Context, userID string) error {
	// Clear the user's cart from Redis
	cartKey := fmt.Sprintf("cart:%s", userID)
	err := r.redisClient.Del(ctx, cartKey).Err()
	if err != nil {
		return fmt.Errorf("failed to clear cart for user %s: %v", userID, err)
	}

	fmt.Printf("DEBUG: Cleared cart for user %s\n", userID)
	return nil
}

func NewPaymentRepository() PaymentRepo {
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
	return &paymentRepo{
		pool:        pool,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}
