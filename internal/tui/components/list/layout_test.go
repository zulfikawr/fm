package list

import (
	"testing"
)

func TestCalculateLayout(t *testing.T) {
	props := Props{
		Width:            80,
		ShowSize:         true,
		ShowDateModified: true,
		SelectMode:       true,
		DateLayout:       "2006-01-02",
	}

	layout := CalculateLayout(props)

	if layout.MarkerWidth != 4 {
		t.Errorf("Expected MarkerWidth 4, got %d", layout.MarkerWidth)
	}
	if layout.NameWidth <= 0 {
		t.Error("Expected positive NameWidth")
	}
}
