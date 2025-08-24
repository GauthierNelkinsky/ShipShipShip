package models

import (
	"time"

	"gorm.io/gorm"
)

// FooterColumnType represents the column position in the footer
type FooterColumnType string

const (
	FooterColumnLeft   FooterColumnType = "left"
	FooterColumnMiddle FooterColumnType = "middle"
	FooterColumnRight  FooterColumnType = "right"
)

type FooterLink struct {
	ID              uint             `json:"id" gorm:"primaryKey"`
	Name            string           `json:"name" gorm:"not null"`
	URL             string           `json:"url" gorm:"not null"`
	Column          FooterColumnType `json:"column" gorm:"not null;default:'left'"`
	Order           int              `json:"order" gorm:"default:0"`
	OpenInNewWindow bool             `json:"open_in_new_window" gorm:"default:false"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `json:"-" gorm:"index"`
}

type CreateFooterLinkRequest struct {
	Name            string           `json:"name" binding:"required"`
	URL             string           `json:"url" binding:"required"`
	Column          FooterColumnType `json:"column" binding:"required"`
	OpenInNewWindow bool             `json:"open_in_new_window"`
}

type UpdateFooterLinkRequest struct {
	Name            *string           `json:"name"`
	URL             *string           `json:"url"`
	Column          *FooterColumnType `json:"column"`
	OpenInNewWindow *bool             `json:"open_in_new_window"`
}

type ReorderFooterLinksRequest struct {
	Links []struct {
		ID    uint `json:"id" binding:"required"`
		Order int  `json:"order" binding:"required"`
	} `json:"links" binding:"required"`
}

// GetFooterLinksByColumn returns footer links grouped by column
func GetFooterLinksByColumn(db *gorm.DB) (map[FooterColumnType][]FooterLink, error) {
	var links []FooterLink
	if err := db.Order("column ASC, `order` ASC, created_at ASC").Find(&links).Error; err != nil {
		return nil, err
	}

	result := make(map[FooterColumnType][]FooterLink)
	for _, link := range links {
		result[link.Column] = append(result[link.Column], link)
	}

	return result, nil
}

// GetFooterLinks returns all footer links ordered by column and order
func GetFooterLinks(db *gorm.DB) ([]FooterLink, error) {
	var links []FooterLink
	if err := db.Order("column ASC, `order` ASC, created_at ASC").Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

// GetNextOrder returns the next order value for a specific column
func GetNextOrder(db *gorm.DB, column FooterColumnType) (int, error) {
	var maxOrder int
	if err := db.Model(&FooterLink{}).Where("column = ?", column).Select("COALESCE(MAX(`order`), -1) + 1").Scan(&maxOrder).Error; err != nil {
		return 0, err
	}
	return maxOrder, nil
}
