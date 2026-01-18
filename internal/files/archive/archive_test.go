package archive

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"fm/internal/testutil"
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

	t.Run("ReadDir root", func(t *testing.T) {
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

	t.Run("ReadDir subfolder", func(t *testing.T) {
		entries, err := fs.ReadDirEntries(ctx, "/folder")
		testutil.AssertNoError(t, err, "ReadDirEntries /folder")

		if len(entries) != 1 || entries[0].Name() != "file.txt" {
			t.Errorf("expected file.txt in /folder, got %v", entries)
		}
	})

	t.Run("Stat file", func(t *testing.T) {
		info, err := fs.Stat(ctx, "/readme.txt")
		testutil.AssertNoError(t, err, "Stat /readme.txt")
		if info.IsDir() {
			t.Errorf("readme.txt should not be a directory")
		}
	})

	t.Run("Open and Read", func(t *testing.T) {
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
