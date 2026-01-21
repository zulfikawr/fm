package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestMain(m *testing.M) {
	// Isolate config for all tests in this package by default
	tempDir, err := os.MkdirTemp("", "fm-config-test-*")
	if err != nil {
		panic(err)
	}

	SetConfigPath(filepath.Join(tempDir, "config.json"))

	code := m.Run()

	_ = os.RemoveAll(tempDir)
	os.Exit(code)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ConfigVersion != CurrentConfigVersion {
		t.Errorf("Expected version %d, got %d", CurrentConfigVersion, cfg.ConfigVersion)
	}
}

func TestConfig_SaveLoad(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	SetConfigPath(configPath)
	defer SetConfigPath("")

	cfg := DefaultConfig()
	cfg.ThemeIndex = 5
	cfg.ShowHidden = true

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := Load()
	if loaded.ThemeIndex != 5 {
		t.Errorf("Expected ThemeIndex 5, got %d", loaded.ThemeIndex)
	}
	if !loaded.ShowHidden {
		t.Error("Expected ShowHidden true")
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("Valid default", func(t *testing.T) {
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Default config should be valid, got %v", err)
		}
	})

	t.Run("Invalid theme index", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.ThemeIndex = 999
		err := invalidCfg.Validate()
		if err == nil {
			t.Error("Expected error for invalid theme index")
		}
	})

	t.Run("Invalid date format", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.DateFormatIndex = 999
		err := invalidCfg.Validate()
		if err == nil {
			t.Error("Expected error for invalid date format")
		}
	})

	t.Run("Invalid size format", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.SizeFormatIndex = 999
		err := invalidCfg.Validate()
		if err == nil {
			t.Error("Expected error for invalid size format")
		}
	})

	t.Run("Invalid editor", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.EditorIndex = 999
		err := invalidCfg.Validate()
		if err == nil {
			t.Error("Expected error for invalid editor")
		}
	})
}

func TestConfig_LoadErrors(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	SetConfigPath(configPath)
	defer SetConfigPath("")

	t.Run("Non-existent file", func(t *testing.T) {
		cfg := Load()
		if cfg.ShowHidden != true { // Default is now true
			t.Errorf("Expected default ShowHidden true, got %v", cfg.ShowHidden)
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		if err := os.WriteFile(configPath, []byte("{ malformed }"), 0644); err != nil {
			t.Fatal(err)
		}
		cfg := Load()
		if cfg.ShowHidden != true {
			t.Error("Expected default config on malformed JSON")
		}
	})

	t.Run("Invalid values in file", func(t *testing.T) {
		if err := os.WriteFile(configPath, []byte(`{"theme_index": 999}`), 0644); err != nil {
			t.Fatal(err)
		}
		cfg := Load()
		if cfg.ThemeIndex != 0 {
			t.Errorf("Expected default ThemeIndex 0 on validation error, got %d", cfg.ThemeIndex)
		}
	})
}

func TestConfig_Migration(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	configPath := filepath.Join(tmpDir, "config.json")
	SetConfigPath(configPath)
	defer SetConfigPath("")

	t.Run("v0 migration", func(t *testing.T) {
		if err := os.WriteFile(configPath, []byte(`{"config_version": 0, "theme_index": 1}`), 0644); err != nil {
			t.Fatal(err)
		}
		cfg := Load()
		if cfg.ConfigVersion != CurrentConfigVersion {
			t.Errorf("Expected version %d, got %d", CurrentConfigVersion, cfg.ConfigVersion)
		}
		// v0 migration currently resets to default
		if cfg.ThemeIndex != 0 {
			t.Errorf("Expected ThemeIndex 0 after v0 reset, got %d", cfg.ThemeIndex)
		}
	})

	t.Run("Future version migration", func(t *testing.T) {
		// Just ensure it doesn't crash and keeps values if version is higher or same
		if err := os.WriteFile(configPath, []byte(`{"config_version": 1, "theme_index": 1}`), 0644); err != nil {
			t.Fatal(err)
		}
		cfg := Load()
		if cfg.ThemeIndex != 1 {
			t.Errorf("Expected ThemeIndex 1, got %d", cfg.ThemeIndex)
		}
	})
}

func TestGetCacheDir(t *testing.T) {
	dir := GetCacheDir()
	if dir == "" {
		t.Error("Cache dir should not be empty")
	}
}

func TestConfig_PathsNoCustom(t *testing.T) {
	oldPath := customConfigPath
	SetConfigPath("")
	defer SetConfigPath(oldPath)

	path := GetConfigPath()
	if path == "" {
		t.Error("Config path should not be empty")
	}

	cache := GetCacheDir()
	if cache == "" {
		t.Error("Cache dir should not be empty")
	}
}

func TestConfig_SaveError(t *testing.T) {
	// Try to save to a directory that cannot be created (or path that is a file)
	tempFile, err := os.CreateTemp("", "fm-config-save-err-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Setting config path to something under a file should make MkdirAll fail
	configPath := filepath.Join(tempFile.Name(), "config.json")
	SetConfigPath(configPath)
	defer SetConfigPath("")

	cfg := DefaultConfig()
	err = cfg.Save()
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestConfig_MarshalError(t *testing.T) {
	oldMarshal := marshalIndent
	defer func() { marshalIndent = oldMarshal }()

	// Mock marshal error
	marshalIndent = func(v any, prefix, indent string) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	}

	tmpDir := testutil.TempDir(t)
	SetConfigPath(filepath.Join(tmpDir, "config.json"))
	defer SetConfigPath("")

	cfg := DefaultConfig()
	err := cfg.Save()
	if err == nil || err.Error() != "marshal error" {
		t.Errorf("Expected marshal error, got %v", err)
	}
}

func TestConfig_PathsError(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	oldXDGCache := os.Getenv("XDG_CACHE_HOME")

	// Unset variables that os.UserConfigDir/os.UserCacheDir use
	os.Unsetenv("HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_CACHE_HOME")

	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)
		os.Setenv("XDG_CACHE_HOME", oldXDGCache)
	}()

	_ = GetConfigPath()
	_ = GetCacheDir()
}

func TestConfig_LoadValidVersion(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	configPath := filepath.Join(tmpDir, "config.json")
	SetConfigPath(configPath)
	defer SetConfigPath("")

	if err := os.WriteFile(configPath, []byte(`{"config_version": 1, "theme_index": 0}`), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := Load()
	if cfg.ConfigVersion != 1 {
		t.Errorf("Expected version 1, got %d", cfg.ConfigVersion)
	}
}
