package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the user preferences.
type Config struct {
	ConfigVersion     int  `json:"config_version"`
	ThemeIndex        int  `json:"theme_index"`
	ShowHidden        bool `json:"show_hidden"`
	CaseSensitive     bool `json:"case_sensitive"`
	ConfirmOperations bool `json:"confirm_operations"`
	WrapNavigation    bool `json:"wrap_navigation"`
	EnableGit         bool `json:"enable_git"`
	ShowSize          bool `json:"show_size"`
	ShowDateModified  bool `json:"show_date_modified"`
	ShowHeader        bool `json:"show_header"`
	DateFormatIndex   int  `json:"date_format_index"`
	SizeFormatIndex   int  `json:"size_format_index"`
	EditorIndex       int  `json:"editor_index"`
	UseTrash          bool `json:"use_trash"`
}

const CurrentConfigVersion = 1

// DefaultConfig returns the initial configuration.
func DefaultConfig() Config {
	return Config{
		ConfigVersion:     CurrentConfigVersion,
		ThemeIndex:        0, // Gruvbox
		ShowHidden:        true,
		CaseSensitive:     false,
		ConfirmOperations: true,
		WrapNavigation:    false,
		EnableGit:         true,
		ShowSize:          true,
		ShowDateModified:  true,
		ShowHeader:        false,
		DateFormatIndex:   0, // Default
		SizeFormatIndex:   0, // Full
		EditorIndex:       0, // Vim
		UseTrash:          false,
	}
}

var customConfigPath string

// SetConfigPath overrides the default config path (useful for testing)
func SetConfigPath(path string) {
	customConfigPath = path
}

// GetConfigPath returns the path to the config file in the user's home directory.
func GetConfigPath() string {
	if customConfigPath != "" {
		return customConfigPath
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "fm", "config.json")
	}
	return filepath.Join(configDir, "fm", "config.json")
}

// GetCacheDir returns the path to the cache directory.
func GetCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".cache", "fm")
	}
	return filepath.Join(cacheDir, "fm")
}

// Load reads the config from disk or returns default if not found.
func Load() Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "config parse failed: %s: %v\n", path, err)
		return DefaultConfig()
	}

	// Validate and migrate if needed
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "config validation failed: %v, using defaults\n", err)
		return DefaultConfig()
	}

	// Auto-migrate old configs
	if cfg.ConfigVersion < CurrentConfigVersion {
		// For v0 -> v1 migration, reset to defaults since we're introducing versioning
		if cfg.ConfigVersion == 0 {
			cfg = DefaultConfig()
		} else {
			cfg.ConfigVersion = CurrentConfigVersion
		}
		_ = cfg.Save()
	}

	return cfg
}

var marshalIndent = json.MarshalIndent

// Save writes the current config to disk.
func (c Config) Save() error {
	path := GetConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := marshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
