package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"e-commerce.com/internal/db"
	"e-commerce.com/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentRepo interface {
	CreateComment(ctx context.Context, order *models.ProductReview) error
}

type commentRepo struct {
	pool        *pgxpool.Pool
	mongoClient *mongo.Client
	redisClient *redis.Client
}

func (r *commentRepo) CreateComment(ctx context.Context, review *models.ProductReview) error {
	commentsCol := r.mongoClient.Database("ecommerce").Collection("comments")
	productsCol := r.mongoClient.Database("ecommerce").Collection("products")

	// 1️⃣ Insert the new review
	if _, err := commentsCol.InsertOne(ctx, review); err != nil {
		fmt.Println("error creating comment:", err)
		return fmt.Errorf("failed to insert review: %w", err)
	}

	// 2️⃣ Find the product
	filter := bson.M{"_id": review.ProductId}

	var product models.Product
	if err := productsCol.FindOne(ctx, filter).Decode(&product); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("product not found for ID %s", review.ProductId)
		}
		return fmt.Errorf("failed to find product: %w", err)
	}

	// 3️⃣ Compute new average rating
	newCount := product.RatingCount + 1
	newAvg := ((product.Rating * (product.RatingCount)) + (review.Rating)) / (newCount)

	// Ensure it's clamped to range (e.g. 1–6)
	if newAvg < 1 {
		newAvg = 1
	} else if newAvg > 5 {
		newAvg = 5
	}

	// 4️⃣ Update rating atomically
	update := bson.M{
		"$set": bson.M{
			"rating":    newAvg,
			"updatedAt": time.Now(),
		},
		"$inc": bson.M{
			"ratingCount": 1,
		},
	}

	if _, err := productsCol.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("failed to update product rating: %w", err)
	}

	return nil
}

func NewCommentRepositry() CommentRepo {
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
	return &commentRepo{
		pool:        pool,
		mongoClient: mongoClient,
		redisClient: redisClient,
	}
}
