package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"shipshipship/database"
	"shipshipship/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ApplyThemeRequest struct {
	ThemeID      string `json:"themeId" binding:"required"`
	ThemeVersion string `json:"themeVersion" binding:"required"`
	BuildFileURL string `json:"buildFileUrl" binding:"required"`
}

type ThemeStoreTheme struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	Version          string `json:"version"`
	BuildFile        string `json:"build_file"`
	SubmissionStatus string `json:"submission_status"`
}

type ThemeStoreResponse struct {
	Items []ThemeStoreTheme `json:"items"`
}

type ApplyThemeResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	IsUpdate   bool   `json:"isUpdate"`
	OldVersion string `json:"oldVersion,omitempty"`
	NewVersion string `json:"newVersion"`
}

// ApplyTheme downloads a theme ZIP file and extracts it to replace the admin build
func ApplyTheme(c *gin.Context) {
	var req ApplyThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.ThemeID == "" || req.ThemeVersion == "" || req.BuildFileURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme ID, version, and build file URL are required"})
		return
	}

	// Download the theme ZIP file
	tempFile, err := downloadThemeFile(req.BuildFileURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download theme file", "details": err.Error()})
		return
	}
	defer os.Remove(tempFile) // Clean up temp file

	// Create backup of current theme build
	// Create backup of current theme
	backupDir := "./data/themes/backup"
	if err := backupCurrentTheme(backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to backup current theme", "details": err.Error()})
		return
	}

	// Extract the new theme (this will remove the previous theme)
	themeDir := "./data/themes/current"
	if err := extractTheme(tempFile, themeDir); err != nil {
		// Restore backup on failure
		restoreThemeBackup(backupDir, themeDir)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract theme", "details": err.Error()})
		return
	}

	// Clean up backup after successful application
	os.RemoveAll(backupDir)
	// Check if this is an update or new application
	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	isUpdate := false
	oldVersion := ""

	if err != nil {
		// Theme was applied but we couldn't update settings - log but don't fail
		fmt.Printf("Warning: Theme applied but couldn't update settings: %v\n", err)
	} else {
		// Check if we're updating an existing theme
		if settings.CurrentThemeID == req.ThemeID && settings.CurrentThemeVersion != "" {
			isUpdate = true
			oldVersion = settings.CurrentThemeVersion
		}

		// Update theme ID and version
		settings.CurrentThemeID = req.ThemeID
		settings.CurrentThemeVersion = req.ThemeVersion
		if err := db.Save(settings).Error; err != nil {
			fmt.Printf("Warning: Theme applied but couldn't save theme info: %v\n", err)
		}
	}

	// Clean up backup after successful application
	os.RemoveAll(backupDir)

	message := "Theme applied successfully"
	if isUpdate {
		message = fmt.Sprintf("Theme updated successfully from %s to %s", oldVersion, req.ThemeVersion)
	}

	c.JSON(http.StatusOK, ApplyThemeResponse{
		Success:    true,
		Message:    message,
		IsUpdate:   isUpdate,
		OldVersion: oldVersion,
		NewVersion: req.ThemeVersion,
	})
}

// GetCurrentTheme returns the currently applied theme ID and version
func GetCurrentTheme(c *gin.Context) {
	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current theme", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currentThemeId":      settings.CurrentThemeID,
		"currentThemeVersion": settings.CurrentThemeVersion,
	})
}

// GetThemeInfo returns detailed information about the current theme installation
func GetThemeInfo(c *gin.Context) {
	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings", "details": err.Error()})
		return
	}

	// Get theme directory info
	themeInfo := listInstalledThemes()

	// Add database info
	themeInfo["database"] = map[string]interface{}{
		"currentThemeId":      settings.CurrentThemeID,
		"currentThemeVersion": settings.CurrentThemeVersion,
	}

	// Add storage path info
	themeInfo["paths"] = map[string]interface{}{
		"themesDirectory": "./data/themes",
		"currentTheme":    "./data/themes/current",
		"backupTheme":     "./data/themes/backup",
	}

	c.JSON(http.StatusOK, themeInfo)
}

// downloadThemeFile downloads a file from URL and saves it to a temporary file
func downloadThemeFile(url string) (string, error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "theme-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	// Copy the response body to the temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return tempFile.Name(), nil
}

