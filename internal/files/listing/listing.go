package listing

import (
	"context"
	"fm/internal/files"
	"fm/internal/files/sorting"
	"strings"
)

// Load reads the contents of the specified directory path.
// It returns a sorted slice of files.Items based on the provided sorting.SortMode and visibility preferences.
// If individual file entries fail to load, they are skipped and the load continues.
func Load(ctx context.Context, fs files.FileSystem, path string, mode sorting.SortMode, showHidden bool, gitStatuses map[string]string) ([]files.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var items []files.Item

	if path != "/" && path != fs.Separator() {
		items = append(items, files.Item{
			Name:  "↑ ..",
			IsDir: true,
			IsUp:  true,
		})
	}

	entries, err := fs.ReadDir(ctx, path)
	if err != nil {
		return items, err
	}

	// Track seen files to identify ghosts later
	seenOnDisk := make(map[string]bool)

	for _, f := range entries {
		if !showHidden && strings.HasPrefix(f.Name(), ".") {
			continue
		}
		seenOnDisk[f.Name()] = true

		size := f.Size()
		if f.IsDir() {
			size = -1
		}

		fPath := fs.Join(path, f.Name())
		fMode := f.Mode()

		// Simplified permission check: check if user, group, or others have R/W bits
		canRead := (fMode.Perm() & 0444) != 0
		canWrite := (fMode.Perm() & 0222) != 0

		items = append(items, files.Item{
			Name:      f.Name(),
			Path:      fPath,
			IsDir:     f.IsDir(),
			GitStatus: gitStatuses[f.Name()],
			Size:      size,
			Mode:      fMode,
			MTime:     f.ModTime(),
			CanRead:   canRead,
			CanWrite:  canWrite,
		})
	}

	// Add Ghost Entries (Deleted files in Git that are NOT on disk)
	for name, status := range gitStatuses {
		if !seenOnDisk[name] && status == "D" {
			items = append(items, files.Item{
				Name:      name,
				Path:      fs.Join(path, name),
				IsDir:     false,
				GitStatus: "D",
				IsGhost:   true,
				Size:      0,
			})
		}
	}

	// Sort items using unified sorting.SortItems, skipping ".." entry if present
	sorting.SortItems(items, mode, true)

	return items, nil
}
