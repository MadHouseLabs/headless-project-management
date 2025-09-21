package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
	"github.com/headless-pm/headless-project-management/pkg/auth"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *database.Database
	jwtManager *auth.JWTManager
}

func NewAuthHandler(db *database.Database, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jwtManager,
	}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=30"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := &models.User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      models.UserRoleMember,
	}

	if err := h.db.Create(user).Error; err != nil {
		if err == gorm.ErrDuplicatedKey {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	accessToken, err := h.jwtManager.Generate(user.ID, user.Email, user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
		"tokens": TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := auth.CheckPassword(req.Password, user.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled"})
		return
	}

	accessToken, err := h.jwtManager.Generate(user.ID, user.Email, user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	h.db.Model(&user).Update("last_login", time.Now())

	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"tokens": TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		},
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.jwtManager.Verify(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	var user models.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled"})
		return
	}

	accessToken, err := h.jwtManager.Generate(user.ID, user.Email, user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   3600,
		TokenType:   "Bearer",
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var user models.User
	if err := h.db.Preload("Teams").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Avatar    string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"avatar":     req.Avatar,
		"updated_at": time.Now(),
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	var user models.User
	h.db.First(&user, userID)
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := auth.CheckPassword(req.CurrentPassword, user.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := h.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}