package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func ExtractToken(c *gin.Context, cookieName string) string {
	token, err := c.Cookie(cookieName)
	if err != nil || token == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}
	return token
}
