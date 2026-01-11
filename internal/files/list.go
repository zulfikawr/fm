package files

import (
	"context"
	"os"
	"sort"
	"strings"
)

// Load reads the contents of the specified directory path.
// It returns a sorted slice of Items based on the provided SortMode and visibility preferences.
// If individual file entries fail to load, they are skipped and the load continues.
func Load(ctx context.Context, fs FileSystem, path string, mode SortMode, showHidden bool, gitStatuses map[string]string) ([]Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var items []Item

	if path != "/" && path != fs.Separator() {
		items = append(items, Item{
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

	var filtered []os.FileInfo
	for _, entry := range entries {
		if !showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		filtered = append(filtered, entry)
		seenOnDisk[entry.Name()] = true
	}

	sort.Slice(filtered, func(i, j int) bool {
		switch mode {
		case SortName:
			return strings.ToLower(filtered[i].Name()) < strings.ToLower(filtered[j].Name())
		case SortNameDesc:
			return strings.ToLower(filtered[i].Name()) > strings.ToLower(filtered[j].Name())
		case SortNewest:
			return filtered[i].ModTime().After(filtered[j].ModTime())
		case SortOldest:
			return filtered[i].ModTime().Before(filtered[j].ModTime())
		case SortSizeDesc:
			return filtered[i].Size() > filtered[j].Size()
		case SortSizeAsc:
			return filtered[i].Size() < filtered[j].Size()
		default: // SortDefault - Directories first, then alphabetical
			d1, d2 := filtered[i].IsDir(), filtered[j].IsDir()
			if d1 != d2 {
				return d1
			}
			return strings.ToLower(filtered[i].Name()) < strings.ToLower(filtered[j].Name())
		}
	})

	for _, f := range filtered {
		size := f.Size()
		if f.IsDir() {
			size = -1
		}

		path := fs.Join(path, f.Name())
		mode := f.Mode()

		// Simplified permission check: check if user, group, or others have R/W bits
		// This is an approximation. On Unix, we could use unix.Access, but for SFTP compatibility
		// and cross-platform ease, we check bits or try a dry-run if needed.
		// For now, let's check Mode bits and refine for SFTP if necessary.
		canRead := (mode.Perm() & 0444) != 0
		canWrite := (mode.Perm() & 0222) != 0

		items = append(items, Item{
			Name:      f.Name(),
			Path:      path,
			IsDir:     f.IsDir(),
			GitStatus: gitStatuses[f.Name()],
			Size:      size,
			Mode:      mode,
			MTime:     f.ModTime(),
			CanRead:   canRead,
			CanWrite:  canWrite,
		})
	}

	// Add Ghost Entries (Deleted files in Git that are NOT on disk)
	for name, status := range gitStatuses {
		if !seenOnDisk[name] && status == "D" {
			items = append(items, Item{
				Name:      name,
				Path:      fs.Join(path, name),
				IsDir:     false,
				GitStatus: "D",
				IsGhost:   true,
				Size:      0,
			})
		}
	}

	return items, nil
}
