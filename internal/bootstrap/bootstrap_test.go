package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"fm/internal/testutil"
)

func TestInitializeApp_Local(t *testing.T) {
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()

	t.Run("Default path (cwd)", func(t *testing.T) {
		app, err := InitializeApp("", []string{})
		if err != nil {
			t.Fatalf("InitializeApp failed: %v", err)
		}
		if app == nil {
			t.Fatal("App is nil")
		}

		cwd, _ := os.Getwd()
		if app.Model.Navigation.Path != cwd {
			t.Errorf("Expected start path %s, got %s", cwd, app.Model.Navigation.Path)
		}
	})

	t.Run("Specific directory", func(t *testing.T) {
		subDir := tmp.Mkdir("subdir")

		app, err := InitializeApp("", []string{subDir})
		if err != nil {
			t.Fatalf("InitializeApp failed: %v", err)
		}

		absSubDir, _ := filepath.Abs(subDir)
		if app.Model.Navigation.Path != absSubDir {
			t.Errorf("Expected start path %s, got %s", absSubDir, app.Model.Navigation.Path)
		}
	})

	t.Run("Non-existent directory", func(t *testing.T) {
		_, err := InitializeApp("", []string{tmp.Join("ghost")})
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("Path is a file", func(t *testing.T) {
		filePath := tmp.WriteFile("file.txt", "hello")

		_, err := InitializeApp("", []string{filePath})
		if err == nil {
			t.Error("Expected error for path being a file")
		}
	})
}
