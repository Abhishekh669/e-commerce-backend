package repository

import (
	"context"
	"fmt"
	"sync"

	"e-commerce.com/internal/db"
	"e-commerce.com/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductRepo interface {
	CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	GetSellerProducts(ctx context.Context, sellerId string, page int, limit int) ([]*models.Product, int64, bool, error)
	UpdateProduct(ctx context.Context, productId string, product *models.UpdateProductRequest) error
	DeleteProduct(ctx context.Context, productId string) error
	GetProductByID(ctx context.Context, productId string) (*models.Product, []models.ProductReview, error)
	GetAllProducts(ctx context.Context, search *string, limit, offset int) (*models.ProductResponse, error)
	UpdateProductStock(ctx context.Context, sellerId, productId string, stock int) error
}

type productRepo struct {
	pool        *pgxpool.Pool
	mongoClient *mongo.Client
	redisClient *redis.Client
}

func (r *productRepo) GetAllProducts(ctx context.Context, search *string, limit, offset int) (*models.ProductResponse, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("products")

	// Build the base filter
	filter := bson.M{}

	// Add search filter if search term is provided
	if search != nil && *search != "" {
		regex := bson.M{"$regex": *search, "$options": "i"}
		filter["$or"] = []bson.M{
			{"name": regex},        // Search in product name
			{"description": regex}, // Search in product description
			{"category": regex},    // Search in category
		}
	}

	// Projection for query
	projection := bson.M{
		"_id":         1,
		"name":        1,
		"description": 1,
		"price":       1,
		"quantity":    1,
		"discount":    1,
		"sellerId":    1,
		"category":    1,
		"images":      1,
		"stock":       1,
		"rating":      1,
		"createdAt":   1,
		"updatedAt":   1,
	}

	opts := options.Find().
		SetProjection(projection).
		SetLimit(int64(limit)).
		SetSkip(int64(offset * limit)).
		SetSort(bson.M{"createdAt": -1})

	// Use the filter that includes search conditions
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*models.Product
	var total int64

	var decodedErr, countErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := cursor.All(ctx, &products); err != nil {
			decodedErr = err
		}
	}()
	go func() {
		defer wg.Done()
		// Use the same filter for counting to include search conditions
		total, err = collection.CountDocuments(ctx, filter)
		if err != nil {
			countErr = err
		}
	}()
	wg.Wait()

	fmt.Println("this is the error for decode: ", decodedErr)
	fmt.Println("this is the count err : ", countErr)

	if decodedErr != nil {
		return nil, decodedErr
	}
	if countErr != nil {
		return nil, countErr
	}

	hasMore := (offset+1)*limit < int(total)
	nextOffset := offset + 1

	productResponse := models.ProductResponse{
		Products:   products,
		Total:      int(total),
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}

	fmt.Println("thisis  prproductreponse : ", productResponse.Products)

	return &productResponse, nil
}

func (r *productRepo) UpdateProduct(ctx context.Context, productId string, product *models.UpdateProductRequest) error {
	collection := r.mongoClient.Database("ecommerce").Collection("products")
	filter := bson.M{"_id": productId}
	update := bson.M{"$set": bson.M{
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"quantity":    product.Quantity,
		"discount":    product.Discount,
		"sellerId":    product.SellerID,
		"category":    product.Category,
		"images":      product.Images,
		"stock":       product.Stock,
	}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *productRepo) DeleteProduct(ctx context.Context, productId string) error {
	collection := r.mongoClient.Database("ecommerce").Collection("products")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": productId})
	if err != nil {
		return err
	}
	return nil
}

func (r *productRepo) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("products")
	_, err := collection.InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepo) GetSellerProducts(ctx context.Context, sellerId string, page int, limit int) ([]*models.Product, int64, bool, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("products")

	// Validate input parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Maximum limit to prevent excessive data retrieval
	}

	filter := bson.M{"sellerId": sellerId}

	// Get total count first
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, false, err
	}

	// Calculate hasMore based on pagination logic
	// Handle edge cases
	if total == 0 {
		return []*models.Product{}, 0, false, nil
	}

	totalPages := int64((int(total) + limit - 1) / limit) // Ceiling division
	hasMore := int64(page) < totalPages

	// Set up find options with pagination and sorting
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, false, err
	}
	defer cursor.Close(ctx)

	products := []*models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, 0, false, err
		}
		products = append(products, &product)
	}

	return products, total, hasMore, nil
}

func (r *productRepo) GetProductByID(ctx context.Context, productId string) (*models.Product, []models.ProductReview, error) {
	collection := r.mongoClient.Database("ecommerce").Collection("products")
	commentCollection := r.mongoClient.Database("ecommerce").Collection("comments")

	filter := bson.M{"_id": productId}
	commentFilter := bson.M{"productId": productId}

	var product models.Product
	var comments []models.ProductReview

	// 1️⃣ Get the product
	if err := collection.FindOne(ctx, filter).Decode(&product); err != nil {
		return nil, nil, err
	}

	// 2️⃣ Get all comments for the product
	cursor, err := commentCollection.Find(ctx, commentFilter)
	if err != nil {
		return &product, nil, err
	}
	defer cursor.Close(ctx)

	// 3️⃣ Decode all comments (can be empty — not an error)
	if err := cursor.All(ctx, &comments); err != nil {
		return &product, nil, err
	}

	return &product, comments, nil
}

func (r *productRepo) UpdateProductStock(ctx context.Context, sellerId, productId string, stock int) error {
	collection := r.mongoClient.Database("ecommerce").Collection("products")
	filter := bson.M{"_id": productId, "sellerId": sellerId}
	update := bson.M{"$set": bson.M{"stock": stock}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func NewProductRepository() ProductRepo {
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
	return &productRepo{
		pool:        pool,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}
