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

	// Create temp directory for testing
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	extractDir := tmp.Join("extract")
	zipFile := tmp.Join("test.zip")

	// 1. Create some test files
	file1 := tmp.WriteFile("source/file1.txt", "hello world")
	file2 := tmp.WriteFile("source/file2.txt", "foo bar")
	subdir := tmp.Mkdir("source/subdir")
	tmp.WriteFile("source/subdir/file3.txt", "nested file")

	// 2. Test Zip
	t.Run("Zip", func(t *testing.T) {
		sources := []string{file1, file2, subdir}
		err := Zip(ctx, fs, sources, zipFile, nil, conflict.Ask)
		testutil.AssertNoError(t, err, "Zip operation")

		_, err = os.Stat(zipFile)
		testutil.AssertNoError(t, err, "Zip file should exist")

		// Verify zip content
		r, err := zip.OpenReader(zipFile)
		testutil.AssertNoError(t, err, "Open zip reader")
		defer r.Close()

		found := make(map[string]bool)
		for _, f := range r.File {
			found[f.Name] = true
		}

		testutil.AssertEqual(t, true, found["file1.txt"], "Should contain file1.txt")
		testutil.AssertEqual(t, true, found["file2.txt"], "Should contain file2.txt")
		testutil.AssertEqual(t, true, found["subdir/"], "Should contain subdir/")
		testutil.AssertEqual(t, true, found["subdir/file3.txt"], "Should contain subdir/file3.txt")
	})

	// 3. Test Unzip
	t.Run("Unzip", func(t *testing.T) {
		err := Unzip(ctx, fs, zipFile, extractDir, nil, conflict.Ask)
		testutil.AssertNoError(t, err, "Unzip operation")

		// Verify extracted content
		content1, err := os.ReadFile(filepath.Join(extractDir, "file1.txt"))
		testutil.AssertNoError(t, err, "Read extracted file1")
		testutil.AssertEqual(t, "hello world", string(content1), "Content file1")

		content3, err := os.ReadFile(filepath.Join(extractDir, "subdir", "file3.txt"))
		testutil.AssertNoError(t, err, "Read extracted file3")
		testutil.AssertEqual(t, "nested file", string(content3), "Content file3")

		// Verify subdirectory exists and has correct permissions (basic check)
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
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	zipFile := tmp.Join("malicious.zip")
	extractDir := tmp.Mkdir("extract")

	// Create a zip with a traversal path
	f, _ := os.Create(zipFile)
	zw := zip.NewWriter(f)

	// Malicious entry trying to escape extractDir
	// Note: on Windows filepath.Join cleans up ../ but we want to test the logic
	maliciousPath := "../../outside.txt"
	header := &zip.FileHeader{
		Name:   maliciousPath,
		Method: zip.Deflate,
	}
	writer, _ := zw.CreateHeader(header)
	_, _ = io.WriteString(writer, "evil")
	zw.Close()
	f.Close()

	// Try Unzip - use Overwrite so it doesn't fail on the existing extractDir
	err := Unzip(ctx, fs, zipFile, extractDir, nil, conflict.Overwrite)
	testutil.AssertNoError(t, err, "Unzip should not fail but skip malicious paths")

	// Verify file was NOT created outside
	outsidePath := tmp.Join("outside.txt")
	_, err2 := os.Stat(outsidePath)
	if !os.IsNotExist(err2) {
		t.Error("Vulnerability: File was created outside the extraction directory!")
	}
}

func TestZip_NoSources(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	err := Zip(ctx, fs, []string{}, "test.zip", nil, conflict.Ask)
	if err == nil {
		t.Error("Expected error for no sources")
	}
}

func TestUnzip_Remote(t *testing.T) {
	ctx := context.Background()

	// Mock FS to simulate remote (IsLocal returns false)
	fs := testutil.NewMockFileSystem()
	fs.IsLocalFunc = func() bool { return false }

	// Create a real zip on local disk to use as source
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	realZip := tmp.Join("source.zip")
	f, _ := os.Create(realZip)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("test.txt")
	_, _ = w.Write([]byte("content"))
	zw.Close()
	f.Close()

	// Mock FS methods needed for Unzip
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

	err := Unzip(ctx, fs, "remote.zip", tmp.Path, nil, conflict.Overwrite)
	testutil.AssertNoError(t, err, "Unzip remote")
}

func TestZip_Recursive(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	// Mock for recursive zipping
	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{NameStr: "src", IsDirBool: true}, nil
		}
		return &testutil.MockFileInfo{NameStr: "file.txt", IsDirBool: false}, nil
	}
	fs.StatFunc = fs.LstatFunc
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		if path == "/src" {
			return []os.FileInfo{&testutil.MockFileInfo{NameStr: "file.txt", IsDirBool: false}}, nil
		}
		return nil, nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("file.txt", []byte("content")), nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("out.zip", nil), nil
	}
	fs.DirFunc = filepath.Dir
	fs.RelFunc = filepath.Rel
	fs.BaseFunc = filepath.Base
	fs.JoinFunc = filepath.Join

	err := Zip(ctx, fs, []string{"/src"}, "/out.zip", nil, conflict.Overwrite)
	testutil.AssertNoError(t, err, "Recursive zip")
}

