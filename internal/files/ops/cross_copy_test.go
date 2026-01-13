package ops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files/local"
	"fm/internal/testutil"
)

func TestCrossCopy(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	srcFS := &local.LocalFS{}
	dstFS := &local.LocalFS{}

	srcPath := filepath.Join(tmpDir, "cross_src.txt")
	dstPath := filepath.Join(tmpDir, "cross_dst.txt")
	content := "cross-filesystem content"
	testutil.CreateTestFile(t, tmpDir, "cross_src.txt", content)

	t.Run("Cross FS File", func(t *testing.T) {
		if err := CrossCopy(context.Background(), srcFS, dstFS, srcPath, dstPath, nil); err != nil {
			t.Fatalf("CrossCopy failed: %v", err)
		}

		got, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != content {
			t.Errorf("Expected content %s, got %s", content, string(got))
		}
	})

	t.Run("Cross FS Dir", func(t *testing.T) {
		srcDir := filepath.Join(tmpDir, "cross_dir_src")
		os.MkdirAll(filepath.Join(srcDir, "sub"), 0755)
		testutil.CreateTestFile(t, srcDir, "f1.txt", "f1")
		testutil.CreateTestFile(t, filepath.Join(srcDir, "sub"), "f2.txt", "f2")

		dstDir := filepath.Join(tmpDir, "cross_dir_dst")

		if err := CrossCopy(context.Background(), srcFS, dstFS, srcDir, dstDir, nil); err != nil {
			t.Fatalf("CrossCopy dir failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(dstDir, "f1.txt")); err != nil {
			t.Error("f1.txt not found in destination")
		}
		if _, err := os.Stat(filepath.Join(dstDir, "sub", "f2.txt")); err != nil {
			t.Error("sub/f2.txt not found in destination")
		}
	})
}
