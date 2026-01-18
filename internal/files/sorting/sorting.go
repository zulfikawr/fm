package sorting

import (
	"sort"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
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

// Next returns the next sort mode in the cycle.
func (s SortMode) Next() SortMode {
	return (s + 1) % 7 // We have 7 modes (0 to 6)
}

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

// SortItems sorts a slice of Items according to the specified mode
// This is the single source of truth for sorting logic
func SortItems(items []core.Item, mode SortMode, skipFirst bool) {
	startIdx := 0
	if skipFirst && len(items) > 0 && items[0].IsUp {
		startIdx = 1
	}

	if startIdx >= len(items) {
		return
	}

	toSort := items[startIdx:]
	sort.SliceStable(toSort, func(i, j int) bool {
		return compareBySortMode(toSort[i], toSort[j], mode)
	})
}

func compareBySortMode(a, b core.Item, mode SortMode) bool {
	switch mode {
	case SortName:
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case SortNameDesc:
		return strings.ToLower(a.Name) > strings.ToLower(b.Name)
	case SortNewest:
		return a.MTime.After(b.MTime)
	case SortOldest:
		return a.MTime.Before(b.MTime)
	case SortSizeDesc:
		return a.Size > b.Size
	case SortSizeAsc:
		return a.Size < b.Size
	default: // SortDefault - Directories first, then alphabetical
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	}
}
