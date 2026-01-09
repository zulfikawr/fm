package config

import (
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
	tmpHome, err := os.MkdirTemp("", "fm-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)

	// Set HOME environment variable to our temp dir
	t.Setenv("HOME", tmpHome)

	cfg := DefaultConfig()
	cfg.ThemeIndex = 2
	cfg.ShowHidden = true

	// Test Saving
	err = cfg.Save()
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
	os.WriteFile(configPath, []byte("invalid json"), 0644)
	invalidCfg := Load()
	if invalidCfg.ThemeIndex != 0 {
		t.Errorf("Expected Default ThemeIndex 0 for invalid config, got %d", invalidCfg.ThemeIndex)
	}
}

func TestConfigSaveErrors(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "fm-config-err-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)
	t.Setenv("HOME", tmpHome)

	// Trigger MkdirAll error by creating a file where the directory should be
	configDir := filepath.Dir(GetConfigPath())
	err = os.MkdirAll(filepath.Dir(configDir), 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(configDir, []byte("i am a file, not a directory"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	err = cfg.Save()
	if err == nil {
		t.Error("Expected error when saving to a path where a file exists instead of a directory")
	}

	// Trigger WriteFile error by making the directory read-only
	// Reset HOME to a new clean dir
	tmpHome2, _ := os.MkdirTemp("", "fm-config-err2-test")
	defer os.RemoveAll(tmpHome2)
	t.Setenv("HOME", tmpHome2)

	configDir2 := filepath.Dir(GetConfigPath())
	os.MkdirAll(configDir2, 0555) // Read-only

	err = cfg.Save()
	if err == nil {
		t.Error("Expected error when writing config to a read-only directory")
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
	tmpHome, _ := os.MkdirTemp("", "fm-config-missing-test")
	defer os.RemoveAll(tmpHome)
	t.Setenv("HOME", tmpHome)

	cfg := Load()
	if cfg.ThemeIndex != 0 {
		t.Error("Expected default config when file is missing")
	}
}
