package models

import (
	"time"

	"gorm.io/gorm"
)

type Vote struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	EventID   uint           `json:"event_id" gorm:"not null;index"`
	IPAddress string         `json:"ip_address" gorm:"not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationship
	Event Event `json:"event" gorm:"foreignKey:EventID"`
}

// TableName sets the table name for the Vote model
func (Vote) TableName() string {
	return "votes"
}
