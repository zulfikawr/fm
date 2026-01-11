package files

import (
	"context"
	"io"
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
		if err := Copy(context.Background(), fs, srcFile, dstFile, nil); err != nil {
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

	t.Run("Delete", func(t *testing.T) {
		path := filepath.Join(tmpDir, "todelete.txt")
		os.WriteFile(path, []byte("delete me"), 0644)
		if err := Delete(context.Background(), fs, path, nil); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if _, err := os.Stat(path); err == nil {
			t.Error("File still exists after delete")
		}
	})

	t.Run("Copy Non-existent", func(t *testing.T) {
		err := Copy(context.Background(), fs, filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"), nil)
		if err == nil {
			t.Error("Expected error when copying non-existent file")
		}
	})

	t.Run("Rename Non-existent", func(t *testing.T) {
		err := Rename(context.Background(), fs, filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"))
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
		if err := Copy(context.Background(), fs, srcDir, dstDir, nil); err != nil {
			t.Fatalf("Recursive copy failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(dstDir, "sub", "inner.txt")); err != nil {
			t.Error("Nested file not found in destination")
		}
	})

	t.Run("Move File (Success)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "move_src.txt")
		dst := filepath.Join(tmpDir, "move_dst.txt")
		os.WriteFile(src, []byte("move me"), 0644)

		if err := Move(context.Background(), fs, src, dst, nil); err != nil {
			t.Fatalf("Move failed: %v", err)
		}

		if _, err := os.Stat(src); err == nil {
			t.Error("Source file still exists after move")
		}
		got, _ := os.ReadFile(dst)
		if string(got) != "move me" {
			t.Errorf("Expected content 'move me', got %s", string(got))
		}
	})

	t.Run("Move Fallback (Cross-device)", func(t *testing.T) {
		// Mock FS that fails Rename but allows Copy/Delete (via LocalFS embedding)
		src := filepath.Join(tmpDir, "fallback_src.txt")
		dst := filepath.Join(tmpDir, "fallback_dst.txt")
		os.WriteFile(src, []byte("fallback"), 0644)

		mock := &MockFS{
			FileSystem: fs,
			RenameFunc: func(ctx context.Context, oldPath, newPath string) error {
				return os.ErrPermission // Simulate error to trigger fallback
			},
		}

		if err := Move(context.Background(), mock, src, dst, nil); err != nil {
			t.Fatalf("Move fallback failed: %v", err)
		}

		if _, err := os.Stat(src); err == nil {
			t.Error("Source file still exists after move fallback")
		}
		got, _ := os.ReadFile(dst)
		if string(got) != "fallback" {
			t.Errorf("Expected content 'fallback', got %s", string(got))
		}
	})

	t.Run("Move Fallback (Copy Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "move_copy_fail_src.txt")
		dst := filepath.Join(tmpDir, "move_copy_fail_dst.txt")
		os.WriteFile(src, []byte("fail copy"), 0644)

		mock := &MockFS{
			FileSystem: fs,
			RenameFunc: func(ctx context.Context, oldPath, newPath string) error {
				return os.ErrPermission
			},
			CreateFunc: func(ctx context.Context, path string) (io.WriteCloser, error) {
				return nil, os.ErrPermission
			},
		}

		if err := Move(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Move fallback copy fails")
		}
	})

	t.Run("Move Fallback (Delete Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "move_delete_fail_src.txt")
		dst := filepath.Join(tmpDir, "move_delete_fail_dst.txt")
		os.WriteFile(src, []byte("fail delete"), 0644)

		mock := &MockFS{
			FileSystem: fs,
			RenameFunc: func(ctx context.Context, oldPath, newPath string) error {
				return os.ErrPermission
			},
			RemoveAllFunc: func(ctx context.Context, path string) error {
				return os.ErrPermission
			},
		}

		if err := Move(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Move fallback delete fails")
		}
	})

	t.Run("Copy (Create Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_create_fail_src.txt")
		os.WriteFile(src, []byte("content"), 0644)
		dst := filepath.Join(tmpDir, "copy_create_fail_dst.txt")

		mock := &MockFS{
			FileSystem: fs,
			CreateFunc: func(ctx context.Context, path string) (io.WriteCloser, error) {
				return nil, os.ErrPermission
			},
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy create fails")
		}
	})

	t.Run("Copy (Open Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_open_fail_src.txt")
		os.WriteFile(src, []byte("content"), 0644)
		dst := filepath.Join(tmpDir, "copy_open_fail_dst.txt")

		mock := &MockFS{
			FileSystem: fs,
			OpenFunc: func(ctx context.Context, path string) (io.ReadCloser, error) {
				return nil, os.ErrPermission
			},
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy open fails")
		}
	})

	t.Run("Copy Dir (MkdirAll Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_dir_mkdir_src")
		os.MkdirAll(src, 0755)
		dst := filepath.Join(tmpDir, "copy_dir_mkdir_dst")

		mock := &MockFS{
			FileSystem: fs,
			MkdirAllFunc: func(ctx context.Context, path string, perm os.FileMode) error {
				return os.ErrPermission
			},
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy dir MkdirAll fails")
		}
	})

	t.Run("Copy Dir (ReadDir Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_dir_read_src")
		os.MkdirAll(src, 0755)
		dst := filepath.Join(tmpDir, "copy_dir_read_dst")

		mock := &MockFS{
			FileSystem: fs,
			ReadDirFunc: func(ctx context.Context, path string) ([]os.FileInfo, error) {
				return nil, os.ErrPermission
			},
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy dir ReadDir fails")
		}
	})
}
