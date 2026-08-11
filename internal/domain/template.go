package domain

import (
	"fmt"
	"strings"
	"time"
)

type EventType string

const (
	EventTypeDonation   EventType = "donation"
	EventTypeSubscriber EventType = "subscriber"
	EventTypeFollower   EventType = "follower"
	EventTypeCheer      EventType = "cheer"
	EventTypeCustom     EventType = "custom"
)

func (e EventType) IsValid() bool {
	switch e {
	case EventTypeDonation, EventTypeSubscriber, EventTypeFollower, EventTypeCheer, EventTypeCustom:
		return true
	}
	return false
}

func ParseEventType(s string) (EventType, error) {
	clean := EventType(strings.TrimSpace(strings.ToLower(s)))
	if !clean.IsValid() {
		return "", fmt.Errorf("invalid event_type '%s'. Valid types are: donation, subscriber, follower, cheer, custom", s)
	}
	return clean, nil
}

type Template struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	ProjectID    uint64    `gorm:"index;not null;column:project_id" json:"project_id" example:"1"`
	Name         string    `gorm:"not null;type:varchar(150)" json:"name" example:"New Donation Alert"`
	EventType    EventType `gorm:"not null;type:varchar(50)" json:"event_type" example:"donation"`
	HTMLTemplate string    `gorm:"not null;type:text" json:"html_template" example:"<div class='alert'><h1>{{name}} donated {{amount}}</h1><p>{{description}}</p></div>"`
	CSSStyle     string    `gorm:"type:text" json:"css_style,omitempty" example:".alert { background: gold; color: black; }"`
	Fields       string    `gorm:"not null;type:text" json:"fields" example:"[\"name\",\"amount\",\"description\"]"`
	Duration     int       `gorm:"default:5000" json:"duration" example:"5000"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type CreateTemplateRequest struct {
	Name         string   `json:"name" example:"New Donation Alert" binding:"required"`
	EventType    string   `json:"event_type" example:"donation" binding:"required"`
	HTMLTemplate string   `json:"html_template" example:"<h1>{{name}} donated {{amount}}</h1><p>{{description}}</p>" binding:"required"`
	CSSStyle     string   `json:"css_style,omitempty"`
	Fields       []string `json:"fields" example:"[\"name\",\"amount\",\"description\"]"`
	Duration     int      `json:"duration,omitempty" example:"5000"`
}

type TriggerCustomAlertRequest struct {
	Payload  map[string]string `json:"payload" binding:"required"`
	Duration int               `json:"duration,omitempty" example:"5000"`
}