// backupCurrentTheme creates a backup of the current theme directory
func backupCurrentTheme(backupDir string) error {
	themeDir := "./data/themes/current"

	// Remove existing backup
	os.RemoveAll(backupDir)

	// Check if theme directory exists
	if _, err := os.Stat(themeDir); os.IsNotExist(err) {
		// No existing theme to backup
		return nil
	}

	// Ensure backup directory parent exists
	if err := os.MkdirAll("./data/themes", 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	// Copy theme directory to backup
	return copyDir(themeDir, backupDir)
}

// restoreThemeBackup restores the backup to the theme directory
func restoreThemeBackup(backupDir, themeDir string) error {
	// Remove current theme
	os.RemoveAll(themeDir)

	// Restore from backup
	return copyDir(backupDir, themeDir)
}

// extractTheme extracts a ZIP file to the target directory
func extractTheme(zipFile, targetDir string) error {
	// Remove existing theme directory completely to ensure clean installation
	fmt.Printf("Removing previous theme from %s\n", targetDir)
	os.RemoveAll(targetDir)

	// Ensure parent themes directory exists
	if err := os.MkdirAll("./data/themes", 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	// Create temporary extraction directory
	tempExtractDir := targetDir + "_temp"
	os.RemoveAll(tempExtractDir)
	if err := os.MkdirAll(tempExtractDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp extraction directory: %w", err)
	}
	defer os.RemoveAll(tempExtractDir)

	// Open ZIP file
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	// Extract files to temp directory
	for _, file := range reader.File {
		path := filepath.Join(tempExtractDir, file.Name)

		// Ensure the file path is within the temp directory (security check)
		if !strings.HasPrefix(path, filepath.Clean(tempExtractDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in ZIP: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			// Create directory
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		// Create file directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		if err := extractFile(file, path); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	// Find build directory in extracted files
	buildDir, err := findBuildDirectory(tempExtractDir)
	if err != nil {
		return fmt.Errorf("failed to find build directory: %w", err)
	}

	// Create final theme directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}

	// Copy build contents directly to theme directory
	if err := copyDir(buildDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy build directory: %w", err)
	}

	fmt.Printf("Theme extracted successfully to %s\n", targetDir)
	return nil
}

// extractFile extracts a single file from ZIP
func extractFile(file *zip.File, destPath string) error {
	// Open file in ZIP
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create destination file
	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy file contents
	_, err = io.Copy(outFile, rc)
	return err
}

// findBuildDirectory finds the build directory in the extracted theme
func findBuildDirectory(rootDir string) (string, error) {
	var buildDir string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && (info.Name() == "build" || info.Name() == "dist") {
			// Check if this directory contains typical build files
			if hasTypicalBuildFiles(path) {
				buildDir = path
				return filepath.SkipDir // Stop walking this branch
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if buildDir == "" {
		// If no build directory found, check if root contains build files
		if hasTypicalBuildFiles(rootDir) {
			return rootDir, nil
		}
		return "", fmt.Errorf("no build directory found in theme package")
	}

	return buildDir, nil
}

// hasTypicalBuildFiles checks if a directory contains typical build files
func hasTypicalBuildFiles(dir string) bool {
	// Check for typical build files/directories
	expectedItems := []string{"index.html", "_app", "assets"}

	for _, item := range expectedItems {
		itemPath := filepath.Join(dir, item)
		if _, err := os.Stat(itemPath); err == nil {
			return true
		}
	}

	return false
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// InitializeDefaultTheme fetches and installs the default theme from Theme store if no theme is currently applied
func InitializeDefaultTheme() error {
	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	// Check if a theme is already applied
	if settings.CurrentThemeID != "" {
		fmt.Println("Theme already applied, skipping default theme initialization")
		return nil
	}

	// Check if theme files already exist
	if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
		fmt.Println("Theme files already exist, skipping default theme initialization")
		return nil
	}

	fmt.Println("No theme applied, initializing default theme...")

	// Fetch default theme from Theme store
	defaultTheme, err := fetchDefaultThemeFromThemeStore()
	if err != nil {
		return fmt.Errorf("failed to fetch default theme: %w", err)
	}

	if defaultTheme == nil {
		fmt.Println("No default theme found in Theme store, creating fallback theme...")
		return createFallbackTheme()
	}

	// Build the file URL
	buildFileURL := fmt.Sprintf("https://api.shipshipship.io/api/files/themes/%s/%s",
		defaultTheme.ID, defaultTheme.BuildFile)

	// Apply the default theme
	err = applyThemeInternal(defaultTheme.ID, defaultTheme.Version, buildFileURL)
	if err != nil {
		return fmt.Errorf("failed to apply default theme: %w", err)
	}

	fmt.Printf("Default theme '%s' (v%s) applied successfully\n",
		defaultTheme.DisplayName, defaultTheme.Version)
	return nil
}

// fetchDefaultThemeFromThemeStore fetches the default theme from Theme store
func fetchDefaultThemeFromThemeStore() (*ThemeStoreTheme, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch themes with name="shipshipship-template-default" and status="approved"
	url := "https://api.shipshipship.io/api/collections/themes/records?filter=(name='shipshipship-template-default'%26%26submission_status='approved')&sort=-created"

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Theme store: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Theme store returned status %d", resp.StatusCode)
	}

	var storeResponse ThemeStoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&storeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Theme store response: %w", err)
	}

	if len(storeResponse.Items) == 0 {
		return nil, nil // No default theme found
	}

	// Return the first (most recent) default theme
	theme := storeResponse.Items[0]
	return &theme, nil
}

// applyThemeInternal applies a theme without going through the HTTP handler
func applyThemeInternal(themeID, themeVersion, buildFileURL string) error {
	// Download the theme ZIP file
	tempFile, err := downloadThemeFile(buildFileURL)
	if err != nil {
		return fmt.Errorf("failed to download theme file: %w", err)
	}
	defer os.Remove(tempFile)

	// Create backup of current theme (if any)
	backupDir := "./data/themes/backup"
	if err := backupCurrentTheme(backupDir); err != nil {
		return fmt.Errorf("failed to backup current theme: %w", err)
	}

	// Extract the new theme
	themeDir := "./data/themes/current"
	if err := extractTheme(tempFile, themeDir); err != nil {
		// Restore backup on failure
		restoreThemeBackup(backupDir, themeDir)
		return fmt.Errorf("failed to extract theme: %w", err)
	}

	// Update settings to track current theme and version
	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	if err == nil {
		settings.CurrentThemeID = themeID
		settings.CurrentThemeVersion = themeVersion
		if err := db.Save(settings).Error; err != nil {
			fmt.Printf("Warning: Theme applied but couldn't save theme info: %v\n", err)
		}
	}

	// Clean up backup after successful application
	os.RemoveAll(backupDir)
	return nil
}

// ensureThemesDirectory creates the themes directory structure if it doesn't exist
func ensureThemesDirectory() error {
	themesDir := "./data/themes"
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}
	return nil
}

// cleanupAllThemes removes all theme-related directories for a clean slate
func cleanupAllThemes() error {
	themesDir := "./data/themes"
	if err := os.RemoveAll(themesDir); err != nil {
		return fmt.Errorf("failed to remove themes directory: %w", err)
	}
	return ensureThemesDirectory()
}

// getCurrentThemeSize returns the size of the current theme directory
func getCurrentThemeSize() (int64, error) {
	themeDir := "./data/themes/current"
	var size int64

	err := filepath.Walk(themeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return size, nil
}

// listInstalledThemes returns information about installed themes
func listInstalledThemes() map[string]interface{} {
	result := make(map[string]interface{})

	// Check current theme
	if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
		if size, err := getCurrentThemeSize(); err == nil {
			result["current"] = map[string]interface{}{
				"exists": true,
				"size":   size,
				"path":   "./data/themes/current",
			}
		}
	} else {
		result["current"] = map[string]interface{}{
			"exists": false,
		}
	}

	// Check backup
	if _, err := os.Stat("./data/themes/backup"); err == nil {
		result["backup"] = map[string]interface{}{
			"exists": true,
			"path":   "./data/themes/backup",
		}
	} else {
		result["backup"] = map[string]interface{}{
			"exists": false,
		}
	}

	return result
}

// createFallbackTheme creates a basic fallback theme when external API fails
func createFallbackTheme() error {
	db := database.GetDB()

	// Ensure themes directory exists
	if err := ensureThemesDirectory(); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	// Check if fallback theme already exists
	if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
		fmt.Println("Fallback theme files found")

		// Update database to mark theme as applied
		settings, err := models.GetOrCreateSettings(db)
		if err == nil && settings.CurrentThemeID == "" {
			settings.CurrentThemeID = "fallback"
			settings.CurrentThemeVersion = "1.0.0"
			if err := db.Save(settings).Error; err != nil {
				return fmt.Errorf("failed to save theme settings: %w", err)
			}
			fmt.Println("Database updated with fallback theme info")
		}
		return nil
	}

	fmt.Println("Fallback theme files not found, this should not happen in normal operation")

	// Update database settings
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	settings.CurrentThemeID = "fallback"
	settings.CurrentThemeVersion = "1.0.0"
	if err := db.Save(settings).Error; err != nil {
		return fmt.Errorf("failed to save theme settings: %w", err)
	}

	fmt.Println("Fallback theme created and applied successfully")
	return nil
}
