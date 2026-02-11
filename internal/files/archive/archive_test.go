package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func createTestZip(t *testing.T) string {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	files := []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"folder/file.txt", "File inside folder"},
	}

	var err error
	for _, file := range files {
		f, err := w.Create(file.Name)
		testutil.AssertNoError(t, err, "create zip entry")
		_, err = f.Write([]byte(file.Body))
		testutil.AssertNoError(t, err, "write zip entry body")
	}

	err = w.Close()
	testutil.AssertNoError(t, err, "close zip writer")

	tmpFile, err := os.CreateTemp("", "test*.zip")
	testutil.AssertNoError(t, err, "create temp zip file")
	_, err = tmpFile.Write(buf.Bytes())
	testutil.AssertNoError(t, err, "write temp zip file")
	err = tmpFile.Close()
	testutil.AssertNoError(t, err, "close temp zip file")

	return tmpFile.Name()
}

func TestArchiveFS(t *testing.T) {
	zipPath := createTestZip(t)
	defer func() {
		if err := os.Remove(zipPath); err != nil {
			t.Errorf("failed to remove test zip %s: %v", zipPath, err)
		}
	}()

	fs, err := NewArchiveFS(zipPath)
	testutil.AssertNoError(t, err, "NewArchiveFS")
	defer func() {
		if err := fs.Close(); err != nil {
			t.Errorf("failed to close archive fs: %v", err)
		}
	}()

	ctx := context.Background()

	t.Run("Zip ReadDir root", func(t *testing.T) {
		entries, err := fs.ReadDirEntries(ctx, "/")
		testutil.AssertNoError(t, err, "ReadDirEntries /")

		foundReadme := false
		foundFolder := false
		for _, e := range entries {
			if e.Name() == "readme.txt" {
				foundReadme = true
			}
			if e.Name() == "folder" && e.IsDir() {
				foundFolder = true
			}
		}

		if !foundReadme {
			t.Errorf("readme.txt not found in root")
		}
		if !foundFolder {
			t.Errorf("folder not found in root")
		}
	})

	t.Run("Zip Open and Read", func(t *testing.T) {
		r, err := fs.Open(ctx, "gopher.txt")
		testutil.AssertNoError(t, err, "Open gopher.txt")
		defer func() {
			if err := r.Close(); err != nil {
				t.Errorf("failed to close reader: %v", err)
			}
		}()

		data, err := io.ReadAll(r)
		testutil.AssertNoError(t, err, "ReadAll gopher.txt")

		if !bytes.Contains(data, []byte("George")) {
			t.Errorf("content mismatch")
		}
	})
}

func createTestTar(t *testing.T) string {
	buf := new(bytes.Buffer)
	w := tar.NewWriter(buf)

	files := []struct {
		Name, Body string
	}{
		{"readme.txt", "Tar archive readme."},
		{"folder/file.txt", "Tar file inside folder"},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}

		err := w.WriteHeader(hdr)
		testutil.AssertNoError(t, err, "write tar header")
		_, err = w.Write([]byte(file.Body))
		testutil.AssertNoError(t, err, "write tar body")
	}

	var err error
	err = w.Close()
	testutil.AssertNoError(t, err, "close tar writer")

	tmpFile, err := os.CreateTemp("", "test*.tar")
	testutil.AssertNoError(t, err, "create temp tar file")
	_, err = tmpFile.Write(buf.Bytes())
	testutil.AssertNoError(t, err, "write temp tar file")
	err = tmpFile.Close()
	testutil.AssertNoError(t, err, "close temp tar file")

	return tmpFile.Name()
}

func TestTarFS(t *testing.T) {
	tarPath := createTestTar(t)
	defer func() {
		if err := os.Remove(tarPath); err != nil {
			t.Errorf("failed to remove test tar %s: %v", tarPath, err)
		}
	}()

	fs, err := NewArchiveFS(tarPath)
	testutil.AssertNoError(t, err, "NewArchiveFS tar")
	defer func() {
		if err := fs.Close(); err != nil {
			t.Errorf("failed to close tar archive fs: %v", err)
		}
	}()

	ctx := context.Background()

	t.Run("Tar ReadDir root", func(t *testing.T) {
		entries, err := fs.ReadDirEntries(ctx, "/")
		testutil.AssertNoError(t, err, "ReadDirEntries /")

		foundReadme := false
		for _, e := range entries {
			if e.Name() == "readme.txt" {
				foundReadme = true
			}
		}

		if !foundReadme {
			t.Errorf("readme.txt not found in tar root")
		}
	})

	t.Run("Tar Open and Read", func(t *testing.T) {
		r, err := fs.Open(ctx, "readme.txt")
		testutil.AssertNoError(t, err, "Open readme.txt")
		defer func() {
			if err := r.Close(); err != nil {
				t.Errorf("failed to close tar reader: %v", err)
			}
		}()

		data, err := io.ReadAll(r)
		testutil.AssertNoError(t, err, "ReadAll")

		if !bytes.Contains(data, []byte("Tar archive readme.")) {
			t.Errorf("content mismatch")
		}
	})
}

