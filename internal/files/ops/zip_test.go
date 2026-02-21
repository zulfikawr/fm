package ops

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/local"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestZipUnzip(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()

	tmpDir := testutil.TempDir(t)
	extractDir := filepath.Join(tmpDir, "extract")
	zipFile := filepath.Join(tmpDir, "test.zip")

	// 1. Create some test files
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(filepath.Join(sourceDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	file1 := filepath.Join(sourceDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}
	file2 := filepath.Join(sourceDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("foo bar"), 0644); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(sourceDir, "subdir")
	file3 := filepath.Join(subdir, "file3.txt")
	if err := os.WriteFile(file3, []byte("nested file"), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Test Zip
	t.Run("Zip", func(t *testing.T) {
		sources := []string{file1, file2, subdir}
		err := Zip(ZipOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Srcs:     sources,
			Dst:      zipFile,
			Conflict: ConflictOptions{Policy: conflict.Ask},
		})
		testutil.AssertNoError(t, err, "Zip operation")

		info, err := os.Stat(zipFile)
		testutil.AssertNoError(t, err, "Zip file should exist")
		if info == nil {
			t.Fatal("Expected non-nil FileInfo for zip file")
		}

		// Verify zip content
		r, err := zip.OpenReader(zipFile)
		testutil.AssertNoError(t, err, "Open zip reader")
		defer func() {
			if err := r.Close(); err != nil {
				t.Errorf("failed to close zip reader: %v", err)
			}
		}()

		found := make(map[string]bool)
		for i := range r.File {
			found[r.File[i].Name] = true
		}

		testutil.AssertEqual(t, true, found["file1.txt"], "Should contain file1.txt")
		testutil.AssertEqual(t, true, found["file2.txt"], "Should contain file2.txt")
		testutil.AssertEqual(t, true, found["subdir/"], "Should contain subdir/")
		testutil.AssertEqual(t, true, found["subdir/file3.txt"], "Should contain subdir/file3.txt")
	})

	// 3. Test Unzip
	t.Run("Unzip", func(t *testing.T) {
		err := Unzip(ZipOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Src:      zipFile,
			Dst:      extractDir,
			Conflict: ConflictOptions{Policy: conflict.Ask},
		})
		testutil.AssertNoError(t, err, "Unzip operation")

		// Verify extracted content
		content1, err := os.ReadFile(filepath.Join(extractDir, "file1.txt"))
		testutil.AssertNoError(t, err, "Read extracted file1")
		testutil.AssertEqual(t, "hello world", string(content1), "Content file1")

		content3, err := os.ReadFile(filepath.Join(extractDir, "subdir", "file3.txt"))
		testutil.AssertNoError(t, err, "Read extracted file3")
		testutil.AssertEqual(t, "nested file", string(content3), "Content file3")

		// Verify subdirectory exists
		info, err := os.Stat(filepath.Join(extractDir, "subdir"))
		testutil.AssertNoError(t, err, "Subdir should exist")
		if !info.IsDir() {
			t.Error("subdir should be a directory")
		}
	})
}

