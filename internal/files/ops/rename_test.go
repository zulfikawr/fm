package ops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files/local"
	"fm/internal/testutil"
)

func TestRename(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	srcFile := filepath.Join(tmpDir, "src.txt")
	testutil.CreateTestFile(t, tmpDir, "src.txt", "content")
	fs := &local.LocalFS{}

	t.Run("Rename File", func(t *testing.T) {
		oldPath := srcFile
		newPath := filepath.Join(tmpDir, "renamed.txt")
		if err := Rename(context.Background(), fs, oldPath, newPath); err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		if _, err := os.Stat(oldPath); err == nil {
			t.Error("Old file still exists after rename")
		}
		if _, err := os.Stat(newPath); err != nil {
			t.Error("New file does not exist after rename")
		}
	})

	t.Run("Rename Non-existent", func(t *testing.T) {
		err := Rename(context.Background(), fs, filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"))
		if err == nil {
			t.Error("Expected error when renaming non-existent file")
		}
	})
}
