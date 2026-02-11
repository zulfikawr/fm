package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestInitializeApp_Local(t *testing.T) {
	tmpDir := testutil.TempDir(t)

	t.Run("Default path (cwd)", func(t *testing.T) {
		app, err := InitializeApp("", []string{})
		if err != nil {
			t.Fatalf("InitializeApp failed: %v", err)
		}
		if app == nil {
			t.Fatal("App is nil")
		}

		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if app.Model.Navigation.Path != cwd {
			t.Errorf("Expected start path %s, got %s", cwd, app.Model.Navigation.Path)
		}
	})

	t.Run("Specific directory", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		app, err := InitializeApp("", []string{subDir})
		if err != nil {
			t.Fatalf("InitializeApp failed: %v", err)
		}

		absSubDir, err := filepath.Abs(subDir)
		if err != nil {
			t.Fatal(err)
		}
		if app.Model.Navigation.Path != absSubDir {
			t.Errorf("Expected start path %s, got %s", absSubDir, app.Model.Navigation.Path)
		}
	})

	t.Run("Non-existent directory", func(t *testing.T) {
		_, err := InitializeApp("", []string{filepath.Join(tmpDir, "ghost")})
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("Path is a file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "file.txt")
		if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := InitializeApp("", []string{filePath})
		if err == nil {
			t.Error("Expected error for path being a file")
		}
	})
}
