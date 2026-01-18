package core

import (
	"os"
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/files/format"
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
	SearchKey     string
	FormattedSize string
	FormattedDate string
	HasMetadata   bool // HasMetadata indicates if Stat/Info has been called
}

// NewItemFromDirEntry creates a skeleton Item from os.DirEntry
func NewItemFromDirEntry(d os.DirEntry, path string, gitStatus string) Item {
	// Try to get info if it's already cached/available (often true for some systems or SFTP)
	if info, err := d.Info(); err == nil {
		return NewItem(info, path, gitStatus)
	}

	return Item{
		Name:        d.Name(),
		Path:        path,
		IsDir:       d.IsDir(),
		GitStatus:   gitStatus,
		HasMetadata: false,
		SearchKey:   strings.ToLower(d.Name()),
	}
}

// NewItem creates a new Item from os.FileInfo
func NewItem(info os.FileInfo, path string, gitStatus string) Item {
	size := info.Size()
	if info.IsDir() {
		size = -1
	}

	mode := info.Mode()
	// Simplified permission check: check if user, group, or others have R/W bits
	canRead := (mode.Perm() & 0o444) != 0
	canWrite := (mode.Perm() & 0o222) != 0

	return Item{
		Name:        info.Name(),
		Path:        path,
		IsDir:       info.IsDir(),
		GitStatus:   gitStatus,
		Size:        size,
		Mode:        mode,
		MTime:       info.ModTime(),
		CanRead:     canRead,
		CanWrite:    canWrite,
		HasMetadata: true,
		SearchKey:   strings.ToLower(info.Name()),
	}
}

// UpdateFormatting updates the display strings for the item
func (item *Item) UpdateFormatting(sizeFormatIdx, dateFormatIdx int) {
	if item.IsUp {
		return
	}

	// Hide size and date for deleted files
	if item.GitStatus == "D" && !item.IsDir {
		item.FormattedSize = ""
		item.FormattedDate = ""
		return
	}

	item.FormattedSize = format.FormatSize(item.Size, sizeFormatIdx)

	// Check for "zero" dates. Some filesystems use 1970 or 1980 as a default/null value.
	isSuspiciouslyOld := item.MTime.Year() <= 1980
	if dateFormatIdx < len(format.DateFormats) && !item.MTime.IsZero() && !isSuspiciouslyOld {
		layout := format.DateFormats[dateFormatIdx].Layout
		item.FormattedDate = item.MTime.Format(layout)
	} else {
		item.FormattedDate = ""
	}
}

// Progress represents the current state of a file operation.
type Progress struct {
	Percent float64
	Label   string
}
