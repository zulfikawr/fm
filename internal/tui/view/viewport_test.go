package view

import (
	"testing"

	"fm/internal/config"
	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"
)

func TestCalculateViewportHeight(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Display.Height = 24
	m.Config.ShowHeader = true
	m.UI.SettingsOpen = false

	h := CalculateViewportHeight(m)
	// 24 - 2 (app header/footer) - 3 (list header) = 19
	if h != 19 {
		t.Errorf("Expected 19, got %d", h)
	}

	m.Config.ShowHeader = false
	h = CalculateViewportHeight(m)
	// 24 - 2 = 22
	if h != 22 {
		t.Errorf("Expected 22, got %d", h)
	}
}

func TestGetViewportHeight(t *testing.T) {
	s := &ViewState{
		ViewportHeight: 15,
	}
	if GetViewportHeight(s) != 15 {
		t.Error("Expected explicit viewport height")
	}

	s.ViewportHeight = 0
	s.Height = 20
	s.Config = &config.Config{ShowHeader: false}
	s.UI = &state.UIState{SettingsOpen: false}
	if GetViewportHeight(s) <= 0 {
		t.Error("Expected calculated viewport height")
	}
}

func TestCalculateViewportHeightFromState(t *testing.T) {
	s := &ViewState{
		Height: 24,
		Config: &config.Config{ShowHeader: true},
		UI:     &state.UIState{SettingsOpen: false},
	}
	h := CalculateViewportHeightFromState(s)
	if h != 19 {
		t.Errorf("Expected 19, got %d", h)
	}
}
