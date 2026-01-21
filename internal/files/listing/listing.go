package listing

import (
	"context"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/files/sorting"
)

// LoadSkeleton reads only the directory entries without full stats.
func LoadSkeleton(ctx context.Context, fs core.FileSystem, path string, showHidden bool, gitStatuses map[string]string) ([]core.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var items []core.Item

	if path != "/" && path != fs.Separator() {
		items = append(items, core.Item{
			Name:      "↑ ..",
			IsDir:     true,
			IsUp:      true,
			SearchKey: "..",
			Path:      fs.Dir(path),
		})
	}

	entries, err := fs.ReadDirEntries(ctx, path)
	if err != nil {
		return items, errors.WrapErrorWithPath(err, "LoadSkeleton", path)
	}

	// Track seen files to identify ghosts later
	seenOnDisk := make(map[string]bool)

	for _, d := range entries {
		if !showHidden && strings.HasPrefix(d.Name(), ".") {
			continue
		}
		seenOnDisk[d.Name()] = true

		fPath := fs.Join(path, d.Name())
		items = append(items, core.NewItemFromDirEntry(d, fPath, gitStatuses[d.Name()]))
	}

	// Add Ghost Entries (Deleted files in Git that are NOT on disk)
	for name, status := range gitStatuses {
		if !seenOnDisk[name] && status == "D" {
			items = append(items, core.Item{
				Name:        name,
				Path:        fs.Join(path, name),
				IsDir:       false,
				GitStatus:   "D",
				IsGhost:     true,
				Size:        0,
				HasMetadata: true, // Ghost items don't have disk info, they are their own metadata
				SearchKey:   strings.ToLower(name),
			})
		}
	}

	return items, nil
}

// Load reads the contents of the specified directory path with full stats.
func Load(ctx context.Context, fs core.FileSystem, path string, mode sorting.SortMode, showHidden bool, gitStatuses map[string]string) ([]core.Item, error) {
	items, err := LoadSkeleton(ctx, fs, path, showHidden, gitStatuses)
	if err != nil {
		return items, err // LoadSkeleton already wraps
	}

	// For standard Load, we still want full metadata for everything (backward compatibility)
	for i := range items {
		if items[i].IsUp || items[i].IsGhost || items[i].HasMetadata {
			continue
		}
		info, statErr := fs.Stat(ctx, items[i].Path)
		if statErr == nil {
			items[i] = core.NewItem(info, items[i].Path, items[i].GitStatus)
		}
	}

	// Sort items using unified sorting.SortItems, skipping ".." entry if present
	sorting.SortItems(items, mode, true)

	return items, nil
}

// EnrichMetadata recursively updates directory metadata (like MTime) based on its contents.
func EnrichMetadata(ctx context.Context, fs core.FileSystem, item *core.Item) {
	if !item.IsDir {
		return
	}

	entries, err := fs.ReadDirEntries(ctx, item.Path)
	if err != nil {
		return
	}

	maxMTime := item.MTime
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			if info.ModTime().After(maxMTime) {
				maxMTime = info.ModTime()
			}
		}
	}
	item.MTime = maxMTime
}
