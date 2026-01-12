package listing

import (
	"context"
	"fm/internal/testutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fm/internal/files/local"
	"fm/internal/files/sorting"
)

func TestLoad(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755)
	testutil.CreateTestFile(t, tmpDir, "file1.txt", "hello")
	testutil.CreateTestFile(t, tmpDir, "a_file.txt", "test")

	t.Run("Default Sort", func(t *testing.T) {
		fs := &local.LocalFS{}
		items, err := Load(context.Background(), fs, tmpDir, sorting.SortDefault, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if items[0].Name != "↑ .." {
			t.Errorf("Expected first item to be '↑ ..', got %s", items[0].Name)
		}
	})

	t.Run("Size Sort", func(t *testing.T) {
		fs := &local.LocalFS{}
		items, err := Load(context.Background(), fs, tmpDir, sorting.SortSizeDesc, false, nil)
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
		newFile := testutil.CreateTestFile(t, tmpDir, "new.txt", "new")
		future := time.Now().Add(1 * time.Hour)
		os.Chtimes(newFile, future, future)

		fs := &local.LocalFS{}
		items, err := Load(context.Background(), fs, tmpDir, sorting.SortNewest, false, nil)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if items[1].Name != "new.txt" {
			t.Errorf("Expected newest item to be new.txt, got %s", items[1].Name)
		}
	})

	t.Run("Show Hidden", func(t *testing.T) {
		testutil.CreateTestFile(t, tmpDir, ".hidden", "hidden")
		fs := &local.LocalFS{}
		items, err := Load(context.Background(), fs, tmpDir, sorting.SortDefault, true, nil)
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
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	gitStatuses := map[string]string{"deleted.txt": "D"}
	fs := &local.LocalFS{}
	items, err := Load(context.Background(), fs, tmpDir, sorting.SortDefault, false, gitStatuses)
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
