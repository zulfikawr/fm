package config

import (
	"os"
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
