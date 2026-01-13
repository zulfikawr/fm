package core

import (
	"fm/internal/files/format"
	"os"
	"time"
)

// Item represents a single file or directory entry in the file manager.
type Item struct {
	Name          string
	Path          string
	IsDir         bool
	Selected      bool   // Selected tracks if the item is marked for bulk operations.
	GitStatus     string // GitStatus represents the porcelain status (e.g., "M", "A", "D").
	IsGhost       bool   // IsGhost indicates a file that exists in Git but not on disk (Deleted).
	Size          int64
	Mode          os.FileMode
	MTime         time.Time
	CanRead       bool // CanRead indicates if the current user has read permission
	CanWrite      bool // CanWrite indicates if the current user has write permission
	IsUp          bool // IsUp indicates if this item represents the parent directory ("..").
	FormattedSize string
	FormattedDate string
}

// NewItem creates a new Item from os.FileInfo
func NewItem(info os.FileInfo, path string, gitStatus string) Item {
	size := info.Size()
	if info.IsDir() {
		size = -1
	}

	mode := info.Mode()
	// Simplified permission check: check if user, group, or others have R/W bits
	canRead := (mode.Perm() & 0444) != 0
	canWrite := (mode.Perm() & 0222) != 0

	return Item{
		Name:      info.Name(),
		Path:      path,
		IsDir:     info.IsDir(),
		GitStatus: gitStatus,
		Size:      size,
		Mode:      mode,
		MTime:     info.ModTime(),
		CanRead:   canRead,
		CanWrite:  canWrite,
	}
}

// UpdateFormatting updates the display strings for the item
func (i *Item) UpdateFormatting(sizeFormatIdx, dateFormatIdx int) {
	if i.IsUp {
		return
	}
	i.FormattedSize = format.FormatSize(i.Size, sizeFormatIdx)
	if dateFormatIdx < len(format.DateFormats) {
		layout := format.DateFormats[dateFormatIdx].Layout
		i.FormattedDate = i.MTime.Format(layout)
	}
}

// Progress represents the current state of a file operation.
type Progress struct {
	Percent float64
	Label   string
}
