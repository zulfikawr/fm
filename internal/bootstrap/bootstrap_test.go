package bootstrap

import (
	"os"
	"testing"

	"fm/internal/testutil"
)

func TestInitializeApp_Local(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	app, err := InitializeApp("", []string{tmpDir})
	if err != nil {
		t.Fatalf("InitializeApp failed: %v", err)
	}

	if app == nil {
		t.Fatal("Expected non-nil app")
	}

	if !app.Model.FS.IsLocal() {
		t.Error("Expected local filesystem")
	}

	if app.Model.Navigation.Path != tmpDir {
		t.Errorf("Expected path %s, got %s", tmpDir, app.Model.Navigation.Path)
	}
}

func TestInitializeApp_NoArgs(t *testing.T) {
	wd, _ := os.Getwd()
	app, err := InitializeApp("", []string{})
	if err != nil {
		t.Fatalf("InitializeApp failed: %v", err)
	}

	if app.Model.Navigation.Path != wd {
		t.Errorf("Expected path %s, got %s", wd, app.Model.Navigation.Path)
	}
}
