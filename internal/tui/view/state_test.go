package view

import (
	"testing"

	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"
)

func TestGetViewState(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Display.Width = 100
	m.Display.Height = 50

	s := GetViewState(m)

	if s.Width != 100 {
		t.Errorf("Expected width 100, got %d", s.Width)
	}
	if s.Height != 50 {
		t.Errorf("Expected height 50, got %d", s.Height)
	}

	// Test auth mode (should NOT hide remote info anymore)
	m.Inputs.Mode = state.InputAuth
	m.Remote.Host = "visible-host"
	s = GetViewState(m)
	if s.RemoteHost != "visible-host" {
		t.Error("Expected remote host to be visible in auth mode")
	}
}
