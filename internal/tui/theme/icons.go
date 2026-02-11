package theme

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/logger"
)

// IconMapping holds the map of file/folder names and extensions to Nerd Font glyphs
type IconMapping struct {
	Files      map[string]string `json:"files"`
	Folders    map[string]string `json:"folders"`
	Extensions map[string]string `json:"extensions"`
}

var (
	mapping *IconMapping
	mu      sync.RWMutex
)

const (
	// Using a curated mapping for Nerd Fonts
	IconURL = "https://raw.githubusercontent.com/zulfikawr/fm/main/assets/icons.json"

	// Generic fallbacks
	GenericFileIcon    = "󰈔"
	GenericFolderIcon  = "󰉋"
	GenericSymlinkIcon = "󰁯"
)

// LoadIcons loads the icon mapping from disk or downloads it if missing
func LoadIcons() error {
	mu.Lock()
	defer mu.Unlock()

	cachePath := filepath.Join(config.GetCacheDir(), "icons.json")

	// For development: if assets/icons.json exists locally, always sync it to cache
	if _, err := os.Stat("assets/icons.json"); err == nil {
		if data, err := os.ReadFile("assets/icons.json"); err == nil {
			logger.LogIfError(os.MkdirAll(filepath.Dir(cachePath), 0o755), "icons: failed to create cache directory")
			logger.LogIfError(os.WriteFile(cachePath, data, 0o644), "icons: failed to write cache file")
		}
	}

	if mapping != nil {
		return nil
	}

	// Try loading from cache
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return fmt.Errorf("icons not downloaded")
	}

	if err := json.Unmarshal(data, &mapping); err != nil {
		return fmt.Errorf("failed to parse icons: %v", err)
	}

	return nil
}

// DownloadIcons downloads the icon mapping from the remote source
func DownloadIcons() error {
	// For local development/testing, check if assets/icons.json exists
	localPath := "assets/icons.json"
	if _, err := os.Stat(localPath); err == nil {
		data, err := os.ReadFile(localPath)
		if err == nil {
			cacheDir := config.GetCacheDir()
			if err := os.MkdirAll(cacheDir, 0o755); err == nil {
				cachePath := filepath.Join(cacheDir, "icons.json")
				if err := os.WriteFile(cachePath, data, 0o644); err == nil {
					mu.Lock()
					mapping = nil
					mu.Unlock()
					return LoadIcons()
				}
			}
		}
	}

	resp, err := http.Get(IconURL)
	if err != nil {
		return err
	}
	defer logger.CloseAndLog(resp.Body, "icon download response body")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download icons: status %d", resp.StatusCode)
	}

	cacheDir := config.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}

	cachePath := filepath.Join(cacheDir, "icons.json")
	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer logger.CloseAndLog(f, "icon cache file during download")

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	// Reset mapping to force reload
	mu.Lock()
	mapping = nil
	mu.Unlock()

	return LoadIcons()
}

// GetIcon returns the appropriate Nerd Font icon for the given item
func GetIcon(item core.Item) string {
	if item.State.IsUp {
		return ""
	}

	mu.RLock()
	defer mu.RUnlock()

	if mapping == nil {
		if item.IsDir {
			return GenericFolderIcon
		}
		return GenericFileIcon
	}

	name := strings.ToLower(item.Name)

	if item.IsDir {
		if icon, ok := mapping.Folders[name]; ok {
			return icon
		}
		return GenericFolderIcon
	}

	// Check full filename
	if icon, ok := mapping.Files[name]; ok {
		return icon
	}

	// Check extension
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	if icon, ok := mapping.Extensions[ext]; ok {
		return icon
	}

	return GenericFileIcon
}

// HasIconsDownloaded checks if the icons mapping file exists
func HasIconsDownloaded() bool {
	cachePath := filepath.Join(config.GetCacheDir(), "icons.json")
	_, err := os.Stat(cachePath)
	return err == nil
}
