package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/zulfikawr/fm/internal/logger"
)

// Config holds the user preferences.
type Config struct {
	ConfigVersion int            `json:"config_version"`
	UI            UIConfig       `json:"ui"`
	Ops           OpsConfig      `json:"ops"`
	External      ExternalConfig `json:"external"`
	Trash         TrashConfig    `json:"trash"`
	Keybindings   []Keybinding   `json:"keybindings"`
}

type UIConfig struct {
	ThemeIndex       int  `json:"theme_index"`
	ShowHidden       bool `json:"show_hidden"`
	ShowSize         bool `json:"show_size"`
	ShowDateModified bool `json:"show_date_modified"`
	ShowHeader       bool `json:"show_header"`
	ShowRAMUsage     bool `json:"show_ram_usage"`
	DateFormatIndex  int  `json:"date_format_index"`
	SizeFormatIndex  int  `json:"size_format_index"`
	EnableMouse      bool `json:"enable_mouse"`
	EnableIcons      bool `json:"enable_icons"`
}

type OpsConfig struct {
	ConfirmOperations bool `json:"confirm_operations"`
	WrapNavigation    bool `json:"wrap_navigation"`
	CaseSensitive     bool `json:"case_sensitive"`
	EnableRegexSearch bool `json:"enable_regex_search"`
}

type ExternalConfig struct {
	EnableGit   bool `json:"enable_git"`
	EditorIndex int  `json:"editor_index"`
}

type TrashConfig struct {
	UseTrash             bool `json:"use_trash"`
	TrashAutoCleanupDays int  `json:"trash_auto_cleanup_days"`
	TrashMaxSizeMB       int  `json:"trash_max_size_mb"`
}

const CurrentConfigVersion = 1

// DefaultConfig returns the initial configuration.
func DefaultConfig() Config {
	return Config{
		ConfigVersion: CurrentConfigVersion,
		UI: UIConfig{
			ThemeIndex:       0, // Gruvbox
			ShowHidden:       true,
			ShowSize:         true,
			ShowDateModified: true,
			ShowHeader:       false,
			DateFormatIndex:  0, // Default
			SizeFormatIndex:  0, // Full
			EnableMouse:      true,
			EnableIcons:      false,
			ShowRAMUsage:     false,
		},
		Ops: OpsConfig{
			ConfirmOperations: true,
			WrapNavigation:    false,
			CaseSensitive:     false,
			EnableRegexSearch: false,
		},
		External: ExternalConfig{
			EnableGit:   true,
			EditorIndex: 0, // Vim
		},
		Trash: TrashConfig{
			UseTrash:             false,
			TrashAutoCleanupDays: 30,
			TrashMaxSizeMB:       0, // Unlimited
		},
		Keybindings: DefaultKeybindings(),
	}
}

var customConfigPath string

func init() {
	// Isolate config for tests automatically
	if os.Getenv("GO_TEST") == "1" {
		tempDir, err := os.MkdirTemp("", "fm-test-config-*")
		if err == nil {
			customConfigPath = filepath.Join(tempDir, "config.json")
		}
	}
}

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
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", "config.json")
		}
		return filepath.Join(home, ".config", "fm", "config.json")
	}
	return filepath.Join(configDir, "fm", "config.json")
}

// GetCacheDir returns the path to the cache directory.
func GetCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", ".cache")
		}
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
		logger.Warnf("config parse failed: %s: %v, using defaults", path, err)
		return DefaultConfig()
	}

	// Validate and migrate if needed
	if err := cfg.Validate(); err != nil {
		logger.Warnf("config validation failed: %v, using defaults", err)
		return DefaultConfig()
	}

	// Auto-migrate old configs
	if cfg.ConfigVersion < CurrentConfigVersion {
		// Migration logic
		cfg.ConfigVersion = CurrentConfigVersion

		if err := cfg.Save(); err != nil {
			logger.Errorf("Failed to save migrated config: %v", err)
		}
	}

	return cfg
}

var marshalIndent = json.MarshalIndent

// Save writes the current config to disk.
func (cfg Config) Save() error {
	path := GetConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := marshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
