package state

import "testing"

func TestUIState(t *testing.T) {
	s := &UIState{}

	s.StartInput()
	if !s.InputActive {
		t.Error("Expected InputActive to be true")
	}

	s.StopInput()
	if s.InputActive {
		t.Error("Expected InputActive false after stop")
	}

	// Test Confirming
	s.StartConfirming()
	if !s.Confirming {
		t.Error("Expected Confirming true")
	}
	s.StopConfirming()
	if s.Confirming {
		t.Error("Expected Confirming false")
	}

	// Test Settings
	s.ToggleSettings()
	if !s.SettingsOpen {
		t.Error("Expected SettingsOpen true")
	}
	s.ToggleSettings()
	if s.SettingsOpen {
		t.Error("Expected SettingsOpen false")
	}
}
