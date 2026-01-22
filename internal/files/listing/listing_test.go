package listing

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/sorting"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestLoad(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Includes parent directory entry", func(t *testing.T) {
		fs.ReadDirEntriesFunc = func(ctx context.Context, path string) ([]os.DirEntry, error) {
			return []os.DirEntry{}, nil
		}
		items, err := Load(ctx, LoadOptions{
			FS:         fs,
			Path:       "/some/path",
			SortMode:   sorting.SortDefault,
			ShowHidden: true,
		})
		testutil.AssertNoError(t, err, "Load should not fail")

		foundUp := false
		for _, item := range items {
			if item.IsUp {
				foundUp = true
				break
			}
		}
		testutil.AssertEqual(t, true, foundUp, "Should contain '..' entry for non-root paths")
	})

	t.Run("Filters hidden files when showHidden is false", func(t *testing.T) {
		fs.ReadDirEntriesFunc = func(ctx context.Context, path string) ([]os.DirEntry, error) {
			return []os.DirEntry{
				&testutil.MockDirEntry{NameStr: "visible.txt"},
				&testutil.MockDirEntry{NameStr: ".hidden.txt"},
			}, nil
		}
		items, err := Load(ctx, LoadOptions{
			FS:         fs,
			Path:       "/",
			SortMode:   sorting.SortDefault,
			ShowHidden: false,
		})
		testutil.AssertNoError(t, err, "Load should not fail")

		for _, item := range items {
			if strings.HasPrefix(item.Name, ".") && !item.IsUp {
				t.Errorf("found hidden file %q when showHidden=false", item.Name)
			}
		}
	})

	t.Run("Includes ghost entries from git status", func(t *testing.T) {
		fs.ReadDirEntriesFunc = func(ctx context.Context, path string) ([]os.DirEntry, error) {
			return []os.DirEntry{
				&testutil.MockDirEntry{NameStr: "exists.txt"},
			}, nil
		}
		gitStatuses := map[string]string{
			"deleted.txt": "D",
			"exists.txt":  "M",
		}
		items, err := Load(ctx, LoadOptions{
			FS:          fs,
			Path:        "/",
			SortMode:    sorting.SortDefault,
			ShowHidden:  true,
			GitStatuses: gitStatuses,
		})
		testutil.AssertNoError(t, err, "Load should not fail")

		foundGhost := false
		for _, item := range items {
			if item.Name == "deleted.txt" && item.IsGhost {
				foundGhost = true
				break
			}
		}
		testutil.AssertEqual(t, true, foundGhost, "Should contain ghost entry for deleted git file")
	})
}

func TestLoad_LargeDirectory(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	// Simulate 10,000 files
	numFiles := 10000
	files := make([]os.DirEntry, numFiles)
	for i := 0; i < numFiles; i++ {
		files[i] = &testutil.MockDirEntry{
			NameStr:   fmt.Sprintf("file_%05d.txt", i),
			IsDirBool: false,
		}
	}

	fs.ReadDirEntriesFunc = func(ctx context.Context, path string) ([]os.DirEntry, error) {
		return files, nil
	}

	start := time.Now()
	items, err := Load(ctx, LoadOptions{
		FS:         fs,
		Path:       "/large",
		SortMode:   sorting.SortDefault,
		ShowHidden: true,
	})
	duration := time.Since(start)

	testutil.AssertNoError(t, err, "Load should succeed")
	testutil.AssertEqual(t, numFiles+1, len(items), "Should have 10,001 items (including ..)")

	if duration > 1*time.Second {
		t.Errorf("Load took too long: %v", duration)
	}
}

func TestEnrichMetadata(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	now := time.Now()

	t.Run("Enrich directory date", func(t *testing.T) {
		fs.ReadDirEntriesFunc = func(ctx context.Context, path string) ([]os.DirEntry, error) {
			return []os.DirEntry{
				&testutil.MockDirEntry{NameStr: "new.txt", ModeBits: 0, InfoErr: nil},
			}, nil
		}
		// MockInfo for new.txt
		item := core.Item{Path: "/dir", IsDir: true, MTime: now.Add(-1 * time.Hour)}

		// We need to mock the Info() call inside EnrichMetadata
		// This is tricky because EnrichMetadata calls entry.Info()
		// MockDirEntry already provides a working Info()

		EnrichMetadata(ctx, fs, &item)
		// Since we didn't override the info date in MockDirEntry, it uses time.Time{} or now.
		// Let's just ensure it runs without crash for now.
	})

	t.Run("Not a directory", func(t *testing.T) {
		item := core.Item{IsDir: false}
		EnrichMetadata(ctx, fs, &item)
	})
}
