package models

import (
	"time"

	"gorm.io/gorm"
)

// ReactionType represents the type of reaction
type ReactionType string

const (
	ReactionThumbsUp   ReactionType = "thumbs_up"   // 👍 Like/Support
	ReactionHeart      ReactionType = "heart"       // ❤️ Love it
	ReactionFire       ReactionType = "fire"        // 🔥 Hot/Urgent
	ReactionParty      ReactionType = "party"       // 🎉 Excited/Celebrating
	ReactionEyes       ReactionType = "eyes"        // 👀 Watching/Interested
	ReactionLightBulb  ReactionType = "lightbulb"   // 💡 Great idea
	ReactionThinking   ReactionType = "thinking"    // 🤔 Needs discussion
	ReactionThumbsDown ReactionType = "thumbs_down" // 👎 Not needed/Disagree
)

// EventReaction represents a user's reaction to an event
type EventReaction struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	EventID      uint           `json:"event_id" gorm:"not null;index"`
	ReactionType ReactionType   `json:"reaction_type" gorm:"not null;index"`
	IPAddress    string         `json:"ip_address" gorm:"index"`
	UserID       *uint          `json:"user_id" gorm:"index"` // nullable for anonymous reactions
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationship
	Event Event `json:"-" gorm:"foreignKey:EventID"`
}

// TableName sets the table name for the EventReaction model
func (EventReaction) TableName() string {
	return "event_reactions"
}

// ReactionCount represents aggregated reaction counts
type ReactionCount struct {
	ReactionType ReactionType `json:"reaction_type"`
	Count        int64        `json:"count"`
}

// ReactionSummary represents a summary of all reactions for an event
type ReactionSummary struct {
	EventID       uint            `json:"event_id"`
	TotalCount    int64           `json:"total_count"`
	Reactions     []ReactionCount `json:"reactions"`
	UserReactions []ReactionType  `json:"user_reactions"` // reactions by current user/IP
}

// GetReactionEmoji returns the emoji for a reaction type
func GetReactionEmoji(reactionType ReactionType) string {
	switch reactionType {
	case ReactionThumbsUp:
		return "👍"
	case ReactionHeart:
		return "❤️"
	case ReactionFire:
		return "🔥"
	case ReactionParty:
		return "🎉"
	case ReactionEyes:
		return "👀"
	case ReactionLightBulb:
		return "💡"
	case ReactionThinking:
		return "🤔"
	case ReactionThumbsDown:
		return "👎"
	default:
		return ""
	}
}

// GetReactionLabel returns a human-readable label for a reaction type
func GetReactionLabel(reactionType ReactionType) string {
	switch reactionType {
	case ReactionThumbsUp:
		return "Like"
	case ReactionHeart:
		return "Love"
	case ReactionFire:
		return "Hot"
	case ReactionParty:
		return "Excited"
	case ReactionEyes:
		return "Watching"
	case ReactionLightBulb:
		return "Great Idea"
	case ReactionThinking:
		return "Thinking"
	case ReactionThumbsDown:
		return "Disagree"
	default:
		return string(reactionType)
	}
}

// ValidReactionTypes returns all valid reaction types
func ValidReactionTypes() []ReactionType {
	return []ReactionType{
		ReactionThumbsUp,
		ReactionHeart,
		ReactionFire,
		ReactionParty,
		ReactionEyes,
		ReactionLightBulb,
		ReactionThinking,
		ReactionThumbsDown,
	}
}

// IsValidReactionType checks if a reaction type is valid
func IsValidReactionType(rt ReactionType) bool {
	for _, validType := range ValidReactionTypes() {
		if rt == validType {
			return true
		}
	}
	return false
}