func TestArchiveHelpers(t *testing.T) {
	fs := &ZipFS{baseArchiveFS: baseArchiveFS{archivePath: "test.zip"}}

	t.Run("Path Helpers", func(t *testing.T) {
		testutil.AssertEqual(t, "/", fs.Separator(), "Separator")
		testutil.AssertEqual(t, "a/b", fs.Join("a", "b"), "Join")
		abs, err := fs.Abs("rel")
		testutil.AssertNoError(t, err, "Abs")
		testutil.AssertEqual(t, "/rel", abs, "Abs path")
		testutil.AssertEqual(t, "file.txt", fs.Base("/a/file.txt"), "Base")
		testutil.AssertEqual(t, "/a", fs.Dir("/a/file.txt"), "Dir")
		testutil.AssertEqual(t, ".txt", fs.Ext("file.txt"), "Ext")
		testutil.AssertEqual(t, "/a/b", fs.Clean("//a/b/"), "Clean")
		home, err := fs.GetHomeDir()
		testutil.AssertNoError(t, err, "GetHomeDir")
		testutil.AssertEqual(t, "/", home, "HomeDir")
		testutil.AssertEqual(t, false, fs.IsLocal(), "IsLocal")
		testutil.AssertEqual(t, "test.zip", fs.Address(), "Address")
		testutil.AssertEqual(t, "", fs.User(), "User")
		ro, err := fs.IsReadOnly(context.Background(), "/")
		testutil.AssertNoError(t, err, "IsReadOnly")
		testutil.AssertEqual(t, true, ro, "IsReadOnly result")
	})

	t.Run("GetDefaultExtractionPath", func(t *testing.T) {
		testutil.AssertEqual(t, "/path/archive", GetDefaultExtractionPath(fs, "/path/archive.zip"), "zip")
		testutil.AssertEqual(t, "/path/archive", GetDefaultExtractionPath(fs, "/path/archive.tar.gz"), "tar.gz")
	})

	t.Run("Unsupported Operations", func(t *testing.T) {
		ctx := context.Background()
		if fs.MkdirAll(ctx, "/dir", 0755) == nil {
			t.Error("MkdirAll should fail")
		}
		if _, err := fs.Create(ctx, "/file"); err == nil {
			t.Error("Create should fail")
		}
		if fs.RemoveAll(ctx, "/dir") == nil {
			t.Error("RemoveAll should fail")
		}
		if fs.Rename(ctx, "/a", "/b") == nil {
			t.Error("Rename should fail")
		}
		if fs.Chmod(ctx, "/a", 0644) == nil {
			t.Error("Chmod should fail")
		}
		if fs.Preallocate(ctx, "/a", 100) == nil {
			t.Error("Preallocate should fail")
		}
	})
}

func TestArchiveFS_Extra(t *testing.T) {
	zipPath := createTestZip(t)
	defer func() {
		if err := os.Remove(zipPath); err != nil {
			t.Errorf("failed to remove test zip %s: %v", zipPath, err)
		}
	}()
	fs, err := NewArchiveFS(zipPath)
	testutil.AssertNoError(t, err, "NewArchiveFS")
	defer func() {
		if err := fs.Close(); err != nil {
			t.Errorf("failed to close archive fs: %v", err)
		}
	}()
	ctx := context.Background()

	t.Run("Stat", func(t *testing.T) {
		info, err := fs.Stat(ctx, "readme.txt")
		testutil.AssertNoError(t, err, "Stat readme.txt")
		testutil.AssertEqual(t, "readme.txt", info.Name(), "Name")
		testutil.AssertEqual(t, false, info.IsDir(), "IsDir")

		info, err = fs.Stat(ctx, "folder")
		testutil.AssertNoError(t, err, "Stat folder")
		testutil.AssertEqual(t, true, info.IsDir(), "IsDir folder")

		if _, err = fs.Stat(ctx, "nonexistent"); err == nil {
			t.Error("Stat nonexistent should fail")
		}
	})

	t.Run("Walk", func(t *testing.T) {
		count := 0
		err := fs.Walk(ctx, "/", func(path string, info os.FileInfo, err error) error {
			count++
			return nil
		})
		testutil.AssertNoError(t, err, "Walk")
		if count < 3 {
			t.Errorf("Expected at least 3 items in walk, got %d", count)
		}
	})

	t.Run("Unsupported Format", func(t *testing.T) {
		if _, err := NewArchiveFS("test.txt"); err == nil {
			t.Error("Unsupported format should fail")
		}
	})
}
