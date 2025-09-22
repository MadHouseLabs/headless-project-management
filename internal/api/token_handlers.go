package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/auth"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type TokenHandler struct {
	db *database.Database
}

func NewTokenHandler(db *database.Database) *TokenHandler {
	return &TokenHandler{
		db: db,
	}
}

type CreateAPITokenRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Scopes      string  `json:"scopes"`
	ExpiresIn   int     `json:"expires_in_days"` // Number of days until expiration
}

type APITokenResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Token       string     `json:"token,omitempty"` // Only returned on creation
	Description string     `json:"description"`
	Scopes      string     `json:"scopes"`
	CreatedBy   string     `json:"created_by"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateAPIToken creates a new API token (admin only)
func (h *TokenHandler) CreateAPIToken(c *gin.Context) {
	var req CreateAPITokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate the token
	token, err := models.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Create token record
	apiToken := &models.APIToken{
		Name:        req.Name,
		Description: req.Description,
		TokenHash:   auth.HashToken(token),
		Scopes:      req.Scopes,
		CreatedBy:   c.GetString("user_id"),
		IsActive:    true,
	}

	// Set expiration if specified
	if req.ExpiresIn > 0 {
		expiresAt := time.Now().AddDate(0, 0, req.ExpiresIn)
		apiToken.ExpiresAt = &expiresAt
	}

	// Default scopes if not specified
	if apiToken.Scopes == "" {
		apiToken.Scopes = "read,write"
	}

	if err := h.db.Create(apiToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Return the token (only time it's visible)
	c.JSON(http.StatusCreated, APITokenResponse{
		ID:          apiToken.ID,
		Name:        apiToken.Name,
		Token:       token, // Return the actual token only on creation
		Description: apiToken.Description,
		Scopes:      apiToken.Scopes,
		CreatedBy:   apiToken.CreatedBy,
		ExpiresAt:   apiToken.ExpiresAt,
		IsActive:    apiToken.IsActive,
		CreatedAt:   apiToken.CreatedAt,
	})
}

// ListAPITokens lists all API tokens (admin only)
func (h *TokenHandler) ListAPITokens(c *gin.Context) {
	var tokens []models.APIToken
	if err := h.db.Order("created_at DESC").Find(&tokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tokens"})
		return
	}

	// Convert to response format (without token values)
	response := make([]APITokenResponse, len(tokens))
	for i, token := range tokens {
		response[i] = APITokenResponse{
			ID:          token.ID,
			Name:        token.Name,
			Description: token.Description,
			Scopes:      token.Scopes,
			CreatedBy:   token.CreatedBy,
			ExpiresAt:   token.ExpiresAt,
			LastUsed:    token.LastUsed,
			IsActive:    token.IsActive,
			CreatedAt:   token.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// RevokeAPIToken revokes an API token (admin only)
func (h *TokenHandler) RevokeAPIToken(c *gin.Context) {
	tokenID := c.Param("id")

	var token models.APIToken
	if err := h.db.First(&token, tokenID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
		return
	}

	token.IsActive = false
	if err := h.db.Save(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}

// ValidateAPIToken validates the current token
func (h *TokenHandler) ValidateAPIToken(c *gin.Context) {
	tokenID, exists := c.Get("token_id")
	if !exists {
		// Using admin token
		c.JSON(http.StatusOK, gin.H{
			"valid": true,
			"user": c.GetString("user_id"),
			"scopes": c.GetString("scopes"),
			"is_admin": true,
		})
		return
	}

	var token models.APIToken
	if err := h.db.First(&token, tokenID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": token.IsValid(),
		"name": token.Name,
		"scopes": token.Scopes,
		"expires_at": token.ExpiresAt,
		"is_admin": token.HasScope("admin"),
	})
}

// GetAPIToken gets details of a specific token (admin only)
func (h *TokenHandler) GetAPIToken(c *gin.Context) {
	tokenID := c.Param("id")

	var token models.APIToken
	if err := h.db.First(&token, tokenID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
		return
	}

	c.JSON(http.StatusOK, APITokenResponse{
		ID:          token.ID,
		Name:        token.Name,
		Description: token.Description,
		Scopes:      token.Scopes,
		CreatedBy:   token.CreatedBy,
		ExpiresAt:   token.ExpiresAt,
		LastUsed:    token.LastUsed,
		IsActive:    token.IsActive,
		CreatedAt:   token.CreatedAt,
	})
}