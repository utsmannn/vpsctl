package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth returns an authentication middleware that validates bearer tokens
func Auth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Skip auth for OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_auth_format",
				"message": "Authorization header must be Bearer token",
			})
			c.Abort()
			return
		}

		// Validate token
		providedToken := parts[1]
		if providedToken != token {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Invalid authentication token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// authMiddleware is the actual middleware function used by the server
func authMiddleware(token string) gin.HandlerFunc {
	return Auth(token)
}
