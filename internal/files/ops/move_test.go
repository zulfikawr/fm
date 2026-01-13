package ops

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files/local"
	"fm/internal/testutil"
)

func TestMove(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	fs := &local.LocalFS{}

	t.Run("Move File (Success)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "move_src.txt")
		dst := filepath.Join(tmpDir, "move_dst.txt")
		testutil.CreateTestFile(t, tmpDir, "move_src.txt", "move me")

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
		src := filepath.Join(tmpDir, "fallback_src.txt")
		dst := filepath.Join(tmpDir, "fallback_dst.txt")
		testutil.CreateTestFile(t, tmpDir, "fallback_src.txt", "fallback")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.RenameFunc = func(ctx context.Context, oldPath, newPath string) error {
			return os.ErrPermission // Simulate error to trigger fallback
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
		testutil.CreateTestFile(t, tmpDir, "move_copy_fail_src.txt", "fail copy")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.RenameFunc = func(ctx context.Context, oldPath, newPath string) error {
			return os.ErrPermission
		}
		mock.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return nil, os.ErrPermission
		}

		if err := Move(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Move fallback copy fails")
		}
	})

	t.Run("Move Fallback (Delete Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "move_delete_fail_src.txt")
		dst := filepath.Join(tmpDir, "move_delete_fail_dst.txt")
		testutil.CreateTestFile(t, tmpDir, "move_delete_fail_src.txt", "fail delete")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.RenameFunc = func(ctx context.Context, oldPath, newPath string) error {
			return os.ErrPermission
		}
		mock.RemoveAllFunc = func(ctx context.Context, path string) error {
			return os.ErrPermission
		}

		if err := Move(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Move fallback delete fails")
		}
	})

	t.Run("Move Transactional (Verification Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "verify_fail_src.txt")
		dst := filepath.Join(tmpDir, "verify_fail_dst.txt")
		testutil.CreateTestFile(t, tmpDir, "verify_fail_src.txt", "verify me")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.RenameFunc = func(ctx context.Context, oldPath, newPath string) error {
			return os.ErrPermission // Force fallback to Copy
		}
		// Intercept Lstat for destination to simulate corruption (size mismatch)
		mock.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			info, err := fs.Lstat(ctx, path)
			if err == nil && path == dst {
				// Return a fake FileInfo with wrong size
				return &testutil.MockFileInfo{
					NameStr:   info.Name(),
					SizeInt:   info.Size() + 1, // Wrong size
					IsDirBool: info.IsDir(),
				}, nil
			}
			return info, err
		}

		err := Move(context.Background(), mock, src, dst, nil)
		if err == nil {
			t.Error("Expected error when Move verification fails")
		}

		// Transactional rollback check:
		// 1. Source should still exist
		if _, err := os.Stat(src); err != nil {
			t.Error("Source file should still exist after verification failure")
		}
		// 2. Destination should have been cleaned up
		if _, err := os.Stat(dst); err == nil {
			t.Error("Destination should have been cleaned up after verification failure")
		}
	})
}
