package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"e-commerce.com/internal/db"
	"e-commerce.com/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	VerifyTokenFromRedis(token string, ctx context.Context) (*models.User, error)
	CreateUserDataInRedis(user *models.User, ctx context.Context) error
	GetUserSessionFromRedis(ctx context.Context, key string) (*models.User, error)
	DeleteUserSessionFromRedis(ctx context.Context, userId string) error
	LoginUser(ctx context.Context, userLogin *models.UserLogin) (*models.User, error)
	CreateUserSessionInRedis(ctx context.Context, user *models.User) error
}

type userRepo struct {
	pool        *pgxpool.Pool
	redisClient *redis.Client
}

func (r *userRepo) CreateUserSessionInRedis(ctx context.Context, user *models.User) error {
	sessionKey := fmt.Sprintf("userId:usersession:%s", user.ID)
	exist, _ := r.redisClient.Exists(ctx, sessionKey).Result()
	if exist == 0 {
		return fmt.Errorf("user session not found")
	}
	return nil
}

func (r *userRepo) LoginUser(ctx context.Context, userLogin *models.UserLogin) (*models.User, error) {

	user, err := r.GetUserByEmail(userLogin.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	fmt.Println("this is the user data : ", user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userLogin.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	sessionKey := fmt.Sprintf("userId:usersession:%s", user.ID)
	exist, _ := r.redisClient.Exists(ctx, sessionKey).Result()
	if exist == 0 {
		if err := r.CreateUserSessionInRedis(ctx, user); err != nil {
			fmt.Println("failed to create user data in redis: %w", err)

		}
	}

	return user, nil

}

func (r *userRepo) DeleteUserSessionFromRedis(ctx context.Context, userId string) error {
	exists, err := r.redisClient.Exists(ctx, userId).Result()
	if err != nil {
		return fmt.Errorf("failed to check user session: %w", err)
	}

	// If no session exists, just return nil (no error)
	if exists == 0 {
		return nil
	}

	// Delete only if it exists
	_, err = r.redisClient.Del(ctx, userId).Result()
	if err != nil {
		return fmt.Errorf("failed to delete user session: %w", err)
	}

	return nil
}

func (r *userRepo) CreateUserDataInRedis(user *models.User, ctx context.Context) error {
	userJson, err := json.Marshal(user)

	if err != nil {
		return fmt.Errorf("failed to encode user: %w", err)
	}

	rerr := r.redisClient.Set(ctx, user.Email, userJson, 10*time.Minute).Err()

	if rerr != nil {
		return fmt.Errorf("failed to store data: %w", rerr)
	}

	return nil
}

func (r *userRepo) VerifyTokenFromRedis(token string, ctx context.Context) (*models.User, error) {
	if token == "" {
		return nil, fmt.Errorf("invalid token")
	}

	// Step 1: Get user ID from Redis
	userEmail, err := r.redisClient.Get(ctx, token).Result()
	if err == redis.Nil {
		fmt.Println("this is error in not data ")
		return nil, fmt.Errorf("token not found in Redis")
	} else if err != nil {
		fmt.Println("this is error in not data in chekcin : ", err)
		return nil, fmt.Errorf("redis error: %v", err)
	}

	userData, err := r.redisClient.Get(ctx, userEmail).Result()

	if err == redis.Nil {
		return nil, fmt.Errorf("token not found in Redis")
	} else if err != nil {
		return nil, fmt.Errorf("redis error: %v", err)
	}

	var user models.User
	if err := json.Unmarshal([]byte(userData), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
	}
	user.IsVerified = true

	if err := r.CreateUser(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to create user in database: %v", err)
	}

	// Step 6: Clean up Redis (token and userID keys)
	r.redisClient.Del(ctx, token)
	r.redisClient.Del(ctx, userEmail)

	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user session: %v", err)
	}
	sessionKey := fmt.Sprintf("userId:usersession:%s", user.ID)
	if err := r.redisClient.Set(ctx, sessionKey, userJSON, 6*24*time.Hour).Err(); err != nil {
		return nil, fmt.Errorf("failed to store user session in Redis: %v", err)
	}

	return &user, nil

}

func (r *userRepo) GetUserSessionFromRedis(ctx context.Context, key string) (*models.User, error) {
	var user models.User
	fmt.Println("this isht e key okie : ", key)

	// Get raw JSON from Redis
	data, err := r.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	} else if err != nil {
		return nil, fmt.Errorf("redis error: %v", err)
	}

	// Unmarshal JSON into user struct
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user session: %v", err)
	}

	return &user, nil
}

func (r *userRepo) CreateUser(ctx context.Context, user *models.User) error {
	if user.Email == "" || user.Password == "" || user.Username == "" {
		return fmt.Errorf("missing required fields")
	}

	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	if user.Role == "" {
		user.Role = models.RoleCustomer
	}

	query := `
        INSERT INTO users (
            id, 
            username, 
            email, 
            password, 
            role,
            is_verified,
            created_at,
            updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        )
        ON CONFLICT (email) DO NOTHING
    `

	result, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	return nil
}

func (r *userRepo) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT 
			id, username, email, password, role, is_verified, created_at, updated_at
		FROM users 
		WHERE email = $1
	`
	var user models.User

	err := r.pool.QueryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func NewUserRepository() UserRepo {
	pool, err := db.GetPostgresPool()
	if err != nil {
		return nil
	}
	redisClient, err := db.GetRedisClient()

	if err != nil {
		return nil
	}
	return &userRepo{pool: pool, redisClient: redisClient}
}
