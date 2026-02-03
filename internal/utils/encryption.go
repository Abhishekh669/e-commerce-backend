package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

func CheckPassword(hashedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return fmt.Errorf("incorrect password")
		default:
			return fmt.Errorf("failed to compare passwords: %w", err)
		}
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SaveVerificationToken(ctx context.Context, rdb *redis.Client, userId string) (string, error) {
	const maxAttempts = 5

	for i := 0; i < maxAttempts; i++ {
		token, err := generateToken()
		if err != nil {
			return "", err
		}
		key := "verifyToken:" + token

		success, err := rdb.SetNX(ctx, key, userId, 8*time.Minute).Result()
		if err != nil {
			return "", err
		}

		if success {
			return token, nil
		}
	}
	return "", errors.New("faield to generate unique token after multiple attempts")

}
