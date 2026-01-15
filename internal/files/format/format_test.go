package format

import (
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		idx      int
		expected string
	}{
		{100, 0, "100 B"},
		{1024, 0, "1.0 K"},
		{1024, 1, "1.0 KB"},
		{1024, 2, "1024 B"},
		{1024 * 1024, 0, "1.0 M"},
		{1024 * 1024, 1, "1.0 MB"},
		{-1, 0, ""},
	}

	for _, tt := range tests {
		result := FormatSize(tt.bytes, tt.idx)
		if result != tt.expected {
			t.Errorf("FormatSize(%d, %d) = %q, want %q", tt.bytes, tt.idx, result, tt.expected)
		}
	}
}
