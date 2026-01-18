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
	tmpFile.Close()

	return tmpFile.Name()
}

func TestArchiveFS(t *testing.T) {
	zipPath := createTestZip(t)
	defer os.Remove(zipPath)

	fs, err := NewArchiveFS(zipPath)
	testutil.AssertNoError(t, err, "NewArchiveFS")
	defer fs.Close()

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
		defer r.Close()

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

	err := w.Close()
	testutil.AssertNoError(t, err, "close tar writer")

	tmpFile, err := os.CreateTemp("", "test*.tar")
	testutil.AssertNoError(t, err, "create temp tar file")
	_, err = tmpFile.Write(buf.Bytes())
	testutil.AssertNoError(t, err, "write temp tar file")
	tmpFile.Close()

	return tmpFile.Name()
}

func TestTarFS(t *testing.T) {
	tarPath := createTestTar(t)
	defer os.Remove(tarPath)

	fs, err := NewArchiveFS(tarPath)
	testutil.AssertNoError(t, err, "NewArchiveFS tar")
	defer fs.Close()

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
		defer r.Close()

		data, err := io.ReadAll(r)
		testutil.AssertNoError(t, err, "ReadAll")

		if !bytes.Contains(data, []byte("Tar archive readme.")) {
			t.Errorf("content mismatch")
		}
	})
}