func TestUnzip_ConflictPolicies(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	zipFile := tmp.Join("test.zip")
	extractDir := tmp.Mkdir("extract")

	// Create a zip
	f, _ := os.Create(zipFile)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("file1.txt")
	_, _ = w.Write([]byte("new content"))
	zw.Close()
	f.Close()

	// Pre-create file1.txt
	tmp.WriteFile("extract/file1.txt", "old content")

	t.Run("Skip", func(t *testing.T) {
		err := Unzip(ctx, fs, zipFile, extractDir, nil, conflict.Skip)
		testutil.AssertNoError(t, err, "Unzip Skip")
		content, _ := os.ReadFile(tmp.Join("extract", "file1.txt"))
		testutil.AssertEqual(t, "old content", string(content), "Should not overwrite")
	})

	t.Run("Overwrite", func(t *testing.T) {
		err := Unzip(ctx, fs, zipFile, extractDir, nil, conflict.Overwrite)
		testutil.AssertNoError(t, err, "Unzip Overwrite")
		content, _ := os.ReadFile(tmp.Join("extract", "file1.txt"))
		testutil.AssertEqual(t, "new content", string(content), "Should overwrite")
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
	fs.DirFunc = filepath.Dir

	err := Zip(ctx, fs, []string{"/src"}, "/out.zip", nil, conflict.Overwrite)
	if err == nil {
		t.Error("Expected error from walkAndZip due to permission error")
	}
}

func TestZip_WalkError_Recursion(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{NameStr: "src", IsDirBool: true}, nil
		}
		if path == "/src/bad" {
			return nil, os.ErrPermission
		}
		return nil, os.ErrNotExist
	}
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		if path == "/src" {
			return []os.FileInfo{&testutil.MockFileInfo{NameStr: "bad"}}, nil
		}
		return nil, nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("out.zip", nil), nil
	}
	fs.DirFunc = filepath.Dir
	fs.JoinFunc = filepath.Join
	fs.RelFunc = filepath.Rel

	err := Zip(ctx, fs, []string{"/src"}, "/out.zip", nil, conflict.Overwrite)
	if err == nil {
		t.Error("Expected error from recursive walkAndZip")
	}
}

func TestZip_Conflict(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	srcFile := tmp.WriteFile("src.txt", "content")
	zipFile := tmp.WriteFile("existing.zip", "existing")

	err := Zip(ctx, fs, []string{srcFile}, zipFile, nil, conflict.Ask)
	if err == nil {
		t.Fatal("Expected conflict error")
	}
	if _, ok := err.(*conflict.ConflictError); !ok {
		t.Errorf("Expected ConflictError, got %T: %v", err, err)
	}
}

func TestUnzip_ConflictError(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	zipFile := tmp.Join("test.zip")
	extractDir := tmp.Mkdir("extract")

	// Create a zip
	f, _ := os.Create(zipFile)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("file1.txt")
	_, _ = w.Write([]byte("content"))
	zw.Close()
	f.Close()

	// Pre-create file1.txt
	tmp.WriteFile("extract/file1.txt", "existing")

	// Unzip with Ask policy should return ConflictError
	err := Unzip(ctx, fs, zipFile, extractDir, nil, conflict.Ask)
	if err == nil {
		t.Fatal("Expected conflict error")
	}
	if _, ok := err.(*conflict.ConflictError); !ok {
		t.Errorf("Expected ConflictError, got %T", err)
	}
}

func TestUnzip_MultipleFiles(t *testing.T) {
	ctx := context.Background()
	fs := local.NewLocalFS()
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	zipFile := tmp.Join("multi.zip")
	extractDir := tmp.Mkdir("extract")

	// Create a zip with 3 files
	f, _ := os.Create(zipFile)
	zw := zip.NewWriter(f)
	for i := 1; i <= 3; i++ {
		w, _ := zw.Create(fmt.Sprintf("file%d.txt", i))
		_, _ = w.Write([]byte(fmt.Sprintf("content%d", i)))
	}
	zw.Close()
	f.Close()

	progChan := make(chan core.Progress, 10)
	err := Unzip(ctx, fs, zipFile, extractDir, progChan, conflict.Overwrite)
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
