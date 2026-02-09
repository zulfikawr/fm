package core

import (
	"os"
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/files/format"
)

// Item represents a single file or directory entry in the file manager.
type Item struct {
	Name     string
	Path     string
	IsDir    bool
	Metadata ItemMetadata
	Display  ItemDisplay
	State    ItemState
}

type ItemMetadata struct {
	Size     int64
	Mode     os.FileMode
	MTime    time.Time
	CanRead  bool
	CanWrite bool
}

type ItemDisplay struct {
	FormattedSize string
	FormattedDate string
	GitStatus     string
	IsGhost       bool
}

type ItemState struct {
	Selected    bool
	IsUp        bool
	HasMetadata bool
	SearchKey   string
}

// NewItemFromDirEntry creates a skeleton Item from os.DirEntry
func NewItemFromDirEntry(d os.DirEntry, path string, gitStatus string) Item {
	// Try to get info if it's already cached/available (often true for some systems or SFTP)
	if info, err := d.Info(); err == nil {
		return NewItem(info, path, gitStatus)
	}

	return Item{
		Name:  d.Name(),
		Path:  path,
		IsDir: d.IsDir(),
		Display: ItemDisplay{
			GitStatus: gitStatus,
		},
		State: ItemState{
			HasMetadata: false,
			SearchKey:   strings.ToLower(d.Name()),
		},
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
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
		Metadata: ItemMetadata{
			Size:     size,
			Mode:     mode,
			MTime:    info.ModTime(),
			CanRead:  canRead,
			CanWrite: canWrite,
		},
		Display: ItemDisplay{
			GitStatus: gitStatus,
		},
		State: ItemState{
			HasMetadata: true,
			SearchKey:   strings.ToLower(info.Name()),
		},
	}
}

// IsArchive returns true if the item is an archive format supported for internal browsing.
func (item *Item) IsArchive() bool {
	if item.IsDir {
		return false
	}
	ext := strings.ToLower(item.Name)
	for _, aExt := range []string{".zip", ".tar", ".gz", ".tgz"} {
		if strings.HasSuffix(ext, aExt) {
			return true
		}
	}
	return false
}

// IsImage returns true if the item is a known image format.
func (item *Item) IsImage() bool {
	if item.IsDir {
		return false
	}
	ext := strings.ToLower(item.Name)
	for _, iExt := range []string{".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", ".bmp", ".ico"} {
		if strings.HasSuffix(ext, iExt) {
			return true
		}
	}
	return false
}

// UpdateFormatting updates the display strings for the item
func (item *Item) UpdateFormatting(sizeFormatIdx, dateFormatIdx int) {
	if item.State.IsUp {
		return
	}

	// Hide size and date for deleted files
	if item.Display.GitStatus == "D" && !item.IsDir {
		item.Display.FormattedSize = ""
		item.Display.FormattedDate = ""
		return
	}

	item.Display.FormattedSize = format.FormatSize(item.Metadata.Size, sizeFormatIdx)

	// Check for "zero" dates. Some filesystems use 1970 or 1980 as a default/null value.
	isSuspiciouslyOld := item.Metadata.MTime.Year() <= 1980
	if dateFormatIdx < len(format.DateFormats) && !item.Metadata.MTime.IsZero() && !isSuspiciouslyOld {
		layout := format.DateFormats[dateFormatIdx].Layout
		item.Display.FormattedDate = item.Metadata.MTime.Format(layout)
	} else {
		item.Display.FormattedDate = ""
	}
}
