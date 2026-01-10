package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-ops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "src.txt")
	content := []byte("hello world")
	os.WriteFile(srcFile, content, 0644)
	fs := &LocalFS{}

	t.Run("Copy File", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "dst.txt")
		if err := Copy(fs, srcFile, dstFile); err != nil {
			t.Fatalf("Copy failed: %v", err)
		}

		got, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(content) {
			t.Errorf("Expected content %s, got %s", string(content), string(got))
		}
	})

	t.Run("Rename File", func(t *testing.T) {
		oldPath := srcFile
		newPath := filepath.Join(tmpDir, "renamed.txt")
		if err := Rename(fs, oldPath, newPath); err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		if _, err := os.Stat(oldPath); err == nil {
			t.Error("Old file still exists after rename")
		}
		if _, err := os.Stat(newPath); err != nil {
			t.Error("New file does not exist after rename")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		path := filepath.Join(tmpDir, "todelete.txt")
		os.WriteFile(path, []byte("delete me"), 0644)
		if err := Delete(fs, path); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if _, err := os.Stat(path); err == nil {
			t.Error("File still exists after delete")
		}
	})

	t.Run("Copy Non-existent", func(t *testing.T) {
		err := Copy(fs, filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"))
		if err == nil {
			t.Error("Expected error when copying non-existent file")
		}
	})

	t.Run("Rename Non-existent", func(t *testing.T) {
		err := Rename(fs, filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"))
		if err == nil {
			t.Error("Expected error when renaming non-existent file")
		}
	})

	t.Run("Copy Dir Recursive", func(t *testing.T) {
		srcDir := filepath.Join(tmpDir, "recursive_src")
		subDir := filepath.Join(srcDir, "sub")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "inner.txt"), []byte("inner"), 0644)

		dstDir := filepath.Join(tmpDir, "recursive_dst")
		if err := Copy(fs, srcDir, dstDir); err != nil {
			t.Fatalf("Recursive copy failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(dstDir, "sub", "inner.txt")); err != nil {
			t.Error("Nested file not found in destination")
		}
	})
}
