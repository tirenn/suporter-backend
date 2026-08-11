package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	Email        string    `gorm:"uniqueIndex;not null;type:varchar(255)" json:"email" example:"user@example.com"`
	PasswordHash string    `gorm:"not null;column:password_hash" json:"-"`
	Name         string    `gorm:"not null;type:varchar(100)" json:"name" example:"John Doe"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type RegisterRequest struct {
	Name     string `json:"name" example:"John Doe" binding:"required"`
	Email    string `json:"email" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"secret123" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"secret123" binding:"required"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1Ni..."`
	User        User   `json:"user"`
}

type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
