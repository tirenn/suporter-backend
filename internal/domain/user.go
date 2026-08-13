package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	Username     string    `gorm:"uniqueIndex;not null;type:varchar(50)" json:"username" example:"johndoe"`
	PasswordHash string    `gorm:"not null;column:password_hash" json:"-"`
	Name         string    `gorm:"not null;type:varchar(100)" json:"name" example:"John Doe"`
	Role         string    `gorm:"not null;type:varchar(20);default:'streamer'" json:"role" example:"streamer"`
	WebhookKey   string    `gorm:"uniqueIndex;type:varchar(64)" json:"webhook_key,omitempty" example:"wk_a1b2c3d4e5f6..."`
	QRISUrl      string    `gorm:"type:text;column:qris_url" json:"qris_url,omitempty" example:"https://i.imgur.com/abc.jpg"`
	IsActive     bool      `gorm:"not null;default:false" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// StreamerPublicProfile is the safe public-facing subset returned to unauthenticated donors.
type StreamerPublicProfile struct {
	Name     string `json:"name" example:"John Doe"`
	Username string `json:"username" example:"johndoe"`
	QRISUrl  string `json:"qris_url" example:"https://i.imgur.com/abc.jpg"`
}

type RegisterRequest struct {
	Name           string `json:"name" example:"John Doe" binding:"required,min=3,max=100"`
	Username       string `json:"username" example:"johndoe" binding:"required,min=3,max=50"`
	Password       string `json:"password" example:"secret123" binding:"required,min=6"`
	Role           string `json:"role" example:"streamer" binding:"required,oneof=streamer"`
	RecaptchaToken string `json:"recaptcha_token" binding:"required"`
}

type LoginRequest struct {
	Username       string `json:"username" example:"johndoe" binding:"required"`
	Password       string `json:"password" example:"secret123" binding:"required"`
	RecaptchaToken string `json:"recaptcha_token" binding:"required"`
}

// UpdateProfileRequest lets a streamer update their QRIS QR code URL.
type UpdateProfileRequest struct {
	QRISUrl string `json:"qris_url" binding:"required"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1Ni..."`
	User        User   `json:"user"`
}

type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
