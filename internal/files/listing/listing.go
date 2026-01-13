package listing

import (
	"context"
	"fm/internal/files/core"
	"fm/internal/files/sorting"
	"strings"
)

// Load reads the contents of the specified directory path.
// It returns a sorted slice of core.Items based on the provided sorting.SortMode and visibility preferences.
// If individual file entries fail to load, they are skipped and the load continues.
func Load(ctx context.Context, fs core.FileSystem, path string, mode sorting.SortMode, showHidden bool, gitStatuses map[string]string) ([]core.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var items []core.Item

	if path != "/" && path != fs.Separator() {
		items = append(items, core.Item{
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

		fPath := fs.Join(path, f.Name())
		items = append(items, core.NewItem(f, fPath, gitStatuses[f.Name()]))
	}

	// Add Ghost Entries (Deleted files in Git that are NOT on disk)
	for name, status := range gitStatuses {
		if !seenOnDisk[name] && status == "D" {
			items = append(items, core.Item{
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
