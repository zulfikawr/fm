package state

import "testing"

func TestDisplayState(t *testing.T) {
	s := &DisplayState{Width: 80, Height: 24}
	if s.Width != 80 || s.Height != 24 {
		t.Error("Expected 80x24")
	}
}
