package ops

import (
	"context"
	"os"
	"testing"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestRename(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Successful Rename", func(t *testing.T) {
		called := false
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			called = true
			return nil
		}
		err := Rename(ctx, fs, "/old", "/new", conflict.Ask)
		testutil.AssertNoError(t, err, "Rename should succeed")
		if !called {
			t.Error("fs.Rename was not called")
		}
	})

	t.Run("Empty Path", func(t *testing.T) {
		err := Rename(ctx, fs, "", "/new", conflict.Ask)
		if err == nil {
			t.Error("Expected error for empty oldPath")
		}
	})

	t.Run("Invalid New Name", func(t *testing.T) {
		err := Rename(ctx, fs, "/old", "/path/invalid*name", conflict.Ask)
		if err == nil {
			t.Error("Expected error for invalid characters in new name")
		}
	})

	t.Run("Conflict Skip", func(t *testing.T) {
		// Mock destination exists
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/new" {
				return &testutil.MockFileInfo{FName: "new"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := Rename(ctx, fs, "/old", "/new", conflict.Skip)
		testutil.AssertNoError(t, err, "Should not error on skip")
	})
}
