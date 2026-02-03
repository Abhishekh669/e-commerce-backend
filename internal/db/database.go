package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"e-commerce.com/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// MongoDB
	mongoClient     *mongo.Client
	mongoOnce       sync.Once
	mongoConnectErr error
	mongoDBName     = "e-commerce" // Set your default MongoDB database name

	// PostgreSQL (Neon)
	pgPool       *pgxpool.Pool
	pgOnce       sync.Once
	pgConnectErr error

	//redis connection
	redisClient     *redis.Client
	redisOnce       sync.Once
	redisConnectErr error

	// Mutexes
	mongoMutex sync.RWMutex
	pgMutex    sync.RWMutex
	redisMutex sync.RWMutex

	// PostgreSQL identifier validation regex
	pgIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// MongoDB Methods

// GetMongoCollection returns a MongoDB collection instance
func GetMongoCollection(collectionName string) (*mongo.Collection, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}

	return client.Database(mongoDBName).Collection(collectionName), nil
}

// GetMongoClient returns the MongoDB client (singleton)
func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		clientOpts := options.Client().
			ApplyURI(config.AppConfig.MongoDBURL).
			SetConnectTimeout(10 * time.Second).
			SetServerSelectionTimeout(10 * time.Second).
			SetMaxPoolSize(100).
			SetMinPoolSize(10)

		mongoMutex.Lock()
		defer mongoMutex.Unlock()

		client, err := mongo.Connect(context.Background(), clientOpts)
		if err != nil {
			mongoConnectErr = fmt.Errorf("failed to connect to MongoDB: %w", err)
			return
		}

		// Verify connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Ping(ctx, nil); err != nil {
			mongoConnectErr = fmt.Errorf("MongoDB ping failed: %w", err)
			return
		}

		mongoClient = client
		log.Println("✅ MongoDB connected successfully")
	})

	return mongoClient, mongoConnectErr
}

// PostgreSQL Methods

// GetPostgresConn returns a connection from the pgx pool
func GetPostgresConn(ctx context.Context) (*pgxpool.Conn, error) {
	pool, err := GetPostgresPool()
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection from pool: %w", err)
	}

	return conn, nil
}

// GetPostgresPool returns the PostgreSQL connection pool (singleton)
func GetPostgresPool() (*pgxpool.Pool, error) {
	pgOnce.Do(func() {
		pgMutex.Lock()
		defer pgMutex.Unlock()

		// Configure the connection pool
		config, err := pgxpool.ParseConfig(config.AppConfig.PostgressURL)
		if err != nil {
			pgConnectErr = fmt.Errorf("failed to parse PostgreSQL config: %w", err)
			return
		}

		// Pool settings
		config.MaxConns = 100
		config.MinConns = 10
		config.MaxConnLifetime = time.Hour
		config.MaxConnIdleTime = 30 * time.Minute
		config.HealthCheckPeriod = time.Minute

		// Create the pool
		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			pgConnectErr = fmt.Errorf("failed to create PostgreSQL pool: %w", err)
			return
		}

		// Verify the connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			pgConnectErr = fmt.Errorf("PostgreSQL ping failed: %w", err)
			return
		}

		pgPool = pool
		log.Println("✅ PostgreSQL (Neon) connected successfully")
	})

	return pgPool, pgConnectErr
}

// ValidateTableName validates a PostgreSQL table name to prevent SQL injection
func ValidateTableName(tableName string) error {
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	if len(tableName) > 63 {
		return fmt.Errorf("table name too long (max 63 characters)")
	}

	if !pgIdentifierRegex.MatchString(tableName) {
		return fmt.Errorf("invalid table name: must start with letter or underscore, contain only letters, numbers, and underscores")
	}

	return nil
}

// QueryPostgres executes a query and returns rows
func QueryPostgres(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	pool, err := GetPostgresPool()
	if err != nil {
		return nil, err
	}

	return pool.Query(ctx, query, args...)
}

// ExecPostgres executes a query without returning rows
func ExecPostgres(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	pool, err := GetPostgresPool()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	return pool.Exec(ctx, query, args...)
}

// func InitializeDatabase() error {

// 	log.Println("Initializing database connections...")

// 	log.Println("📊 Connecting to MongoDB...")
// 	if _, err := GetMongoClient(); err != nil {
// 		return err
// 	}

// 	log.Println("🐘 Connecting to PostgreSQL...")
// 	if _, err := GetPostgresPool(); err != nil {
// 		return err
// 	}
// 	return nil

// }

func GetRedisClient() (*redis.Client, error) {
	redisOnce.Do(func() {
		redisMutex.Lock()
		defer redisMutex.Unlock()

		client := redis.NewClient(&redis.Options{
			Addr:     config.AppConfig.RedisUrl,
			Username: config.AppConfig.RedisUserName,
			Password: config.AppConfig.RedisPassword,
			DB:       0,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			redisConnectErr = fmt.Errorf("redis ping failed: %w", err)
			return
		}

		redisClient = client
		log.Println("⚡ Redis connected successfully")
	})
	return redisClient, redisConnectErr
}

func InitializeDatabase() error {
	ctx := context.Background()
	var wg sync.WaitGroup
	var errMongo, errPostgres, errRedis error
	var postgressPool *pgxpool.Pool

	log.Println("Initializing database connections...")
	wg.Add(3)

	go func() {
		defer wg.Done()
		log.Println("📊 Connecting to MongoDB...")
		_, errMongo = GetMongoClient()
	}()

	go func() {
		defer wg.Done()
		log.Println("🐘 Connecting to PostgreSQL...")
		postgressPool, errPostgres = GetPostgresPool()
	}()

	go func() {
		defer wg.Done()
		log.Println("⚡ Connecting to Redis...")
		_, errRedis = GetRedisClient()
	}()

	wg.Wait()
	if errMongo != nil {
		return fmt.Errorf("MongoDB connection failed: %w", errMongo)
	}
	if errPostgres != nil {
		return fmt.Errorf("PostgreSQL connection failed: %w", errPostgres)
	}
	if errRedis != nil {
		return fmt.Errorf("redis connection failed: %w", errRedis)
	}

	if err := CreatePostgresTables(ctx, postgressPool); err != nil {
		return fmt.Errorf("error in creating table : %v", err)
	}
	return nil
}

// Cleanup closes all database connections
func Cleanup() {
	mongoMutex.Lock()
	defer mongoMutex.Unlock()

	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("⚠️ Failed to disconnect MongoDB: %v", err)
		} else {
			log.Println("🚪 MongoDB disconnected")
		}
	}

	pgMutex.Lock()
	defer pgMutex.Unlock()

	if pgPool != nil {
		pgPool.Close()
		log.Println("🚪 PostgreSQL (Neon) disconnected")
	}
	redisMutex.Lock()
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("⚠️ Failed to close Redis: %v", err)
		} else {
			log.Println("🚪 Redis disconnected")
		}
	}
	redisMutex.Unlock()

}

func SetupGracefulShutdown() {
	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🧹 Cleaning up database connections...")
		Cleanup()
		os.Exit(0)
	}()
}
