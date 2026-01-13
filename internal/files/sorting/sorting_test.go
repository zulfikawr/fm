package sorting

import (
	"fm/internal/files/core"
	"fmt"
	"pgregory.net/rapid"
	"strings"
	"testing"
	"time"
)

// GenItem generates random Item values for property testing
func GenItem() *rapid.Generator[core.Item] {
	return rapid.Custom(func(t *rapid.T) core.Item {
		return core.Item{
			Name:  rapid.String().Draw(t, "name"),
			IsDir: rapid.Bool().Draw(t, "isDir"),
			Size:  rapid.Int64Range(0, 1<<40).Draw(t, "size"),
			MTime: time.Unix(rapid.Int64Range(0, time.Now().Unix()).Draw(t, "mtime"), 0),
		}
	})
}

// GenSortMode generates random SortMode values
func GenSortMode() *rapid.Generator[SortMode] {
	return rapid.Map(rapid.IntRange(0, 6), func(i int) SortMode {
		return SortMode(i)
	})
}

func TestSortItems_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 1: Sorting Correctness
	rapid.Check(t, func(t *rapid.T) {
		items := rapid.SliceOf(GenItem()).Draw(t, "items")
		mode := GenSortMode().Draw(t, "mode")

		// Make a copy to sort
		sorted := make([]core.Item, len(items))
		copy(sorted, items)

		SortItems(sorted, mode, false)

		// Verify ordering
		for i := 1; i < len(sorted); i++ {
			if !isCorrectOrder(sorted[i-1], sorted[i], mode) {
				msg := fmt.Sprintf("Items at %d and %d are not correctly ordered for mode %v\nItem 1: %+v\nItem 2: %+v",
					i-1, i, mode, sorted[i-1], sorted[i])
				t.Error(msg)
			}
		}
	})
}

func isCorrectOrder(a, b core.Item, mode SortMode) bool {
	switch mode {
	case SortName:
		return strings.ToLower(a.Name) <= strings.ToLower(b.Name)
	case SortNameDesc:
		return strings.ToLower(a.Name) >= strings.ToLower(b.Name)
	case SortNewest:
		return a.MTime.After(b.MTime) || a.MTime.Equal(b.MTime)
	case SortOldest:
		return a.MTime.Before(b.MTime) || a.MTime.Equal(b.MTime)
	case SortSizeDesc:
		return a.Size >= b.Size
	case SortSizeAsc:
		return a.Size <= b.Size
	default: // SortDefault - Directories first, then alphabetical
		if a.IsDir != b.IsDir {
			if a.IsDir {
				return true
			}
			return false
		}
		return strings.ToLower(a.Name) <= strings.ToLower(b.Name)
	}
}

func TestSortModeString(t *testing.T) {
	tests := []struct {
		mode     SortMode
		expected string
	}{
		{SortDefault, "[ ⇅ ] Default"},
		{SortName, "[ A-Z ] Name (Asc)"},
		{SortNameDesc, "[ Z-A ] Name (Desc)"},
		{SortNewest, "[ ↓ ] Newest"},
		{SortOldest, "[ ↑ ] Oldest"},
		{SortSizeDesc, "[ ▼ ] Size (Lrg)"},
		{SortSizeAsc, "[ ▲ ] Size (Sml)"},
		{SortMode(99), "[ ? ] Unknown"},
	}

	for _, tt := range tests {
		if tt.mode.String() != tt.expected {
			t.Errorf("SortMode(%d).String() = %s; want %s", tt.mode, tt.mode.String(), tt.expected)
		}
	}
}
