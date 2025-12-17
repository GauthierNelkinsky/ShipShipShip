package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxFileSize = 10 << 20 // 10MB
)

// Initialize uploads directory
func init() {
	if err := os.MkdirAll("./data", 0755); err != nil {
		panic(fmt.Sprintf("Failed to create data directory: %v", err))
	}
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create uploads directory: %v", err))
	}
}

// UploadImage handles image uploads
func UploadImage(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large or invalid form data"})
		return
	}

	// Get the file from form data
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB limit"})
		return
	}

	// Read the first 512 bytes to detect file type by magic numbers
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	buffer = buffer[:n]

	// Validate file type by magic numbers (actual file content)
	detectedType := detectImageType(buffer)
	if detectedType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPEG, PNG, GIF, WebP, and ICO are allowed"})
		return
	}

	// Reset file pointer to beginning for copying
	if _, err := file.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}

	// Validate Content-Type header as additional check
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Content-Type header"})
		return
	}

	// Generate unique filename
	ext := getFileExtension(header.Filename)
	if ext == "" {
		ext = getExtensionFromMimeType(contentType)
	}

	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // Clean up on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return the URL for the uploaded image
	imageURL := fmt.Sprintf("/api/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      imageURL,
		"filename": filename,
		"size":     header.Size,
	})
}

// ServeUploadedFile serves uploaded files
func ServeUploadedFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	// Validate filename to prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}

	filePath := filepath.Join(uploadsDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Set appropriate content type
	ext := filepath.Ext(filename)
	contentType := getMimeTypeFromExtension(ext)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year

	c.File(filePath)
}

// Helper functions

// detectImageType validates file type by checking magic numbers (file signatures)
func detectImageType(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return "image/png"
	}

	// GIF: GIF87a or GIF89a
	if len(data) >= 6 && bytes.Equal(data[0:3], []byte("GIF")) {
		if bytes.Equal(data[3:6], []byte("87a")) || bytes.Equal(data[3:6], []byte("89a")) {
			return "image/gif"
		}
	}

	// WebP: RIFF....WEBP
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp"
	}

	// ICO: 00 00 01 00
	if len(data) >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 && data[3] == 0x00 {
		return "image/x-icon"
	}

	return ""
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/x-icon",
		"image/vnd.microsoft.icon",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

func getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	default:
		return ""
	}
}

func getMimeTypeFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
