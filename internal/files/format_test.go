package files

import (
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 K"},
		{1024 * 1024, "1.0 M"},
		{1024 * 1024 * 1024, "1.0 G"},
	}

	for _, tt := range tests {
		result := FormatSize(tt.bytes, 0)
		if result != tt.expected {
			t.Errorf("FormatSize(%d, 0) = %s; want %s", tt.bytes, result, tt.expected)
		}
	}

	// Test Full format
	if res := FormatSize(1024, 1); res != "1.0 KB" {
		t.Errorf("Expected 1.0 KB, got %s", res)
	}

	// Test Bytes format
	if res := FormatSize(1024, 2); res != "1024 B" {
		t.Errorf("Expected 1024 B, got %s", res)
	}
}
