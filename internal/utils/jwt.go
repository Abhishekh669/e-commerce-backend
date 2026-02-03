package utils

import (
	"fmt"
	"time"

	"e-commerce.com/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId   string `json:"userId"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type JwtDataType struct {
	UserId   string
	FullName string
	Email    string
}

func GenerateJWT(jwtData JwtDataType) (string, error) {
	if jwtData.Email == "" || jwtData.FullName == "" || jwtData.UserId == "" {
		return "", fmt.Errorf("invalid user data")
	}

	jwtSecret := config.AppConfig.JWTSecret

	if jwtSecret == "" {
		return "", fmt.Errorf("invalid jwt secret")
	}

	claims := &Claims{
		UserId:   jwtData.UserId,
		FullName: jwtData.FullName,
		Email:    jwtData.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "golang-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ParseJwt(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("no valid token string")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexprected signing method : %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
