package local

import (
	"context"
	"fm/internal/testutil"
	"testing"
)

func TestLocalFS(t *testing.T) {
	ctx := context.Background()
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()

	fs := NewLocalFS()

	t.Run("Stat and ReadDir", func(t *testing.T) {
		tmp.WriteFile("test.txt", "content")

		info, err := fs.Stat(ctx, tmp.Path)
		testutil.AssertNoError(t, err, "Stat tempDir should succeed")
		testutil.AssertEqual(t, true, info.IsDir(), "tempDir should be a directory")

		entries, err := fs.ReadDir(ctx, tmp.Path)
		testutil.AssertNoError(t, err, "ReadDir should succeed")
		testutil.AssertEqual(t, 1, len(entries), "Should find one file")
		testutil.AssertEqual(t, "test.txt", entries[0].Name(), "File name should match")
	})

	t.Run("Write and Remove", func(t *testing.T) {
		filePath := tmp.Join("write_test.txt")
		f, err := fs.Create(ctx, filePath)
		testutil.AssertNoError(t, err, "Create should succeed")
		f.Write([]byte("data"))
		f.Close()

		err = fs.RemoveAll(ctx, filePath)
		testutil.AssertNoError(t, err, "RemoveAll should succeed")
	})

	t.Run("MkdirAll and Rename", func(t *testing.T) {
		dirPath := tmp.Join("a/b/c")
		err := fs.MkdirAll(ctx, dirPath, 0755)
		testutil.AssertNoError(t, err, "MkdirAll should succeed")

		newPath := tmp.Join("new_dir")
		err = fs.Rename(ctx, tmp.Join("a"), newPath)
		testutil.AssertNoError(t, err, "Rename should succeed")
	})

	t.Run("Path helpers", func(t *testing.T) {
		testutil.AssertEqual(t, "file.txt", fs.Base("/a/b/file.txt"), "Base should work")
		testutil.AssertEqual(t, "/a/b", fs.Dir("/a/b/file.txt"), "Dir should work")
		testutil.AssertEqual(t, "/a/b/c", fs.Join("/a/b", "c"), "Join should work")
	})
}
