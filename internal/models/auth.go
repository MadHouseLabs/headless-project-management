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
	Teams        []Team    `json:"teams,omitempty" gorm:"many2many:team_members;"`
	OwnedProjects []Project `json:"owned_projects,omitempty" gorm:"foreignKey:OwnerID"`
}

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleMember  UserRole = "member"
	UserRoleViewer  UserRole = "viewer"
)

type Team struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	OwnerID     uint      `json:"owner_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Owner       *User     `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Members     []User    `json:"members,omitempty" gorm:"many2many:team_members;"`
	Projects    []Project `json:"projects,omitempty" gorm:"foreignKey:TeamID"`
}

type TeamMember struct {
	TeamID    uint         `json:"team_id" gorm:"primaryKey"`
	UserID    uint         `json:"user_id" gorm:"primaryKey"`
	Role      TeamRole     `json:"role" gorm:"default:'member'"`
	JoinedAt  time.Time    `json:"joined_at"`
}

type TeamRole string

const (
	TeamRoleOwner   TeamRole = "owner"
	TeamRoleAdmin   TeamRole = "admin"
	TeamRoleMember  TeamRole = "member"
	TeamRoleGuest   TeamRole = "guest"
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