package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SortMode defines the sorting strategy for directory contents.
type SortMode int

const (
	SortDefault SortMode = iota
	SortName
	SortNameDesc
	SortNewest
	SortOldest
	SortSizeDesc
	SortSizeAsc
)

// String returns a human-readable representation of the sort mode.
func (s SortMode) String() string {
	switch s {
	case SortDefault:
		return "[ ⇅ ] Default"
	case SortName:
		return "[ A-Z ] Name (Asc)"
	case SortNameDesc:
		return "[ Z-A ] Name (Desc)"
	case SortNewest:
		return "[ ↓ ] Newest"
	case SortOldest:
		return "[ ↑ ] Oldest"
	case SortSizeDesc:
		return "[ ▼ ] Size (Lrg)"
	case SortSizeAsc:
		return "[ ▲ ] Size (Sml)"
	default:
		return "[ ? ] Unknown"
	}
}

// Delete removes a file or directory recursively.
func Delete(path string) error {
	return os.RemoveAll(path)
}

// Rename moves or renames a file or directory.
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// Copy copies a file or directory recursively from src to dst.
func Copy(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// Item represents a single file or directory entry in the file manager.
type Item struct {
	Name      string
	Path      string
	IsDir     bool
	Selected  bool   // Selected tracks if the item is marked for bulk operations.
	GitStatus string // GitStatus represents the porcelain status (e.g., "M", "A", "D").
	IsGhost   bool   // IsGhost indicates a file that exists in Git but not on disk (Deleted).
	Size      int64
	Mode      os.FileMode
	MTime     os.FileInfo
	IsUp      bool // IsUp indicates if this item represents the parent directory ("..").
}

// Load reads the contents of the specified directory path.
// It returns a sorted slice of Items based on the provided SortMode and visibility preferences.
func Load(path string, mode SortMode, showHidden bool, gitStatuses map[string]string) ([]Item, error) {
	var items []Item

	if path != "/" {
		items = append(items, Item{
			Name:  "..",
			IsDir: true,
			IsUp:  true,
		})
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return items, fmt.Errorf("failed to read directory: %w", err)
	}

	// Track seen files to identify ghosts later
	seenOnDisk := make(map[string]bool)

	// Filter and Sort logic
	var filtered []os.FileInfo
	for _, e := range entries {
		if !showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		filtered = append(filtered, info)
		seenOnDisk[e.Name()] = true
	}

	sort.Slice(filtered, func(i, j int) bool {
		d1, d2 := filtered[i].IsDir(), filtered[j].IsDir()
		if d1 != d2 {
			return d1
		}

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
		default: // SortDefault
			return strings.ToLower(filtered[i].Name()) < strings.ToLower(filtered[j].Name())
		}
	})

	for _, f := range filtered {
		items = append(items, Item{
			Name:      f.Name(),
			Path:      filepath.Join(path, f.Name()),
			IsDir:     f.IsDir(),
			GitStatus: gitStatuses[f.Name()],
			Size:      f.Size(),
			Mode:      f.Mode(),
		})
	}

	// Add Ghost Entries (Deleted files in Git that are NOT on disk)
	for name, status := range gitStatuses {
		if !seenOnDisk[name] && status == "D" {
			items = append(items, Item{
				Name:      name,
				Path:      filepath.Join(path, name),
				IsDir:     false,
				GitStatus: "D",
				IsGhost:   true,
				Size:      0,
			})
		}
	}

	return items, nil
}

// FormatSize converts a byte count into a human-readable string (e.g., "1.5 MB").
func FormatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %c", float64(b)/float64(div), "KMGTPE"[exp])
}
