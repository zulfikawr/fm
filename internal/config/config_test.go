package config

import (
	"fm/internal/testutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ThemeIndex != 0 {
		t.Errorf("Expected theme index 0, got %d", cfg.ThemeIndex)
	}
	if !cfg.ConfirmOperations {
		t.Error("Expected ConfirmOperations to be true by default")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Mock home directory for testing
	tmpHome, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set HOME environment variable to our temp dir
	t.Setenv("HOME", tmpHome)

	cfg := DefaultConfig()
	cfg.ThemeIndex = 2
	cfg.ShowHidden = true

	// Test Saving
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists in the expected location
	configPath := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}

	// Test Loading
	loadedCfg := Load()
	if loadedCfg.ThemeIndex != 2 {
		t.Errorf("Expected ThemeIndex 2, got %d", loadedCfg.ThemeIndex)
	}
	if !loadedCfg.ShowHidden {
		t.Error("Expected ShowHidden to be true")
	}

	// Test Invalid Config
	testutil.CreateTestFile(t, filepath.Dir(configPath), filepath.Base(configPath), "invalid json")
	invalidCfg := Load()
	if invalidCfg.ThemeIndex != 0 {
		t.Errorf("Expected Default ThemeIndex 0 for invalid config, got %d", invalidCfg.ThemeIndex)
	}
}

func TestConfigSaveMarshalError(t *testing.T) {
	// Mock marshalIndent to return an error
	oldMarshal := marshalIndent
	marshalIndent = func(v any, prefix, indent string) ([]byte, error) {
		return nil, os.ErrPermission // Just some error
	}
	defer func() { marshalIndent = oldMarshal }()

	cfg := DefaultConfig()
	err := cfg.Save()
	if err == nil {
		t.Error("Expected error when marshalIndent fails")
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmpHome, cleanup := testutil.TempDir(t)
	defer cleanup()
	t.Setenv("HOME", tmpHome)

	cfg := Load()
	if cfg.ThemeIndex != 0 {
		t.Error("Expected default config when file is missing")
	}
}

func TestLoadPartialConfig(t *testing.T) {
	tmpHome, cleanup := testutil.TempDir(t)
	defer cleanup()
	t.Setenv("HOME", tmpHome)

	// Create a partial config file (missing show_date_modified)
	partialJSON := `{"config_version": 1, "show_hidden": true}`

	configPath := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	testutil.CreateTestFile(t, filepath.Dir(configPath), filepath.Base(configPath), partialJSON)

	cfg := Load()
	if !cfg.ShowDateModified {
		t.Error("Expected ShowDateModified to be true (default) when missing from config, got false")
	}
	if !cfg.ShowHidden {
		t.Error("Expected ShowHidden to be true (from file), got false")
	}
}

func TestGetCacheDir(t *testing.T) {
	tmpHome, cleanup := testutil.TempDir(t)
	defer cleanup()
	t.Setenv("HOME", tmpHome)

	cacheDir := GetCacheDir()
	expected := filepath.Join(tmpHome, ".cache", "fm")
	if cacheDir != expected {
		t.Errorf("Expected cache dir %s, got %s", expected, cacheDir)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"valid", DefaultConfig(), false},
		{"invalid theme low", func() Config { c := DefaultConfig(); c.ThemeIndex = -1; return c }(), true},
		{"invalid theme high", func() Config { c := DefaultConfig(); c.ThemeIndex = 99; return c }(), true},
		{"invalid date low", func() Config { c := DefaultConfig(); c.DateFormatIndex = -1; return c }(), true},
		{"invalid date high", func() Config { c := DefaultConfig(); c.DateFormatIndex = 99; return c }(), true},
		{"invalid size low", func() Config { c := DefaultConfig(); c.SizeFormatIndex = -1; return c }(), true},
		{"invalid size high", func() Config { c := DefaultConfig(); c.SizeFormatIndex = 99; return c }(), true},
		{"invalid editor low", func() Config { c := DefaultConfig(); c.EditorIndex = -1; return c }(), true},
		{"invalid editor high", func() Config { c := DefaultConfig(); c.EditorIndex = 99; return c }(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadMigration(t *testing.T) {
	tmpHome, cleanup := testutil.TempDir(t)
	defer cleanup()
	t.Setenv("HOME", tmpHome)

	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test v0 migration
	v0JSON := `{"config_version": 0, "theme_index": 5}`
	testutil.CreateTestFile(t, configDir, filepath.Base(configPath), v0JSON)
	cfg := Load()
	if cfg.ConfigVersion != CurrentConfigVersion {
		t.Errorf("Expected version %d, got %d", CurrentConfigVersion, cfg.ConfigVersion)
	}
	// v0 migration resets to default in this implementation
	if cfg.ThemeIndex != 0 {
		t.Errorf("Expected theme index 0 after v0 migration, got %d", cfg.ThemeIndex)
	}

	// Test intermediate version migration (if CurrentConfigVersion was > 1)
	vMinus1JSON := `{"config_version": -1, "theme_index": 5}`
	testutil.CreateTestFile(t, configDir, filepath.Base(configPath), vMinus1JSON)
	cfg = Load()
	if cfg.ConfigVersion != CurrentConfigVersion {
		t.Errorf("Expected version %d, got %d", CurrentConfigVersion, cfg.ConfigVersion)
	}
	// In else branch it doesn't reset to default, but we don't have many fields to check.
	// ThemeIndex 5 should be preserved if we didn't reset.
	if cfg.ThemeIndex != 5 {
		t.Errorf("Expected theme index 5 after v-1 migration, got %d", cfg.ThemeIndex)
	}
}

func TestGetConfigPathError(t *testing.T) {
	// Unset HOME and XDG_CONFIG_HOME to trigger error in os.UserConfigDir
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	path := GetConfigPath()
	// Should fallback to .config/fm/config.json relative to empty home
	expected := filepath.Join(".config", "fm", "config.json")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestGetCacheDirError(t *testing.T) {
	// Unset HOME and XDG_CACHE_HOME to trigger error in os.UserCacheDir
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	path := GetCacheDir()
	// Should fallback to .cache/fm relative to empty home
	expected := filepath.Join(".cache", "fm")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestLoadValidationFailure(t *testing.T) {
	tmpHome, cleanup := testutil.TempDir(t)
	defer cleanup()
	t.Setenv("HOME", tmpHome)

	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Invalid theme_index
	invalidJSON := `{"config_version": 1, "theme_index": 99}`
	testutil.CreateTestFile(t, configDir, filepath.Base(configPath), invalidJSON)

	cfg := Load()
	if cfg.ThemeIndex != 0 {
		t.Errorf("Expected default theme index 0 after validation failure, got %d", cfg.ThemeIndex)
	}
}
