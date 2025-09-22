package models

import (
	"time"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"unique;not null"`
	Username     string    `json:"username" gorm:"unique;not null"`
	Password     string    `json:"-" gorm:"not null"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Avatar       string    `json:"avatar"`
	Role         UserRole  `json:"role" gorm:"default:'member'"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	OwnedProjects []Project `json:"owned_projects,omitempty" gorm:"foreignKey:OwnerID"`
}

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleMember  UserRole = "member"
	UserRoleViewer  UserRole = "viewer"
)

type Session struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type APIToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	Scope     string    `json:"scope"` // e.g., "read", "write", "admin"
	ExpiresAt *time.Time `json:"expires_at"` // nil means no expiration
	LastUsed  *time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}