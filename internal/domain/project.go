package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	UUID        uuid.UUID `gorm:"uniqueIndex;not null;type:varchar(36)" json:"uuid" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3d0001"`
	UserID      uint64    `gorm:"index;not null;column:user_id" json:"user_id" example:"1"`
	Name        string    `gorm:"not null;type:varchar(150)" json:"name" example:"Main Stream"`
	Description string    `gorm:"type:text" json:"description,omitempty" example:"Stream overlay for YouTube"`
	OBSUrl      string    `gorm:"-" json:"obs_url" example:"http://localhost:8080/overlay/9b1deb4d-3b7d-4bad-9bdd-2b0d7b3d0001"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CreateProjectRequest struct {
	Name        string `json:"name" example:"My Stream Project" binding:"required"`
	Description string `json:"description" example:"Overlay for my gaming channel"`
}