func TestUnzip_ZipSlip(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmpDir := testutil.TempDir(t)
	zipFile := filepath.Join(tmpDir, "malicious.zip")
	extractDir := filepath.Join(tmpDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a zip with a traversal path
	f, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(f)

	maliciousPath := "../../outside.txt"
	header := &zip.FileHeader{
		Name:   maliciousPath,
		Method: zip.Deflate,
	}
	writer, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}
	if n, err := io.WriteString(writer, "evil"); err != nil {
		t.Errorf("failed to write string (wrote %d bytes): %v", n, err)
	}
	if err := zw.Close(); err != nil {
		t.Errorf("failed to close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}

	err = Unzip(ZipOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		Src:      zipFile,
		Dst:      extractDir,
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	testutil.AssertNoError(t, err, "Unzip should not fail but skip malicious paths")

	outsidePath := filepath.Join(tmpDir, "outside.txt")
	info, err2 := os.Stat(outsidePath)
	if !os.IsNotExist(err2) {
		t.Errorf("Vulnerability: File was created outside the extraction directory! (info: %+v)", info)
	}
}

func TestZip_NoSources(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	err := Zip(ZipOptions{
		OpCtx: OpContext{Context: ctx, FS: fs},
		Srcs:  []string{},
		Dst:   "test.zip",
	})
	if err == nil {
		t.Error("Expected error for no sources")
	}
}

func TestUnzip_Remote(t *testing.T) {
	ctx := context.Background()

	fs := testutil.NewMockFileSystem()
	fs.IsLocalFunc = func() bool { return false }

	tmpDir := testutil.TempDir(t)
	realZip := filepath.Join(tmpDir, "source.zip")
	f, err := os.Create(realZip)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("test.txt")
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}
	if n, err := w.Write([]byte("content")); err != nil {
		t.Errorf("failed to write content (wrote %d bytes): %v", n, err)
	}
	if err := zw.Close(); err != nil {
		t.Errorf("failed to close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return os.Stat(realZip)
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return os.Open(realZip)
	}
	fs.AbsFunc = func(path string) (string, error) { return path, nil }
	fs.RelFunc = func(base, target string) (string, error) {
		return filepath.Rel(base, target)
	}
	fs.MkdirAllFunc = func(ctx context.Context, path string, perm os.FileMode) error { return nil }
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile(path, nil), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err = Unzip(ZipOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		Src:      "remote.zip",
		Dst:      tmpDir,
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	testutil.AssertNoError(t, err, "Unzip remote")
}

func TestZip_Recursive(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{FName: "src", FIsDir: true}, nil
		}
		return &testutil.MockFileInfo{FName: "file.txt", FIsDir: false}, nil
	}
	fs.StatFunc = fs.LstatFunc
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		if path == "/src" {
			return []os.FileInfo{&testutil.MockFileInfo{FName: "file.txt", FIsDir: false}}, nil
		}
		return nil, nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("file.txt", []byte("content")), nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("out.zip", nil), nil
	}
	fs.DirFunc = func(path string) string { return filepath.Dir(path) }
	fs.RelFunc = func(base, target string) (string, error) { return filepath.Rel(base, target) }
	fs.JoinFunc = func(elem ...string) string { return filepath.Join(elem...) }

	err := Zip(ZipOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		Srcs:     []string{"/src"},
		Dst:      "/out.zip",
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	testutil.AssertNoError(t, err, "Recursive zip")
}

func TestUnzip_ConflictPolicies(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmpDir := testutil.TempDir(t)
	zipFile := filepath.Join(tmpDir, "test.zip")
	extractDir := filepath.Join(tmpDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a zip
	f, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("file1.txt")
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}
	if n, err := w.Write([]byte("new content")); err != nil {
		t.Errorf("failed to write content (wrote %d bytes): %v", n, err)
	}
	if err := zw.Close(); err != nil {
		t.Errorf("failed to close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}

	// Pre-create file1.txt
	if err := os.WriteFile(filepath.Join(extractDir, "file1.txt"), []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("Skip", func(t *testing.T) {
		err := Unzip(ZipOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Src:      zipFile,
			Dst:      extractDir,
			Conflict: ConflictOptions{Policy: conflict.Skip},
		})
		testutil.AssertNoError(t, err, "Unzip Skip")
		content, err := os.ReadFile(filepath.Join(extractDir, "file1.txt"))
		testutil.AssertNoError(t, err, "ReadFile")
		testutil.AssertEqual(t, "old content", string(content), "Should not overwrite")
	})

	t.Run("Overwrite", func(t *testing.T) {
		err := Unzip(ZipOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Src:      zipFile,
			Dst:      extractDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite},
		})
		testutil.AssertNoError(t, err, "Unzip Overwrite")
		content, err := os.ReadFile(filepath.Join(extractDir, "file1.txt"))
		testutil.AssertNoError(t, err, "ReadFile")
		testutil.AssertEqual(t, "new content", string(content), "Should overwrite")
	})

	t.Run("OverwriteDirectoryWithSameName", func(t *testing.T) {
		// Create a directory with the same name as the destination
		destDir := filepath.Join(tmpDir, "myarchive")
		if err := os.MkdirAll(filepath.Join(destDir, "subdir"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(destDir, "existing.txt"), []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create a zip with different content
		zipFile2 := filepath.Join(tmpDir, "myarchive.zip")
		f2, err := os.Create(zipFile2)
		if err != nil {
			t.Fatalf("failed to create zip: %v", err)
		}
		zw2 := zip.NewWriter(f2)
		w2, err := zw2.Create("newfile.txt")
		if err != nil {
			t.Fatalf("failed to create entry: %v", err)
		}
		if _, err := w2.Write([]byte("new file content")); err != nil {
			t.Fatalf("failed to write content: %v", err)
		}
		if err := zw2.Close(); err != nil {
			t.Fatalf("failed to close zip writer: %v", err)
		}
		if err := f2.Close(); err != nil {
			t.Fatalf("failed to close file: %v", err)
		}

		// Unzip with overwrite - should remove the existing directory first
		err = Unzip(ZipOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Src:      zipFile2,
			Dst:      destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite},
		})
		testutil.AssertNoError(t, err, "Unzip should overwrite existing directory")

		// Verify old files are gone
		if _, err := os.Stat(filepath.Join(destDir, "existing.txt")); !os.IsNotExist(err) {
			t.Error("Old file should be removed")
		}
		if _, err := os.Stat(filepath.Join(destDir, "subdir")); !os.IsNotExist(err) {
			t.Error("Old subdirectory should be removed")
		}

		// Verify new file exists
		content, err := os.ReadFile(filepath.Join(destDir, "newfile.txt"))
		testutil.AssertNoError(t, err, "ReadFile")
		testutil.AssertEqual(t, "new file content", string(content), "Should have new content")
	})
}

func TestZip_WalkError(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return nil, os.ErrPermission
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("out.zip", nil), nil
	}
	fs.DirFunc = func(path string) string { return filepath.Dir(path) }

	err := Zip(ZipOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		Srcs:     []string{"/src"},
		Dst:      "/out.zip",
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	if err == nil {
		t.Error("Expected error from walkAndZip due to permission error")
	}
}

func TestZip_Conflict(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmpDir := testutil.TempDir(t)
	srcFile := filepath.Join(tmpDir, "src.txt")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	zipFile := filepath.Join(tmpDir, "existing.zip")
	if err := os.WriteFile(zipFile, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Zip(ZipOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		Srcs:     []string{srcFile},
		Dst:      zipFile,
		Conflict: ConflictOptions{Policy: conflict.Ask},
	})
	if err == nil {
		t.Fatal("Expected conflict error")
	}
	conflictErr, ok := err.(*conflict.ConflictError)
	if !ok {
		t.Errorf("Expected ConflictError, got %T: %v (conflictErr: %+v)", err, err, conflictErr)
	}
}

func TestUnzip_ConflictError(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmpDir := testutil.TempDir(t)
	zipFile := filepath.Join(tmpDir, "test.zip")
	extractDir := filepath.Join(tmpDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a zip
	f, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("file1.txt")
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}
	if n, err := w.Write([]byte("content")); err != nil {
		t.Errorf("failed to write content (wrote %d bytes): %v", n, err)
	}
	if err := zw.Close(); err != nil {
		t.Errorf("failed to close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}

	// Pre-create file1.txt
	if err := os.WriteFile(filepath.Join(extractDir, "file1.txt"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	// Unzip with Ask policy should return ConflictError
	err = Unzip(ZipOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		Src:      zipFile,
		Dst:      extractDir,
		Conflict: ConflictOptions{Policy: conflict.Ask},
	})
	if err == nil {
		t.Fatal("Expected conflict error")
	}
	conflictErr, ok := err.(*conflict.ConflictError)
	if !ok {
		t.Errorf("Expected ConflictError, got %T (conflictErr: %+v)", err, conflictErr)
	}
}

func TestUnzip_MultipleFiles(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmpDir := testutil.TempDir(t)
	zipFile := filepath.Join(tmpDir, "multi.zip")
	extractDir := filepath.Join(tmpDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a zip with 3 files
	f, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	for i := 1; i <= 3; i++ {
		w, err := zw.Create(fmt.Sprintf("file%d.txt", i))
		if err != nil {
			t.Fatalf("failed to create entry %d: %v", i, err)
		}
		if n, err := fmt.Fprintf(w, "content%d", i); err != nil {
			t.Errorf("failed to write content %d (wrote %d bytes): %v", i, n, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Errorf("failed to close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("failed to close file: %v", err)
	}

	progChan := make(chan core.Progress, 10)
	err = Unzip(ZipOptions{
		OpCtx: OpContext{
			Context:  ctx,
			FS:       fs,
			Progress: progChan,
		},
		Src:      zipFile,
		Dst:      extractDir,
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	testutil.AssertNoError(t, err, "Unzip multi")

	// Verify progress was sent
	count := 0
loop:
	for {
		select {
		case <-progChan:
			count++
		default:
			break loop
		}
	}
	if count == 0 {
		t.Error("Expected progress updates for multi-file unzip")
	}
}
