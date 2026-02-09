package listing

import (
	"context"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/files/sorting"
)

// LoadSkeleton reads only the directory entries without full stats.
func LoadSkeleton(ctx context.Context, opts LoadOptions) ([]core.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var items []core.Item

	if !core.IsRoot(opts.FS, opts.Path) {
		items = append(items, core.Item{
			Name:  "↑ ..",
			IsDir: true,
			State: core.ItemState{
				IsUp:      true,
				SearchKey: "..",
			},
			Path: core.GetParent(opts.FS, opts.Path),
		})
	}

	entries, err := opts.FS.ReadDirEntries(ctx, opts.Path)
	if err != nil {
		return items, errors.WrapErrorWithPath(err, "LoadSkeleton", opts.Path)
	}

	// Track seen files to identify ghosts later
	seenOnDisk := make(map[string]bool)

	for _, d := range entries {
		if !opts.ShowHidden && strings.HasPrefix(d.Name(), ".") {
			continue
		}
		seenOnDisk[d.Name()] = true

		fPath := opts.FS.Join(opts.Path, d.Name())
		items = append(items, core.NewItemFromDirEntry(d, fPath, opts.GitStatuses[d.Name()]))
	}

	// Add Ghost Entries (Deleted files in Git that are NOT on disk)
	for name, status := range opts.GitStatuses {
		if !seenOnDisk[name] && status == "D" {
			items = append(items, core.Item{
				Name:  name,
				Path:  opts.FS.Join(opts.Path, name),
				IsDir: false,
				Display: core.ItemDisplay{
					GitStatus: "D",
					IsGhost:   true,
				},
				Metadata: core.ItemMetadata{
					Size: 0,
				},
				State: core.ItemState{
					HasMetadata: true, // Ghost items don't have disk info, they are their own metadata
					SearchKey:   strings.ToLower(name),
				},
			})
		}
	}

	return items, nil
}

// Load reads the contents of the specified directory path with full stats.
func Load(ctx context.Context, opts LoadOptions) ([]core.Item, error) {
	items, err := LoadSkeleton(ctx, opts)
	if err != nil {
		return items, err // LoadSkeleton already wraps
	}

	// For standard Load, we still want full metadata for everything (backward compatibility)
	for i := range items {
		if items[i].State.IsUp || items[i].Display.IsGhost || items[i].State.HasMetadata {
			continue
		}
		info, statErr := opts.FS.Stat(ctx, items[i].Path)
		if statErr == nil {
			items[i] = core.NewItem(info, items[i].Path, items[i].Display.GitStatus)
		}
	}

	// Sort items using unified sorting.SortItems, skipping ".." entry if present
	sorting.SortItems(items, opts.SortMode, true)

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

	maxMTime := item.Metadata.MTime
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			if info.ModTime().After(maxMTime) {
				maxMTime = info.ModTime()
			}
		}
	}
	item.Metadata.MTime = maxMTime
}
