package ops

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files/local"
	"fm/internal/testutil"
)

func TestCopy(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	srcFile := filepath.Join(tmpDir, "src.txt")
	content := "hello world"
	testutil.CreateTestFile(t, tmpDir, "src.txt", content)
	fs := &local.LocalFS{}

	t.Run("Copy File", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "dst.txt")
		if err := Copy(context.Background(), fs, srcFile, dstFile, nil); err != nil {
			t.Fatalf("Copy failed: %v", err)
		}

		got, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != content {
			t.Errorf("Expected content %s, got %s", content, string(got))
		}
	})

	t.Run("Copy Non-existent", func(t *testing.T) {
		err := Copy(context.Background(), fs, filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"), nil)
		if err == nil {
			t.Error("Expected error when copying non-existent file")
		}
	})

	t.Run("Copy Dir Recursive", func(t *testing.T) {
		srcDir := filepath.Join(tmpDir, "recursive_src")
		subDir := filepath.Join(srcDir, "sub")
		os.MkdirAll(subDir, 0755)
		testutil.CreateTestFile(t, subDir, "inner.txt", "inner")

		dstDir := filepath.Join(tmpDir, "recursive_dst")
		if err := Copy(context.Background(), fs, srcDir, dstDir, nil); err != nil {
			t.Fatalf("Recursive copy failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(dstDir, "sub", "inner.txt")); err != nil {
			t.Error("Nested file not found in destination")
		}
	})

	t.Run("Copy (Create Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_create_fail_src.txt")
		testutil.CreateTestFile(t, tmpDir, "copy_create_fail_src.txt", "content")
		dst := filepath.Join(tmpDir, "copy_create_fail_dst.txt")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return nil, os.ErrPermission
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy create fails")
		}
	})

	t.Run("Copy (Open Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_open_fail_src.txt")
		testutil.CreateTestFile(t, tmpDir, "copy_open_fail_src.txt", "content")
		dst := filepath.Join(tmpDir, "copy_open_fail_dst.txt")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return nil, os.ErrPermission
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy open fails")
		}
	})

	t.Run("Copy Dir (MkdirAll Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_dir_mkdir_src")
		os.MkdirAll(src, 0755)
		dst := filepath.Join(tmpDir, "copy_dir_mkdir_dst")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.MkdirAllFunc = func(ctx context.Context, path string, perm os.FileMode) error {
			return os.ErrPermission
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy dir MkdirAll fails")
		}
	})

	t.Run("Copy Dir (ReadDir Fails)", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_dir_read_src")
		os.MkdirAll(src, 0755)
		dst := filepath.Join(tmpDir, "copy_dir_read_dst")

		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
			return nil, os.ErrPermission
		}

		if err := Copy(context.Background(), mock, src, dst, nil); err == nil {
			t.Error("Expected error when Copy dir ReadDir fails")
		}
	})
}

func TestParallelCopyDir(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	srcDir := filepath.Join(tmpDir, "parallel_src")
	os.MkdirAll(srcDir, 0755)

	// Create many small files to test parallelism
	numFiles := 100
	for i := 0; i < numFiles; i++ {
		fname := fmt.Sprintf("file_%d.txt", i)
		testutil.CreateTestFile(t, srcDir, fname, fmt.Sprintf("content %d", i))
	}

	// Create a sub-directory with more files
	subDir := filepath.Join(srcDir, "sub")
	os.MkdirAll(subDir, 0755)
	for i := 0; i < 50; i++ {
		fname := fmt.Sprintf("sub_file_%d.txt", i)
		testutil.CreateTestFile(t, subDir, fname, fmt.Sprintf("sub content %d", i))
	}

	dstDir := filepath.Join(tmpDir, "parallel_dst")
	fs := &local.LocalFS{}

	if err := Copy(context.Background(), fs, srcDir, dstDir, nil); err != nil {
		t.Fatalf("Parallel copy failed: %v", err)
	}

	// Verify all files are present
	for i := 0; i < numFiles; i++ {
		fname := fmt.Sprintf("file_%d.txt", i)
		if _, err := os.Stat(filepath.Join(dstDir, fname)); err != nil {
			t.Errorf("File %s not found in destination", fname)
		}
	}

	for i := 0; i < 50; i++ {
		fname := fmt.Sprintf("sub_file_%d.txt", i)
		if _, err := os.Stat(filepath.Join(dstDir, "sub", fname)); err != nil {
			t.Errorf("Sub-file %s not found in destination", fname)
		}
	}
}

func TestParallelCopyDir_Error(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	srcDir := filepath.Join(tmpDir, "error_src")
	os.MkdirAll(srcDir, 0755)
	testutil.CreateTestFile(t, srcDir, "file1.txt", "content1")
	testutil.CreateTestFile(t, srcDir, "file2.txt", "content2")

	dstDir := filepath.Join(tmpDir, "error_dst")

	localFS := &local.LocalFS{}
	mock := testutil.NewMockFileSystem()
	mock.FileSystem = localFS

	// Fail when creating file2.txt in destination
	mock.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		if filepath.Base(path) == "file2.txt" {
			return nil, os.ErrPermission
		}
		return localFS.Create(ctx, path)
	}

	err := Copy(context.Background(), mock, srcDir, dstDir, nil)
	if err == nil {
		t.Error("Expected error from parallel copy, got nil")
	}
}
