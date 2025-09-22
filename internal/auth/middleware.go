package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

// HashToken creates a SHA-256 hash of the token
func HashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// AuthMiddleware creates a middleware that validates API tokens
func AuthMiddleware(db *database.Database, requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Try X-API-Key header as fallback
			authHeader = c.GetHeader("X-API-Key")
		}

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authentication token",
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		// Check if it's the admin token from environment
		adminToken := os.Getenv("ADMIN_API_TOKEN")
		if adminToken != "" && token == adminToken {
			// Admin has full access
			c.Set("user_id", "admin")
			c.Set("is_admin", true)
			c.Set("scopes", "*")
			c.Next()
			return
		}

		// Look up token in database
		var apiToken models.APIToken
		tokenHash := HashToken(token)
		err := db.Where("token_hash = ?", tokenHash).First(&apiToken).Error
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Check if token has expired
		if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is expired",
			})
			c.Abort()
			return
		}

		// Check required scopes (simplified - just check if scope field contains required scope)
		for _, scope := range requiredScopes {
			if !strings.Contains(apiToken.Scope, scope) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Insufficient permissions",
					"required_scope": scope,
				})
				c.Abort()
				return
			}
		}

		// Update last used timestamp
		now := time.Now()
		apiToken.LastUsed = &now
		db.Save(&apiToken)

		// Set user context
		c.Set("token_id", apiToken.ID)
		c.Set("user_id", apiToken.UserID)
		c.Set("scopes", apiToken.Scope)
		c.Set("is_admin", strings.Contains(apiToken.Scope, "admin"))

		c.Next()
	}
}

// AdminOnly middleware ensures only admin tokens can access
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// OptionalAuth allows requests with or without authentication
func OptionalAuth(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.GetHeader("X-API-Key")
		}

		if authHeader != "" {
			// Try to authenticate but don't fail if token is invalid
			token := strings.TrimPrefix(authHeader, "Bearer ")
			token = strings.TrimSpace(token)

			// Check admin token
			adminToken := os.Getenv("ADMIN_API_TOKEN")
			if adminToken != "" && token == adminToken {
				c.Set("authenticated", true)
				c.Set("user_id", "admin")
				c.Set("is_admin", true)
				c.Next()
				return
			}

			// Try database token
			var apiToken models.APIToken
			tokenHash := HashToken(token)
			err := db.Where("token_hash = ?", tokenHash).First(&apiToken).Error
			if err == nil && (apiToken.ExpiresAt == nil || apiToken.ExpiresAt.After(time.Now())) {
				c.Set("authenticated", true)
				c.Set("token_id", apiToken.ID)
				c.Set("user_id", apiToken.UserID)
				c.Set("scopes", apiToken.Scope)
			}
		}

		c.Set("authenticated", false)
		c.Next()
	}
}