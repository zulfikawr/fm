package list

import (
	"testing"

	"fm/internal/files/core"
)

func TestRenderMarker(t *testing.T) {
	props := Props{SelectMode: true}

	// Test Selected
	item := core.Item{Selected: true}
	res := renderMarker(props, item)
	if res != "[x] " {
		t.Errorf("Expected [x], got %s", res)
	}

	// Test Unselected
	item.Selected = false
	res = renderMarker(props, item)
	if res != "[ ] " {
		t.Errorf("Expected [ ], got %s", res)
	}

	// Test Up dir
	item.IsUp = true
	res = renderMarker(props, item)
	if res != "    " {
		t.Errorf("Expected empty space for up dir, got %s", res)
	}
}
