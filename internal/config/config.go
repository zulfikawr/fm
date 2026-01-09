package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the user preferences.
type Config struct {
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

// DefaultConfig returns the initial configuration.
func DefaultConfig() Config {
	return Config{
		ThemeIndex:        0,
		ShowHidden:        false,
		CaseSensitive:     false,
		ConfirmOperations: true,
		WrapNavigation:    false,
		EnableGit:         true,
		ShowSize:          true,
		ShowDateModified:  true,
		ShowHeader:        true,
		DateFormatIndex:   0,
		SizeFormatIndex:   0,
		EditorIndex:       0,
		UseTrash:          true,
	}
}

// GetConfigPath returns the path to the config file in the user's home directory.
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fm", "config.json")
}

// Load reads the config from disk or returns default if not found.
func Load() Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		return DefaultConfig()
	}
	return cfg
}

var marshalIndent = json.MarshalIndent

// Save writes the current config to disk.
func (c Config) Save() error {
	path := GetConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := marshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
