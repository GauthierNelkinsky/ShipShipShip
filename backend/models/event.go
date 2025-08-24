package models

import (
	"time"

	"gorm.io/gorm"
)

type EventStatus string

const (
	StatusBacklogs EventStatus = "Backlogs"
	StatusProposed EventStatus = "Proposed"
	StatusUpcoming EventStatus = "Upcoming"
	StatusRelease  EventStatus = "Release"
	StatusArchived EventStatus = "Archived"
)

type Tag struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex"`
	Color     string    `json:"color" gorm:"not null;default:#3B82F6"` // Default blue color
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Events    []Event   `json:"-" gorm:"many2many:event_tags;"`
}

type Event struct {
	ID           uint              `json:"id" gorm:"primaryKey"`
	Title        string            `json:"title" gorm:"not null"`
	Slug         string            `json:"slug" gorm:"uniqueIndex"`
	Tags         []Tag             `json:"tags" gorm:"many2many:event_tags;"`
	Media        string            `json:"media"` // JSON string of array
	Status       EventStatus       `json:"status" gorm:"not null"`
	Date         string            `json:"date"` // Can be "Q3 2025", "12.Jul.2025", or formatted date
	Votes        int               `json:"votes" gorm:"default:0"`
	Content      string            `json:"content"`                                  // Markdown content
	Order        int               `json:"order" gorm:"column:sort_order;default:0"` // Order for sorting within status
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	DeletedAt    gorm.DeletedAt    `json:"-" gorm:"index"`
	IsPublic     bool              `json:"is_public" gorm:"default:true"`      // Controls if event appears on public page
	HasPublicUrl bool              `json:"has_public_url" gorm:"default:true"` // Controls if event has individual public URL
	Publication  *EventPublication `json:"publication,omitempty" gorm:"foreignKey:EventID"`
}

type EventPublication struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	EventID         uint       `json:"event_id" gorm:"not null;uniqueIndex"`
	EmailSent       bool       `json:"email_sent" gorm:"default:false"`
	EmailSubject    string     `json:"email_subject"`
	EmailContent    string     `json:"email_content"`
	EmailTemplate   string     `json:"email_template"` // "upcoming_feature" or "new_release"
	EmailSentAt     *time.Time `json:"email_sent_at"`
	SubscriberCount int        `json:"subscriber_count" gorm:"default:0"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// EventEmailHistory tracks the history of all emails sent for an event
type EventEmailHistory struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	EventID         uint      `json:"event_id" gorm:"not null;index"`
	EventStatus     string    `json:"event_status"`
	EmailSubject    string    `json:"email_subject"`
	EmailTemplate   string    `json:"email_template"` // "upcoming_feature" or "new_release"
	SubscriberCount int       `json:"subscriber_count" gorm:"default:0"`
	SentAt          time.Time `json:"sent_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required"`
}

type UpdateTagRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

type CreateEventRequest struct {
	Title   string      `json:"title" binding:"required"`
	TagIDs  []uint      `json:"tag_ids"` // Array of tag IDs instead of strings
	Media   []string    `json:"media"`
	Status  EventStatus `json:"status" binding:"required"`
	Date    string      `json:"date"`
	Content string      `json:"content"`
	Order   *int        `json:"order"`
}

type UpdateEventRequest struct {
	Title   *string      `json:"title"`
	TagIDs  *[]uint      `json:"tag_ids"` // Pointer to array of tag IDs to distinguish nil from empty
	Media   []string     `json:"media"`
	Status  *EventStatus `json:"status"`
	Date    *string      `json:"date"`
	Content *string      `json:"content"`
	Order   *int         `json:"order"`
}

type VoteRequest struct {
	EventID uint `json:"event_id" binding:"required"`
}

type EventPublishRequest struct {
	IsPublic     *bool `json:"is_public"`
	HasPublicUrl *bool `json:"has_public_url"`
}

type EventNewsletterRequest struct {
	Subject  string `json:"subject" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Template string `json:"template" binding:"required"`
}
