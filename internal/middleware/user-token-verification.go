package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"e-commerce.com/internal/utils"
	"github.com/gin-gonic/gin"
)

// UserTokenVerification middleware verifies JWT tokens and extracts user information
func UserTokenVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := utils.ExtractToken(c, "user_token")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Token is required",
			})
			c.Abort()
		}

		// Parse and validate the JWT token
		claims, err := utils.ParseJwt(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// Validate that required claims are present
		if claims.UserId == "" || claims.Email == "" || claims.FullName == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Token missing required user information",
			})
			c.Abort()
			return
		}

		// Store user information in the context for use in handlers
		c.Set("userId", claims.UserId)
		c.Set("userEmail", claims.Email)
		c.Set("userFullName", claims.FullName)
		c.Set("userClaims", claims)

		// Continue to the next middleware or handler
		c.Next()
	}
}

// OptionalUserTokenVerification middleware verifies JWT tokens if present, but doesn't require them
func OptionalUserTokenVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without user context
			c.Next()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, continue without user context
			c.Next()
			return
		}

		// Extract the token from the header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			// Empty token, continue without user context
			c.Next()
			return
		}

		fmt.Println("this is token string ; ", tokenString)

		// Parse and validate the JWT token
		claims, err := utils.ParseJwt(tokenString)
		if err != nil {
			// Invalid token, continue without user context
			c.Next()
			return
		}

		fmt.Println("this is claims ; ", claims)

		// Validate that required claims are present
		if claims.UserId == "" || claims.Email == "" || claims.FullName == "" {
			// Invalid claims, continue without user context
			c.Next()
			return
		}

		// Store user information in the context for use in handlers
		c.Set("userId", claims.UserId)
		c.Set("userEmail", claims.Email)
		c.Set("userFullName", claims.FullName)
		c.Set("userClaims", claims)

		// Continue to the next middleware or handler
		c.Next()
	}
}

// GetUserFromContext extracts user information from the Gin context
func GetUserFromContext(c *gin.Context) (string, string, string, bool) {
	userId, exists := c.Get("userId")
	if !exists {
		return "", "", "", false
	}

	userEmail, exists := c.Get("userEmail")
	if !exists {
		return "", "", "", false
	}

	userFullName, exists := c.Get("userFullName")
	if !exists {
		return "", "", "", false
	}

	return userId.(string), userEmail.(string), userFullName.(string), true
}

// GetUserClaimsFromContext extracts the complete user claims from the Gin context
func GetUserClaimsFromContext(c *gin.Context) (*utils.Claims, bool) {
	claims, exists := c.Get("userClaims")
	if !exists {
		return nil, false
	}

	if userClaims, ok := claims.(*utils.Claims); ok {
		return userClaims, true
	}

	return nil, false
}
