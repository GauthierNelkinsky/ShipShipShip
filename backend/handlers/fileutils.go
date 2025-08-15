package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var uploadsDir = "./data/uploads"

// extractFilenameFromURL extracts the filename from an image URL
// Expected format: /api/uploads/filename.ext or /uploads/filename.ext
func extractFilenameFromURL(url string) string {
	if url == "" {
		return ""
	}

	// Check if it's a valid upload URL format
	if !isImageURL(url) {
		return ""
	}

	// Split the URL and ensure it follows the expected format
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}

	// Find the uploads part and ensure filename is directly after it
	uploadsIndex := -1
	for i, part := range parts {
		if part == "uploads" {
			uploadsIndex = i
			break
		}
	}

	// Ensure we found uploads and there's exactly one more part (the filename)
	if uploadsIndex == -1 || uploadsIndex+1 != len(parts)-1 {
		return ""
	}

	filename := parts[len(parts)-1]

	// Validate that this looks like a filename with extension
	if !strings.Contains(filename, ".") {
		return ""
	}

	// Additional security check - ensure no path traversal characters
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return ""
	}

	return filename
}

// deleteImageFile safely deletes an image file from the uploads directory
func deleteImageFile(filename string) error {
	if filename == "" {
		return nil
	}

	// Security check
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename: %s", filename)
	}

	filePath := filepath.Join(uploadsDir, filename)

	// Check if file exists before trying to delete
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, consider it already deleted
		return nil
	}

	return os.Remove(filePath)
}

// deleteImageFromURL deletes an image file using its URL
func deleteImageFromURL(url string) error {
	filename := extractFilenameFromURL(url)
	if filename == "" {
		return nil // Nothing to delete
	}

	return deleteImageFile(filename)
}

// cleanupMediaFiles deletes all image files referenced in a media JSON array
func cleanupMediaFiles(mediaJSON string) error {
	if mediaJSON == "" {
		return nil
	}

	var mediaURLs []string
	if err := json.Unmarshal([]byte(mediaJSON), &mediaURLs); err != nil {
		// If we can't parse the JSON, we can't clean up - log but don't fail
		fmt.Printf("Warning: Could not parse media JSON for cleanup: %v\n", err)
		return nil
	}

	var errors []string
	for _, url := range mediaURLs {
		if err := deleteImageFromURL(url); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete %s: %v", url, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// extractImagesFromContent finds all image URLs in TipTap HTML content
func extractImagesFromContent(content string) []string {
	if content == "" {
		return nil
	}

	var imageURLs []string

	// Simple regex to find img src attributes
	// This handles both <img src="/api/uploads/..." and <img src="/uploads/..."
	imgRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	matches := imgRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			url := match[1]
			if isImageURL(url) {
				imageURLs = append(imageURLs, url)
			}
		}
	}

	return imageURLs
}

// cleanupContentImages removes all uploaded images found in content
func cleanupContentImages(content string) error {
	imageURLs := extractImagesFromContent(content)

	var errors []string
	for _, url := range imageURLs {
		if err := deleteImageFromURL(url); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete %s: %v", url, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("content cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// isImageURL checks if a URL appears to be an image upload URL
func isImageURL(url string) bool {
	if url == "" {
		return false
	}

	// Check if it's one of our upload URLs
	return strings.Contains(url, "/api/uploads/") || strings.Contains(url, "/uploads/")
}
