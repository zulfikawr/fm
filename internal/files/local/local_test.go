package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestLocalFS(t *testing.T) {
	ctx := context.Background()
	tmpDir := testutil.TempDir(t)

	var fs core.FileSystem = NewLocalFS()
	fs = core.NewErrorWrappedFS(fs)
	fs = core.NewCachedFS(fs, 100, 0)
	fs = core.NewContextFS(fs)

	t.Run("Stat and ReadDir", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		info, err := fs.Stat(ctx, tmpDir)
		testutil.AssertNoError(t, err, "Stat tempDir should succeed")
		testutil.AssertEqual(t, true, info.IsDir(), "tempDir should be a directory")

		entries, err := fs.ReadDir(ctx, tmpDir)
		testutil.AssertNoError(t, err, "ReadDir should succeed")
		testutil.AssertEqual(t, 1, len(entries), "Should find one file")
		testutil.AssertEqual(t, "test.txt", entries[0].Name(), "File name should match")
	})

	t.Run("Write and Remove", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "write_test.txt")
		f, err := fs.Create(ctx, filePath)
		testutil.AssertNoError(t, err, "Create should succeed")
		if n, err := f.Write([]byte("data")); err != nil {
			t.Errorf("failed to write (wrote %d bytes): %v", n, err)
		}
		if err := f.Close(); err != nil {
			t.Errorf("failed to close: %v", err)
		}

		err = fs.RemoveAll(ctx, filePath)
		testutil.AssertNoError(t, err, "RemoveAll should succeed")
	})

	t.Run("MkdirAll and Rename", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "a/b/c")
		err := fs.MkdirAll(ctx, dirPath, 0755)
		testutil.AssertNoError(t, err, "MkdirAll should succeed")

		newPath := filepath.Join(tmpDir, "new_dir")
		err = fs.Rename(ctx, filepath.Join(tmpDir, "a"), newPath)
		testutil.AssertNoError(t, err, "Rename should succeed")
	})

	t.Run("Path helpers", func(t *testing.T) {
		testutil.AssertEqual(t, "file.txt", fs.Base("/a/b/file.txt"), "Base should work")
		testutil.AssertEqual(t, "/a/b", fs.Dir("/a/b/file.txt"), "Dir should work")
		testutil.AssertEqual(t, "/a/b/c", fs.Join("/a/b", "c"), "Join should work")
		testutil.AssertEqual(t, ".txt", fs.Ext("file.txt"), "Ext should work")
		testutil.AssertEqual(t, "/", fs.Clean("//"), "Clean should work")
	})

	t.Run("Open and Read", func(t *testing.T) {
		path := filepath.Join(tmpDir, "read_test.txt")
		if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}

		f, err := fs.Open(ctx, path)
		testutil.AssertNoError(t, err, "Open should succeed")
		defer func() {
			if err := f.Close(); err != nil {
				t.Errorf("failed to close reader: %v", err)
			}
		}()

		data, err := io.ReadAll(f)
		testutil.AssertNoError(t, err, "ReadAll should succeed")
		testutil.AssertEqual(t, "hello", string(data), "Content should match")
	})

	t.Run("Chmod", func(t *testing.T) {
		path := filepath.Join(tmpDir, "chmod_test.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		err := fs.Chmod(ctx, path, 0600)
		testutil.AssertNoError(t, err, "Chmod should succeed")

		info, err := fs.Stat(ctx, path)
		testutil.AssertNoError(t, err, "Stat should succeed")
		if info.Mode().Perm() != 0600 && runtime.GOOS != "windows" {
			t.Errorf("expected 0600, got %o", info.Mode().Perm())
		}
	})

	t.Run("ReadDirEntries", func(t *testing.T) {
		path := filepath.Join(tmpDir, "entry.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		entries, err := fs.ReadDirEntries(ctx, tmpDir)
		testutil.AssertNoError(t, err, "ReadDirEntries should succeed")
		found := false
		for i := range entries {
			if entries[i].Name() == "entry.txt" {
				found = true
				break
			}
		}
		testutil.AssertEqual(t, true, found, "Should find entry.txt")
	})

	t.Run("Lstat and Abs", func(t *testing.T) {
		path := filepath.Join(tmpDir, "lstat_test.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		info, err := fs.Lstat(ctx, path)
		testutil.AssertNoError(t, err, "Lstat should succeed")
		if info == nil {
			t.Fatal("Expected non-nil FileInfo from Lstat")
		}

		abs, err := fs.Abs("some_file")
		testutil.AssertNoError(t, err, "Abs should succeed")
		if !filepath.IsAbs(abs) {
			t.Errorf("expected absolute path, got %s", abs)
		}
	})

	t.Run("Rel", func(t *testing.T) {
		rel, err := fs.Rel("/a/b", "/a/b/c/d")
		testutil.AssertNoError(t, err, "Rel should succeed")
		testutil.AssertEqual(t, filepath.Join("c", "d"), rel, "Rel should match")
	})

	t.Run("System Info", func(t *testing.T) {
		testutil.AssertEqual(t, true, fs.IsLocal(), "IsLocal should be true")
		testutil.AssertEqual(t, "", fs.Address(), "Address should be empty")
		testutil.AssertEqual(t, "", fs.User(), "User should be empty")
		testutil.AssertEqual(t, string(os.PathSeparator), fs.Separator(), "Separator should match")

		home, err := fs.GetHomeDir()
		testutil.AssertNoError(t, err, "GetHomeDir should succeed")
		if home == "" {
			t.Error("Expected non-empty home directory")
		}
	})

	t.Run("Walk", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(tmpDir, "walk_a/b"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "walk_a/1.txt"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "walk_a/b/2.txt"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		count := 0
		err := fs.Walk(ctx, filepath.Join(tmpDir, "walk_a"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				count++
			}
			return nil
		})
		testutil.AssertNoError(t, err, "Walk should succeed")
		if count < 2 {
			t.Errorf("expected at least 2 files, got %d", count)
		}
	})

	t.Run("IsReadOnly", func(t *testing.T) {
		tmpDir := testutil.TempDir(t)
		ro, err := fs.IsReadOnly(ctx, tmpDir)
		testutil.AssertNoError(t, err, "IsReadOnly should succeed on tempDir")
		testutil.AssertEqual(t, false, ro, "tempDir should not be read-only")

		if runtime.GOOS != "windows" && os.Getuid() != 0 {
			path := filepath.Join(tmpDir, "ro_file.txt")
			if err := os.WriteFile(path, []byte(""), 0400); err != nil {
				t.Fatal(err)
			}
			ro, err = fs.IsReadOnly(ctx, path)
			testutil.AssertNoError(t, err, "IsReadOnly should succeed on RO file")
			testutil.AssertEqual(t, true, ro, "file should be read-only")
		}
	})

	t.Run("Preallocate", func(t *testing.T) {
		path := filepath.Join(tmpDir, "prealloc.txt")
		// Just ensure it doesn't crash; error is okay if not supported
		err := fs.Preallocate(ctx, path, 1024)
		if err != nil {
			t.Logf("Preallocate (expectedly) failed: %v", err)
		}
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		ictx, cancel := context.WithCancel(context.Background())
		cancel()
		info, err := fs.Stat(ictx, tmpDir)
		if err == nil {
			t.Errorf("expected error on cancelled context (got info: %+v)", info)
		}
		entries, err := fs.ReadDir(ictx, tmpDir)
		if err == nil {
			t.Errorf("expected error on cancelled context (got entries: %v)", entries)
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := fs.Close()
		testutil.AssertNoError(t, err, "Close should succeed")
	})

	t.Run("getStatInfo", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			res := getStatInfo(nil)
			testutil.AssertEqual(t, (*statInfo)(nil), res, "should be nil on windows")
		} else {
			info, err := os.Stat(tmpDir)
			if err != nil {
				t.Fatalf("failed to stat temp dir: %v", err)
			}
			res := getStatInfo(info.Sys())
			if res == nil {
				t.Errorf("expected non-nil statInfo on unix")
			}
			if nilStat := getStatInfo(nil); nilStat != nil {
				t.Errorf("expected nil result for nil input, got %+v", nilStat)
			}
		}
	})
}
