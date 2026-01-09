package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-list-test")
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

		if items[0].Name != "↑ .." {
			t.Errorf("Expected first item to be '↑ ..', got %s", items[0].Name)
		}
	})

	t.Run("Size Sort", func(t *testing.T) {
		items, err := Load(tmpDir, SortSizeDesc, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		foundAFile := false
		for _, item := range items {
			if item.Name == "file1.txt" {
				if foundAFile {
					t.Error("file1.txt should come before a_file.txt in SizeDesc")
				}
			}
			if item.Name == "a_file.txt" {
				foundAFile = true
			}
		}
	})

	t.Run("Newest Sort", func(t *testing.T) {
		newFile := filepath.Join(tmpDir, "new.txt")
		os.WriteFile(newFile, []byte("new"), 0644)
		future := time.Now().Add(1 * time.Hour)
		os.Chtimes(newFile, future, future)

		items, err := Load(tmpDir, SortNewest, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if items[1].Name != "new.txt" {
			t.Errorf("Expected newest item to be new.txt, got %s", items[1].Name)
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

func TestLoadGhostEntries(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-ghost-test")
	defer os.RemoveAll(tmpDir)

	gitStatuses := map[string]string{"deleted.txt": "D"}
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
			found = true
		}
	}
	if !found {
		t.Error("Expected to find ghost entry 'deleted.txt'")
	}
}

func TestGetDirSize(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-dir-size-test")
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "f1.txt"), []byte("12345"), 0644)
	os.WriteFile(filepath.Join(subDir, "f2.txt"), []byte("123"), 0644)

	size := GetDirSize(tmpDir)
	if size != 8 {
		t.Errorf("Expected size 8, got %d", size)
	}
}
