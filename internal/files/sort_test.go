package files

import (
	"testing"
)

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
