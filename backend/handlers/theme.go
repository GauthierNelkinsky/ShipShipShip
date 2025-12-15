package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"shipshipship/constants"
	"shipshipship/database"
	"shipshipship/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ApplyThemeRequest struct {
	ThemeID       string              `json:"themeId" binding:"required"`
	ThemeVersion  string              `json:"themeVersion" binding:"required"`
	BuildFileURL  string              `json:"buildFileUrl" binding:"required"`
	Compatibility *ThemeCompatibility `json:"compatibility,omitempty"`
}

type ThemeCompatibility struct {
	MinVersion string `json:"minVersion,omitempty"`
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

	// Check compatibility if provided
	if req.Compatibility != nil && req.Compatibility.MinVersion != "" {
		// Get current app version from constants (should match admin/package.json)
		currentAppVersion := constants.AppVersion

		if !isVersionCompatible(currentAppVersion, req.Compatibility.MinVersion) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Theme requires version %s or higher. Current version is %s",
					req.Compatibility.MinVersion, currentAppVersion),
				"incompatible":    true,
				"requiredVersion": req.Compatibility.MinVersion,
				"currentVersion":  currentAppVersion,
			})
			return
		}
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

		// Load theme manifest and create default statuses/mappings
		manifest, err := models.LoadThemeManifest(themeDir)
		if err != nil {
			fmt.Printf("Warning: Theme applied but failed to load manifest: %v\n", err)
		} else {
			// Create default statuses from theme categories if none exist
			if err := models.CreateDefaultStatusesFromTheme(db, req.ThemeID, manifest); err != nil {
				fmt.Printf("Warning: Theme applied but failed to create default statuses: %v\n", err)
			} else {
				fmt.Printf("Successfully created default statuses from theme %s\n", req.ThemeID)
			}

			// Create default mappings for all statuses
			if err := models.CreateDefaultMappings(db, req.ThemeID, manifest); err != nil {
				fmt.Printf("Warning: Theme applied but failed to create default mappings: %v\n", err)
			} else {
				fmt.Printf("Successfully created default status mappings for theme %s\n", req.ThemeID)
			}
		}
	}

	// Clean up backup after successful application
	os.RemoveAll(backupDir)

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

	// Check if theme files actually exist
	themeFilesExist := false
	if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
		themeFilesExist = true
	}

	// If database says theme is installed but files are missing, clear the database
	if settings.CurrentThemeID != "" && !themeFilesExist {
		fmt.Println("Warning: Theme marked in database but files are missing. Clearing database entry.")
		settings.CurrentThemeID = ""
		settings.CurrentThemeVersion = ""
		if err := db.Save(settings).Error; err != nil {
			fmt.Printf("Error clearing theme info: %v\n", err)
		}
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

// RedownloadTheme redownloads the current theme by fetching it from the theme store
func RedownloadTheme(c *gin.Context) {
	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings", "details": err.Error()})
		return
	}

	if settings.CurrentThemeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No theme is currently installed"})
		return
	}

	// Fetch the theme from PocketBase API
	storeURL := fmt.Sprintf("https://api.shipshipship.io/api/collections/themes/records/%s", settings.CurrentThemeID)
	resp, err := http.Get(storeURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch theme from store", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Theme store returned status %d", resp.StatusCode)})
		return
	}

	var themeRecord struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Version     string `json:"version"`
		BuildFile   string `json:"build_file"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&themeRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse theme data", "details": err.Error()})
		return
	}

	if themeRecord.BuildFile == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Theme build file URL not found"})
		return
	}

	// Build full URL for the build file
	buildFileURL := fmt.Sprintf("https://api.shipshipship.io/api/files/themes/%s/%s", themeRecord.ID, themeRecord.BuildFile)

	// Use the current version or the latest version from store
	themeVersion := settings.CurrentThemeVersion
	if themeVersion == "" {
		themeVersion = themeRecord.Version
	}

	// Use applyThemeInternal to handle download, backup, extraction, and settings update
	if err := applyThemeInternal(themeRecord.ID, themeVersion, buildFileURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply theme", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Theme redownloaded successfully",
		"themeId":   themeRecord.ID,
		"themeName": themeRecord.DisplayName,
		"version":   themeVersion,
	})
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

	// Check if theme files already exist (most important check)
	if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
		fmt.Println("Theme files already exist")
		// Ensure database is in sync with reality
		if settings.CurrentThemeID == "" {
			settings.CurrentThemeID = "existing"
			settings.CurrentThemeVersion = "unknown"
			db.Save(settings)
			fmt.Println("Database updated to reflect existing theme files")
		}
		return nil
	}

	// Check if a theme is marked as applied in DB but files don't exist
	if settings.CurrentThemeID != "" {
		fmt.Printf("Warning: Theme '%s' is marked as applied in database but files are missing. Clearing database and re-initializing...\n", settings.CurrentThemeID)
		settings.CurrentThemeID = ""
		settings.CurrentThemeVersion = ""
		db.Save(settings)
	}

	fmt.Println("No theme applied, initializing default theme...")

	// Try to fetch and apply default theme with retries
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			fmt.Printf("Retry attempt %d of %d after %v...\n", attempt, maxRetries, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}

		// Fetch default theme from Theme store
		defaultTheme, err := fetchDefaultThemeFromThemeStore()
		if err != nil {
			fmt.Printf("Attempt %d: Failed to fetch default theme: %v\n", attempt, err)
			if attempt < maxRetries {
				continue
			}
			fmt.Println("All retry attempts failed. System will run without a theme.")
			return fmt.Errorf("failed to fetch default theme after %d attempts: %w", maxRetries, err)
		}

		if defaultTheme == nil {
			fmt.Println("No default theme found in Theme store")
			if attempt < maxRetries {
				continue
			}
			fmt.Println("No default theme available after all retry attempts.")
			return fmt.Errorf("no default theme found in theme store")
		}

		// Build the file URL
		buildFileURL := fmt.Sprintf("https://api.shipshipship.io/api/files/themes/%s/%s",
			defaultTheme.ID, defaultTheme.BuildFile)

		// Apply the default theme
		err = applyThemeInternal(defaultTheme.ID, defaultTheme.Version, buildFileURL)
		if err != nil {
			fmt.Printf("Attempt %d: Failed to apply default theme: %v\n", attempt, err)
			if attempt < maxRetries {
				continue
			}
			fmt.Println("All retry attempts failed. System will run without a theme.")
			return fmt.Errorf("failed to apply default theme after %d attempts: %w", maxRetries, err)
		}

		fmt.Printf("Default theme '%s' (v%s) applied successfully\n",
			defaultTheme.DisplayName, defaultTheme.Version)
		return nil
	}

	return fmt.Errorf("failed to initialize default theme after %d attempts", maxRetries)
}

// fetchDefaultThemeFromThemeStore fetches the default theme from Theme store
func fetchDefaultThemeFromThemeStore() (*ThemeStoreTheme, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Build filter based on environment
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "production"
	}

	var url string
	if environment == "development" {
		// In development, fetch approved OR staging themes
		url = "https://api.shipshipship.io/api/collections/themes/records?filter=(name='shipshipship-template-default'%26%26(submission_status='approved'||submission_status='staging'))&sort=-created"
	} else {
		// In production, only fetch approved themes
		url = "https://api.shipshipship.io/api/collections/themes/records?filter=(name='shipshipship-template-default'%26%26submission_status='approved')&sort=-created"
	}

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

		// Load theme manifest and create default statuses/mappings
		manifest, err := models.LoadThemeManifest(themeDir)
		if err != nil {
			fmt.Printf("Warning: Theme applied but failed to load manifest: %v\n", err)
		} else {
			// Create default statuses from theme categories if none exist
			if err := models.CreateDefaultStatusesFromTheme(db, themeID, manifest); err != nil {
				fmt.Printf("Warning: Theme applied but failed to create default statuses: %v\n", err)
			} else {
				fmt.Printf("Successfully created default statuses from theme %s\n", themeID)
			}

			// Create default mappings for all statuses
			if err := models.CreateDefaultMappings(db, themeID, manifest); err != nil {
				fmt.Printf("Warning: Theme applied but failed to create default mappings: %v\n", err)
			} else {
				fmt.Printf("Successfully created default status mappings for theme %s\n", themeID)
			}
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
		// Theme files exist
		themeData := map[string]interface{}{
			"exists": true,
			"path":   "./data/themes/current",
		}
		if size, err := getCurrentThemeSize(); err == nil {
			themeData["size"] = size
		}
		result["current"] = themeData
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

	fmt.Println("ERROR: Fallback theme files not found. Cannot mark theme as applied without actual files.")
	fmt.Println("The system will continue to run, but will serve the admin interface on the root path.")

	// DO NOT update database settings if there are no actual theme files
	// This prevents the frontend from thinking a theme is installed when it's not
	return fmt.Errorf("fallback theme files not found - system will run without a public theme")
}

// isVersionCompatible checks if the current version meets the minimum required version
func isVersionCompatible(currentVersion, minVersion string) bool {
	currentParts := parseVersion(currentVersion)
	minParts := parseVersion(minVersion)

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if i >= len(currentParts) || i >= len(minParts) {
			break
		}
		if currentParts[i] > minParts[i] {
			return true
		}
		if currentParts[i] < minParts[i] {
			return false
		}
	}

	return true // Equal versions are compatible
}

// parseVersion parses a semantic version string into numeric parts
func parseVersion(version string) []int {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	result := make([]int, 0, 3)

	for _, part := range parts {
		// Parse numeric part only (handles cases like "1.2.3-beta")
		numStr := ""
		for _, ch := range part {
			if ch >= '0' && ch <= '9' {
				numStr += string(ch)
			} else {
				break
			}
		}

		num := 0
		if numStr != "" {
			fmt.Sscanf(numStr, "%d", &num)
		}
		result = append(result, num)

		if len(result) >= 3 {
			break
		}
	}

	// Pad with zeros if needed
	for len(result) < 3 {
		result = append(result, 0)
	}

	return result
}
