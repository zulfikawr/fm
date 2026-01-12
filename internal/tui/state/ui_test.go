package state

import "testing"

func TestUIState(t *testing.T) {
	s := &UIState{}

	t.Run("Start Input", func(t *testing.T) {
		s.StartInput()
		if !s.InputActive {
			t.Error("Expected InputActive to be true")
		}
	})

	t.Run("Stop Input", func(t *testing.T) {
		s.StopInput()
		if s.InputActive {
			t.Error("Expected InputActive to be false")
		}
	})

	t.Run("Start Confirming", func(t *testing.T) {
		s.StartConfirming()
		if !s.Confirming {
			t.Error("Expected Confirming to be true")
		}
		if s.InputActive {
			t.Error("Expected InputActive to be false")
		}
	})

	t.Run("Toggle Settings", func(t *testing.T) {
		s.ToggleSettings()
		if !s.SettingsOpen {
			t.Error("Expected SettingsOpen to be true")
		}
		if s.InputActive {
			t.Error("Expected InputActive to be false")
		}
	})
}
