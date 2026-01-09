package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSortModeString(t *testing.T) {
	tests := []struct {
		mode     SortMode
		expected string
	}{
		{SortDefault, "[ ⇅ ] Default"},
		{SortName, "[ A-Z ] Name (Asc)"},
		{SortSizeDesc, "[ ▼ ] Size (Lrg)"},
		{SortMode(99), "[ ? ] Unknown"},
	}

	for _, tt := range tests {
		if tt.mode.String() != tt.expected {
			t.Errorf("SortMode(%d).String() = %s; want %s", tt.mode, tt.mode.String(), tt.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 K"},
		{1024 * 1024, "1.0 M"},
		{1024 * 1024 * 1024, "1.0 G"},
	}

	for _, tt := range tests {
		result := FormatSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatSize(%d) = %s; want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "a_file.txt"), []byte("test"), 0644)

	t.Run("Default Sort", func(t *testing.T) {
		items, err := Load(tmpDir, SortDefault, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if items[0].Name != ".." {
			t.Errorf("Expected first item to be '..', got %s", items[0].Name)
		}
		if items[1].Name != "dir1" {
			t.Errorf("Expected second item to be 'dir1', got %s", items[1].Name)
		}
	})

	t.Run("Size Sort", func(t *testing.T) {
		items, err := Load(tmpDir, SortSizeDesc, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		// Expectation: .. (up), dir1 (dir), file1.txt (5 bytes), a_file.txt (4 bytes)
		if items[2].Name != "file1.txt" {
			t.Errorf("Expected largest file to be file1.txt, got %s", items[2].Name)
		}
	})

	t.Run("Name Sort", func(t *testing.T) {
		items, err := Load(tmpDir, SortNameDesc, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		// Expectation: .. (up), dir1 (dir), file1.txt, a_file.txt
		// In NameDesc: file1.txt comes before a_file.txt
		if items[2].Name != "file1.txt" {
			t.Errorf("Expected first file to be file1.txt in NameDesc, got %s", items[2].Name)
		}
	})

	t.Run("Newest Sort", func(t *testing.T) {
		// Create a newer file
		newFile := filepath.Join(tmpDir, "new.txt")
		os.WriteFile(newFile, []byte("new"), 0644)

		// Explicitly set a newer modification time to avoid flakes
		future := time.Now().Add(1 * time.Hour)
		os.Chtimes(newFile, future, future)

		items, err := Load(tmpDir, SortNewest, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		// Expectation: .. (up), dir1 (dir), new.txt (newest)
		if items[2].Name != "new.txt" {
			t.Errorf("Expected newest file to be new.txt, got %s", items[2].Name)
		}
	})

	t.Run("Oldest Sort", func(t *testing.T) {
		items, err := Load(tmpDir, SortOldest, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		// Expectation: .. (up), dir1 (dir), oldest file
		// dir1 was created first, but it's a directory.
		// Among files, file1.txt and a_file.txt were created early.
		if items[2].Name != "file1.txt" && items[2].Name != "a_file.txt" {
			t.Errorf("Expected one of the original files, got %s", items[2].Name)
		}
	})

	t.Run("Size Asc Sort", func(t *testing.T) {
		items, err := Load(tmpDir, SortSizeAsc, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		// Expectation: .. (up), dir1 (dir), smallest file (new.txt is 3 bytes)
		if items[2].Name != "new.txt" {
			t.Errorf("Expected smallest file to be new.txt, got %s", items[2].Name)
		}
	})

	t.Run("Show Hidden", func(t *testing.T) {
		os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644)
		items, err := Load(tmpDir, SortDefault, true, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		found := false
		for _, item := range items {
			if item.Name == ".hidden" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find hidden file when showHidden is true")
		}
	})
}

func TestFileOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-ops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "src.txt")
	content := []byte("hello world")
	os.WriteFile(srcFile, content, 0644)

	t.Run("Copy File", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "dst.txt")
		if err := Copy(srcFile, dstFile); err != nil {
			t.Fatalf("Copy failed: %v", err)
		}

		got, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(content) {
			t.Errorf("Expected content %s, got %s", string(content), string(got))
		}
	})

	t.Run("Rename File", func(t *testing.T) {
		oldPath := srcFile
		newPath := filepath.Join(tmpDir, "renamed.txt")
		if err := Rename(oldPath, newPath); err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		if _, err := os.Stat(oldPath); err == nil {
			t.Error("Old file still exists after rename")
		}
		if _, err := os.Stat(newPath); err != nil {
			t.Error("New file does not exist after rename")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		path := filepath.Join(tmpDir, "todelete.txt")
		os.WriteFile(path, []byte("delete me"), 0644)
		if err := Delete(path); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if _, err := os.Stat(path); err == nil {
			t.Error("File still exists after delete")
		}
	})

	t.Run("Copy Non-existent", func(t *testing.T) {
		err := Copy(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"))
		if err == nil {
			t.Error("Expected error when copying non-existent file")
		}
	})

	t.Run("Rename Non-existent", func(t *testing.T) {
		err := Rename(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "out"))
		if err == nil {
			t.Error("Expected error when renaming non-existent file")
		}
	})

	t.Run("Copy Dir Recursive", func(t *testing.T) {
		srcDir := filepath.Join(tmpDir, "recursive_src")
		subDir := filepath.Join(srcDir, "sub")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "inner.txt"), []byte("inner"), 0644)

		dstDir := filepath.Join(tmpDir, "recursive_dst")
		if err := Copy(srcDir, dstDir); err != nil {
			t.Fatalf("Recursive copy failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(dstDir, "sub", "inner.txt")); err != nil {
			t.Error("Nested file not found in destination")
		}
	})
}

func TestLoadGhostEntries(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-ghost-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock git statuses: "deleted.txt" is marked as Deleted in Git but doesn't exist on disk
	gitStatuses := map[string]string{
		"deleted.txt": "D",
	}

	items, err := Load(tmpDir, SortDefault, false, gitStatuses)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	found := false
	for _, item := range items {
		if item.Name == "deleted.txt" {
			if !item.IsGhost {
				t.Error("Expected deleted.txt to be marked as Ghost")
			}
			if item.GitStatus != "D" {
				t.Errorf("Expected GitStatus D, got %s", item.GitStatus)
			}
			found = true
		}
	}

	if !found {
		t.Error("Expected to find ghost entry 'deleted.txt'")
	}
}
