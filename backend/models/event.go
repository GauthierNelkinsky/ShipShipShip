package models

import (
	"time"

	"gorm.io/gorm"
)

type EventStatus string

const (
	StatusBacklogs EventStatus = "Backlogs"
	StatusDoing    EventStatus = "Doing"
	StatusRelease  EventStatus = "Release"
	StatusUpcoming EventStatus = "Upcoming"
	StatusArchived EventStatus = "Archived"
)

type Event struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"not null"`
	Tags      string         `json:"tags"`  // JSON string of array
	Media     string         `json:"media"` // JSON string of array
	Status    EventStatus    `json:"status" gorm:"not null"`
	Date      string         `json:"date"` // Can be "Q3 2025", "12.Jul.2025", or formatted date
	Votes     int            `json:"votes" gorm:"default:0"`
	Content   string         `json:"content"`                                  // Markdown content
	Order     int            `json:"order" gorm:"column:sort_order;default:0"` // Order for sorting within status
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateEventRequest struct {
	Title   string      `json:"title" binding:"required"`
	Tags    []string    `json:"tags"`
	Media   []string    `json:"media"`
	Status  EventStatus `json:"status" binding:"required"`
	Date    string      `json:"date"`
	Content string      `json:"content"`
	Order   *int        `json:"order"`
}

type UpdateEventRequest struct {
	Title   *string      `json:"title"`
	Tags    []string     `json:"tags"`
	Media   []string     `json:"media"`
	Status  *EventStatus `json:"status"`
	Date    *string      `json:"date"`
	Content *string      `json:"content"`
	Order   *int         `json:"order"`
}

type VoteRequest struct {
	EventID uint `json:"event_id" binding:"required"`
}
