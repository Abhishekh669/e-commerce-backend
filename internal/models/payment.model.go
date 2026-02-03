package models

import "time"

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusSuccess  PaymentStatus = "success"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type Payment struct {
	ID              string        `json:"id" bson:"_id"`
	Amount          int64         `json:"amount" bson:"amount"`
	UserId          string        `json:"userId" bson:"userId"`
	TransactionUuid string        `json:"transactionUuid" bson:"transactionUuid"`
	ProductIDs      []string      `json:"productIds" bson:"productIds"`
	Status          PaymentStatus `json:"status" bson:"status"`
	CreatedAt       time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt" bson:"updatedAt"`
}
