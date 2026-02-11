package tui_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
)

func TestMain(m *testing.M) {
	// Isolate config for all tests in this package
	tempDir, err := os.MkdirTemp("", "fm-tui-test-*")
	if err != nil {
		panic(err)
	}

	config.SetConfigPath(filepath.Join(tempDir, "config.json"))

	code := m.Run()

	// Clean up
	if err := os.RemoveAll(tempDir); err != nil {
		panic(err)
	}
	os.Exit(code)
}

// SetupTestApp creates a new App with a MockFileSystem and MockGitService for testing
func SetupTestApp(startPath string) (*tui.App, *testutil.MockFileSystem) {
	fs := testutil.NewMockFileSystem()
	gs := testutil.NewMockGitService()

	// Set some defaults so the app doesn't crash on init
	fs.GetHomeDirFunc = func() (string, error) { return "/home/user", nil }
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) { return nil, nil }

	m := tuictx.NewModel(fs, startPath)
	m.GS = gs // Inject mock git service

	return tui.NewApp(m), fs
}
