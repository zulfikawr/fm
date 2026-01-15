package tui

import (
	"context"
	"os"

	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"
)

// SetupTestApp creates a new App with a MockFileSystem and MockGitService for testing
func SetupTestApp(startPath string) (*App, *testutil.MockFileSystem) {
	fs := testutil.NewMockFileSystem()
	gs := testutil.NewMockGitService()

	// Set some defaults so the app doesn't crash on init
	fs.GetHomeDirFunc = func() (string, error) { return "/home/user", nil }
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) { return nil, nil }

	m := tuictx.NewModel(fs, startPath)
	m.GS = gs // Inject mock git service

	return NewApp(m), fs
}
