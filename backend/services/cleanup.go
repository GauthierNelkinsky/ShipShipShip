package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	// Files older than this duration and not referenced will be deleted
	orphanedFileAge = 24 * time.Hour
	// How often to run the cleanup
	cleanupInterval = 6 * time.Hour
)

// CleanupService handles periodic cleanup of orphaned files
type CleanupService struct {
	db         *gorm.DB
	uploadsDir string
	stopChan   chan struct{}
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(db *gorm.DB, uploadsDir string) *CleanupService {
	return &CleanupService{
		db:         db,
		uploadsDir: uploadsDir,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the periodic cleanup process
func (cs *CleanupService) Start() {
	fmt.Println("Cleanup service started")

	// Run immediately on start
	cs.runCleanup()

	// Then run periodically
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cs.runCleanup()
			case <-cs.stopChan:
				ticker.Stop()
				fmt.Println("Cleanup service stopped")
				return
			}
		}
	}()
}

// Stop stops the cleanup service
func (cs *CleanupService) Stop() {
	close(cs.stopChan)
}

// runCleanup performs the actual cleanup operation
func (cs *CleanupService) runCleanup() {
	fmt.Println("Running orphaned file cleanup...")

	// Get all files in uploads directory
	files, err := os.ReadDir(cs.uploadsDir)
	if err != nil {
		fmt.Printf("Error reading uploads directory: %v\n", err)
		return
	}

	// Get all referenced image URLs from database
	referencedFiles := cs.getReferencedFiles()

	deletedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		filePath := filepath.Join(cs.uploadsDir, filename)

		// Get file info
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Error getting file info for %s: %v\n", filename, err)
			errorCount++
			continue
		}

		// Check if file is old enough to be considered orphaned
		fileAge := time.Since(fileInfo.ModTime())
		if fileAge < orphanedFileAge {
			skippedCount++
			continue // File is too new, might still be in use
		}

		// Check if file is referenced in any event
		if cs.isFileReferenced(filename, referencedFiles) {
			skippedCount++
			continue // File is still in use
		}

		// File is orphaned, delete it
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Error deleting orphaned file %s: %v\n", filename, err)
			errorCount++
		} else {
			fmt.Printf("Deleted orphaned file: %s (age: %v)\n", filename, fileAge.Round(time.Hour))
			deletedCount++
		}
	}

	fmt.Printf("Cleanup complete: %d deleted, %d kept, %d errors\n", deletedCount, skippedCount, errorCount)
}

// getReferencedFiles retrieves all filenames referenced in events
func (cs *CleanupService) getReferencedFiles() map[string]bool {
	referenced := make(map[string]bool)

	// Query to get all Media and Content fields from events
	var results []struct {
		Media   string
		Content string
	}

	if err := cs.db.Table("events").Select("media, content").Find(&results).Error; err != nil {
		fmt.Printf("Error querying events: %v\n", err)
		return referenced
	}

	// Extract filenames from Media JSON arrays
	for _, result := range results {
		// Parse Media field (JSON array of URLs)
		if result.Media != "" {
			filenames := cs.extractFilenamesFromMediaJSON(result.Media)
			for _, filename := range filenames {
				referenced[filename] = true
			}
		}

		// Parse Content field (HTML with img tags)
		if result.Content != "" {
			filenames := cs.extractFilenamesFromHTML(result.Content)
			for _, filename := range filenames {
				referenced[filename] = true
			}
		}
	}

	// Also check branding settings for logo/favicon
	var brandingResults []struct {
		Logo    string
		Favicon string
	}

	if err := cs.db.Table("branding_settings").Select("logo, favicon").Find(&brandingResults).Error; err == nil {
		for _, branding := range brandingResults {
			if branding.Logo != "" {
				if filename := cs.extractFilenameFromURL(branding.Logo); filename != "" {
					referenced[filename] = true
				}
			}
			if branding.Favicon != "" {
				if filename := cs.extractFilenameFromURL(branding.Favicon); filename != "" {
					referenced[filename] = true
				}
			}
		}
	}

	return referenced
}

// extractFilenamesFromMediaJSON extracts filenames from Media JSON array
func (cs *CleanupService) extractFilenamesFromMediaJSON(mediaJSON string) []string {
	var filenames []string

	// Simple parsing without json.Unmarshal for robustness
	// Look for patterns like "/api/uploads/filename.ext" or "/uploads/filename.ext"
	urls := cs.findUploadURLs(mediaJSON)
	for _, url := range urls {
		if filename := cs.extractFilenameFromURL(url); filename != "" {
			filenames = append(filenames, filename)
		}
	}

	return filenames
}

// extractFilenamesFromHTML extracts filenames from HTML content
func (cs *CleanupService) extractFilenamesFromHTML(html string) []string {
	var filenames []string

	urls := cs.findUploadURLs(html)
	for _, url := range urls {
		if filename := cs.extractFilenameFromURL(url); filename != "" {
			filenames = append(filenames, filename)
		}
	}

	return filenames
}

// findUploadURLs finds all upload URLs in a string
func (cs *CleanupService) findUploadURLs(text string) []string {
	var urls []string

	// Look for /api/uploads/ or /uploads/ patterns
	for _, prefix := range []string{"/api/uploads/", "/uploads/"} {
		index := 0
		for {
			pos := strings.Index(text[index:], prefix)
			if pos == -1 {
				break
			}

			start := index + pos
			end := start + len(prefix)

			// Find the end of the filename (space, quote, or bracket)
			for end < len(text) && !strings.ContainsRune(" \"'<>[](){}", rune(text[end])) {
				end++
			}

			if end > start+len(prefix) {
				urls = append(urls, text[start:end])
			}

			index = end
		}
	}

	return urls
}

// extractFilenameFromURL extracts just the filename from a URL
func (cs *CleanupService) extractFilenameFromURL(url string) string {
	// Remove any query parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Find the last occurrence of "uploads/"
	uploadsIndex := strings.LastIndex(url, "uploads/")
	if uploadsIndex == -1 {
		return ""
	}

	filename := url[uploadsIndex+8:] // 8 is len("uploads/")

	// Validate filename
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
		return ""
	}

	// Must have an extension
	if !strings.Contains(filename, ".") {
		return ""
	}

	return filename
}

// isFileReferenced checks if a filename is in the referenced files map
func (cs *CleanupService) isFileReferenced(filename string, referencedFiles map[string]bool) bool {
	return referencedFiles[filename]
}
