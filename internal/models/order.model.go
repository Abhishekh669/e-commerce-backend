package models

import (
	"time"
)

type ProductItem struct {
	ProductID string `json:"productId" bson:"productId"`
	SellerID  string `json:"sellerId" bson:"sellerId"` // Add seller ID to track which seller the product belongs to
	Quantity  int64  `json:"quantity" bson:"quantity"`
	Price     int64  `json:"price" bson:"price"`
}

type Order struct {
	ID            string        `json:"id" bson:"_id,omitempty"`
	User          string        `json:"userId" bson:"userId"`
	Amount        int64         `json:"amount" bson:"amount"`
	Products      []ProductItem `json:"products,omitempty" bson:"products"`
	TransactionID string        `json:"transactionId" bson:"transactionId"`
	Status        OrderStatus   `json:"status" bson:"status"`
	CreatedAt     time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt" bson:"updatedAt"`
}

type OrderStatus string

const (
	OrderStatusCreated           OrderStatus = "created"
	OrderStatusPaidAndProcessing OrderStatus = "paid and processing"
	OrderStatusShipping          OrderStatus = "shipping"
	OrderStatusDelivered         OrderStatus = "delivered"
	OrderStatusCancelled         OrderStatus = "cancelled"
)

type OrderWithProductDetails struct {
	ID            string                `json:"id" bson:"_id,omitempty"`
	User          string                `json:"userId" bson:"userId"`
	Amount        int64                 `json:"amount" bson:"amount"`
	Products      []ProductWithQuantity `json:"products,omitempty" bson:"products"`
	TransactionID string                `json:"transactionId" bson:"transactionId"`
	Status        OrderStatus           `json:"status" bson:"status"`
	CreatedAt     time.Time             `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time             `json:"updatedAt" bson:"updatedAt"`
}

type ProductWithQuantity struct {
	ID             string      `json:"id" bson:"_id"`
	Name           string      `json:"name" bson:"name"`
	Description    string      `json:"description" bson:"description"`
	Price          int         `json:"price" bson:"price"`
	Quantity       int         `json:"quantity" bson:"quantity"`
	Discount       int         `json:"discount" bson:"discount"`
	SellerID       string      `json:"sellerId" bson:"sellerId"`
	Category       interface{} `json:"category" bson:"category"`
	Images         []string    `json:"images" bson:"images"`
	Stock          int         `json:"stock" bson:"stock"`
	CreatedAt      time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt" bson:"updatedAt"`
	BoughtQuantity int64       `json:"boughtQuantity" bson:"boughtQuantity"`
}
