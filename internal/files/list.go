package files

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Load reads the contents of the specified directory path.
// It returns a sorted slice of Items based on the provided SortMode and visibility preferences.
func Load(fs FileSystem, path string, mode SortMode, showHidden bool, gitStatuses map[string]string) ([]Item, error) {
	var items []Item

	if path != "/" && path != fs.Separator() {
		items = append(items, Item{
			Name:  "↑ ..",
			IsDir: true,
			IsUp:  true,
		})
	}

	entries, err := fs.ReadDir(path)
	if err != nil {
		return items, fmt.Errorf("failed to read directory: %w", err)
	}

	// Track seen files to identify ghosts later
	seenOnDisk := make(map[string]bool)

	var filtered []os.FileInfo
	for _, e := range entries {
		if !showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		filtered = append(filtered, e)
		seenOnDisk[e.Name()] = true
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

		items = append(items, Item{
			Name:      f.Name(),
			Path:      fs.Join(path, f.Name()),
			IsDir:     f.IsDir(),
			GitStatus: gitStatuses[f.Name()],
			Size:      size,
			Mode:      f.Mode(),
			MTime:     f.ModTime(),
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

// GetDirSize calculates the total size of a directory recursively.
func GetDirSize(fs FileSystem, path string) int64 {
	var size int64

	var walk func(string)
	walk = func(currPath string) {
		entries, err := fs.ReadDir(currPath)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				walk(fs.Join(currPath, e.Name()))
			} else {
				size += e.Size()
			}
		}
	}

	walk(path)
	return size
}
