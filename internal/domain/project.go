package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	UUID         uuid.UUID `gorm:"uniqueIndex;not null;type:varchar(36)" json:"uuid" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3d0001"`
	UserID       uint64    `gorm:"index;not null;column:user_id" json:"user_id" example:"1"`
	Name         string    `gorm:"not null;type:varchar(150)" json:"name" example:"Main Stream Overlay"`
	Description  string    `gorm:"type:text" json:"description,omitempty" example:"Superchat donation alert overlay"`
	EventType    string    `gorm:"not null;type:varchar(50);default:'donation'" json:"event_type" example:"donation"`
	HTMLTemplate string    `gorm:"not null;type:text" json:"html_template"`
	CSSStyle     string    `gorm:"type:text" json:"css_style,omitempty"`
	Fields       string    `gorm:"not null;type:text" json:"fields" example:"[\"name\",\"amount\",\"message\"]"`
	Duration     int       `gorm:"default:7000" json:"duration" example:"7000"`
	OBSUrl       string    `gorm:"-" json:"obs_url" example:"http://localhost:8080/overlay/9b1deb4d-3b7d-4bad-9bdd-2b0d7b3d0001"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CreateProjectRequest struct {
	Name         string `json:"name" example:"My Stream Project" binding:"required"`
	Description  string `json:"description,omitempty" example:"Overlay for my gaming channel"`
	HTMLTemplate string `json:"html_template,omitempty"`
	CSSStyle     string `json:"css_style,omitempty"`
	Duration     int    `json:"duration,omitempty" example:"7000"`
}

type UpdateProjectRequest struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	HTMLTemplate string `json:"html_template,omitempty"`
	CSSStyle     string `json:"css_style,omitempty"`
	Duration     int    `json:"duration,omitempty"`
}

type TriggerAlertRequest struct {
	Name     string `json:"name" example:"Alex" binding:"required"`
	Amount   string `json:"amount" example:"$50.00" binding:"required"`
	Message  string `json:"message" example:"Keep up the awesome stream!" binding:"required"`
	Duration int    `json:"duration,omitempty" example:"7000"`
}
