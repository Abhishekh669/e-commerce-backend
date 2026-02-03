package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"e-commerce.com/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Example structs for your e-commerce app
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"-"` // Don't include in JSON
	Created  time.Time          `bson:"created" json:"created"`
}

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CategoryID  int       `json:"category_id"`
	Created     time.Time `json:"created"`
}

// MongoDB Usage Examples
func MongoDBExamples() {
	ctx := context.Background()

	// 1. Get a collection
	userCollection, err := db.GetMongoCollection("users")
	if err != nil {
		log.Fatal("Failed to get users collection:", err)
	}

	// 2. Insert a new user
	newUser := User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "hashedpassword123",
		Created:  time.Now(),
	}

	result, err := userCollection.InsertOne(ctx, newUser)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return
	}
	fmt.Printf("Inserted user with ID: %v\n", result.InsertedID)

	// 3. Find a user by email
	var foundUser User
	filter := bson.M{"email": "john@example.com"}
	err = userCollection.FindOne(ctx, filter).Decode(&foundUser)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		return
	}
	fmt.Printf("Found user: %+v\n", foundUser)

	// 4. Update a user
	updateFilter := bson.M{"email": "john@example.com"}
	update := bson.M{
		"$set": bson.M{
			"name": "John Smith",
		},
	}
	updateResult, err := userCollection.UpdateOne(ctx, updateFilter, update)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return
	}
	fmt.Printf("Updated %d document(s)\n", updateResult.ModifiedCount)

	// 5. Find multiple users with options
	findOptions := options.Find()
	findOptions.SetLimit(10)
	findOptions.SetSort(bson.D{{"created", -1}}) // Sort by created date, newest first

	cursor, err := userCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("Error finding users: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		log.Printf("Error decoding users: %v", err)
		return
	}

	fmt.Printf("Found %d users\n", len(users))
	for _, user := range users {
		fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
	}

	// 6. Delete a user
	deleteFilter := bson.M{"email": "john@example.com"}
	deleteResult, err := userCollection.DeleteOne(ctx, deleteFilter)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		return
	}
	fmt.Printf("Deleted %d document(s)\n", deleteResult.DeletedCount)
}

// PostgreSQL Usage Examples
func PostgreSQLExamples() {
	ctx := context.Background()

	// 1. Validate table name before using (important for security)
	tableName := "products"
	if err := db.ValidateTableName(tableName); err != nil {
		log.Printf("Invalid table name: %v", err)
		return
	}

	// 2. Create a table (you might do this in migrations)
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			stock INTEGER NOT NULL DEFAULT 0,
			category_id INTEGER,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	
	_, err := db.ExecPostgres(ctx, createTableQuery)
	if err != nil {
		log.Printf("Error creating table: %v", err)
		return
	}
	fmt.Println("Products table created/verified")

	// 3. Insert a product
	insertQuery := `
		INSERT INTO products (name, description, price, stock, category_id) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created`
	
	var productID int
	var created time.Time
	err = db.QueryPostgres(ctx, insertQuery, "iPhone 14", "Latest iPhone model", 999.99, 50, 1).Scan(&productID, &created)
	if err != nil {
		log.Printf("Error inserting product: %v", err)
		return
	}
	fmt.Printf("Inserted product with ID: %d at %v\n", productID, created)

	// 4. Query products with parameters
	selectQuery := `
		SELECT id, name, description, price, stock, category_id, created 
		FROM products 
		WHERE price BETWEEN $1 AND $2 
		ORDER BY created DESC`
	
	rows, err := db.QueryPostgres(ctx, selectQuery, 500.0, 1500.0)
	if err != nil {
		log.Printf("Error querying products: %v", err)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.Created)
		if err != nil {
			log.Printf("Error scanning product: %v", err)
			continue
		}
		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return
	}

	fmt.Printf("Found %d products in price range\n", len(products))
	for _, product := range products {
		fmt.Printf("Product: %s - $%.2f (Stock: %d)\n", product.Name, product.Price, product.Stock)
	}

	// 5. Update product stock
	updateQuery := `UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`
	result, err := db.ExecPostgres(ctx, updateQuery, 5, productID)
	if err != nil {
		log.Printf("Error updating product stock: %v", err)
		return
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Println("No products updated (insufficient stock or product not found)")
	} else {
		fmt.Printf("Updated stock for %d product(s)\n", rowsAffected)
	}

	// 6. Get a single product
	var product Product
	getQuery := `SELECT id, name, description, price, stock, category_id, created FROM products WHERE id = $1`
	row := db.QueryPostgres(ctx, getQuery, productID)
	err = row.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CategoryID, &product.Created)
	if err != nil {
		log.Printf("Error getting product: %v", err)
		return
	}
	fmt.Printf("Retrieved product: %+v\n", product)
}

// Using with transactions
func PostgreSQLTransactionExample() {
	ctx := context.Background()
	
	// Get a connection for the transaction
	conn, err := db.GetPostgresConn(ctx)
	if err != nil {
		log.Printf("Error getting connection: %v", err)
		return
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return
	}

	// Ensure we rollback on error
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Perform multiple operations in transaction
	_, err = tx.Exec(ctx, "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3)", "Product A", 10.99, 100)
	if err != nil {
		log.Printf("Error inserting Product A: %v", err)
		return
	}

	_, err = tx.Exec(ctx, "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3)", "Product B", 15.99, 75)
	if err != nil {
		log.Printf("Error inserting Product B: %v", err)
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	fmt.Println("Transaction completed successfully")
}

// Service layer example - how you might structure your business logic
type UserService struct{}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*User, error) {
	// Get collection
	collection, err := db.GetMongoCollection("users")
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Check if user already exists
	var existingUser User
	err = collection.FindOne(ctx, bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Create new user
	user := User{
		Name:     name,
		Email:    email,
		Password: password, // In real app, hash this!
		Created:  time.Now(),
	}

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return &user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	collection, err := db.GetMongoCollection("users")
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var user User
	err = collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func main() {
	// Make sure to cleanup connections when your app shuts down
	defer db.Cleanup()

	fmt.Println("=== MongoDB Examples ===")
	MongoDBExamples()

	fmt.Println("\n=== PostgreSQL Examples ===")
	PostgreSQLExamples()

	fmt.Println("\n=== PostgreSQL Transaction Example ===")
	PostgreSQLTransactionExample()

	fmt.Println("\n=== Service Layer Example ===")
	userService := &UserService{}
	
	ctx := context.Background()
	user, err := userService.CreateUser(ctx, "Alice Johnson", "alice@example.com", "hashedpassword")
	if err != nil {
		log.Printf("Error creating user: %v", err)
	} else {
		fmt.Printf("Created user: %+v\n", user)
	}
}