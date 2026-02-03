package models

import "time"

type Product struct {
	ID          string      `json:"id" bson:"_id"`
	Name        string      `json:"name" bson:"name"`
	Description string      `json:"description" bson:"description"`
	Price       int         `json:"price" bson:"price"`
	Quantity    int         `json:"quantity" bson:"quantity"`
	Discount    int         `json:"discount" bson:"discount"`
	SellerID    string      `json:"sellerId" bson:"sellerId"`
	Category    interface{} `json:"category" bson:"category"`
	Images      []string    `json:"images" bson:"images"`
	Stock       int         `json:"stock" bson:"stock"`
	Rating      int         `json:"rating" bson:"rating"`
	RatingCount int         `json:"ratingCount" bson:"ratingCount"`
	CreatedAt   time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt" bson:"updatedAt"`
}

type UpdateProductRequest struct {
	Name        string      `json:"name" bson:"name"`
	Description string      `json:"description" bson:"description"`
	Price       int         `json:"price" bson:"price"`
	Quantity    int         `json:"quantity" bson:"quantity"`
	Discount    int         `json:"discount" bson:"discount"`
	SellerID    string      `json:"sellerId" bson:"sellerId"`
	Category    interface{} `json:"category" bson:"category"`
	Images      []string    `json:"images" bson:"images"`
	Stock       int         `json:"stock" bson:"stock"`
}

type CreateProductRequest struct {
	Name        string      `json:"name" bson:"name"`
	Description string      `json:"description" bson:"description"`
	Price       int         `json:"price" bson:"price"`
	Quantity    int         `json:"quantity" bson:"quantity"`
	Discount    int         `json:"discount" bson:"discount"`
	SellerID    string      `json:"sellerId" bson:"sellerId"`
	Category    interface{} `json:"category" bson:"category"`
	Images      []string    `json:"images" bson:"images"`
	Stock       int         `json:"stock" bson:"stock"`
	Rating      int         `json:"rating" bson:"rating"`
}

type ProductResponse struct {
	Products   []*Product `json:"products"`
	Total      int        `json:"total"`
	HasMore    bool       `json:"hasMore"`
	NextOffset int        `json:"nextOffset"`
}
