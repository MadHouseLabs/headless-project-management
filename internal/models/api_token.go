package models

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"gorm.io/gorm"
)

type APIToken struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name"`
	Token       string         `gorm:"uniqueIndex" json:"token,omitempty"`
	TokenHash   string         `gorm:"uniqueIndex" json:"-"` // Store hashed token
	Description string         `json:"description"`
	CreatedBy   string         `json:"created_by"`
	LastUsed    *time.Time     `json:"last_used,omitempty"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Scopes      string         `json:"scopes"` // Comma-separated list of scopes
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// GenerateToken creates a new random token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HasScope checks if the token has a specific scope
func (t *APIToken) HasScope(scope string) bool {
	if t.Scopes == "*" || t.Scopes == "admin" {
		return true
	}
	scopes := strings.Split(t.Scopes, ",")
	for _, s := range scopes {
		if strings.TrimSpace(s) == scope {
			return true
		}
	}
	return false
}

// IsExpired checks if the token has expired
func (t *APIToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// IsValid checks if the token is valid for use
func (t *APIToken) IsValid() bool {
	return t.IsActive && !t.IsExpired()
}