package domain

import "time"

type Donation struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	StreamerID  uint64    `gorm:"index;not null;column:streamer_id" json:"streamer_id" example:"1"`
	SenderName  string    `gorm:"not null;type:varchar(100)" json:"sender_name" example:"Jane Donor"`
	Amount      int64     `gorm:"not null" json:"amount" example:"50000"`
	UniqueCode  int       `gorm:"not null" json:"unique_code" example:"142"`
	TotalAmount int64     `gorm:"index;not null" json:"total_amount" example:"50142"`
	Message     string    `gorm:"type:text" json:"message" example:"Keep up the great stream!"`
	Status      string    `gorm:"not null;type:varchar(20);default:'pending'" json:"status" example:"pending"`
	IsTest      bool      `gorm:"not null;default:false" json:"is_test" example:"false"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CreateDonationRequest struct {
	StreamerUsername string `json:"streamer_username" example:"streamer123" binding:"required"`
	SenderName       string `json:"sender_name" example:"Jane Donor" binding:"required"`
	Amount           int64  `json:"amount" example:"50000" binding:"required,min=5000,max=10000000"`
	Message          string `json:"message" example:"Keep up the great stream!"`
	RecaptchaToken   string `json:"recaptcha_token" binding:"required"`
}

type WebhookDonationRequest struct {
	Amount int64 `json:"amount" example:"50142" binding:"required"`
}
