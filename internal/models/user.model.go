package models

import (
	"time"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleSeller   Role = "seller"
)

type User struct {
	ID         string    `json:"id" db:"id"`                       // UUID
	Username   string    `json:"userName" db:"username"`           // snake_case for DB
	Email      string    `json:"email" db:"email"`                 // unique
	Password   string    `json:"password,omitempty" db:"password"` // hashed
	Role       Role      `json:"role" db:"role"`                   // ENUM: customer/seller
	IsVerified bool      `json:"isVerified" db:"is_verified"`      // snake_case
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`        // snake_case
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`        // snake_case
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SafeUser struct {
	ID         string    `json:"id"`
	Username   string    `json:"userName"`
	Email      string    `json:"email"`
	Role       Role      `json:"role"`
	IsVerified bool      `json:"isVerified"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
