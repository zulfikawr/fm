package actions

import (
	"testing"

	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"
)

func TestOpenClosePrompt(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Display.Width = 80

	// Test Open
	OpenPrompt(m, state.InputSearch, "initial")
	if !m.UI.InputActive {
		t.Error("Expected InputActive to be true")
	}
	if m.Inputs.Mode != state.InputSearch {
		t.Errorf("Expected mode Search, got %v", m.Inputs.Mode)
	}
	if m.Inputs.ActiveInput.Value() != "initial" {
		t.Errorf("Expected value 'initial', got '%s'", m.Inputs.ActiveInput.Value())
	}

	// Test Close
	ClosePrompt(m)
	if m.UI.InputActive {
		t.Error("Expected InputActive to be false after close")
	}
	if m.Inputs.Mode != state.InputNone {
		t.Error("Expected mode None after close")
	}
	if m.Inputs.ActiveInput.Value() != "" {
		t.Error("Expected value to be cleared")
	}
}

func TestOpenPrompt_Modes(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Display.Width = 80

	modes := []state.InputMode{
		state.InputSearch,
		state.InputRename,
		state.InputGoto,
		state.InputAuth,
	}

	for _, mode := range modes {
		OpenPrompt(m, mode, "")
		if m.Inputs.Mode != mode {
			t.Errorf("Expected mode %v, got %v", mode, m.Inputs.Mode)
		}
	}
}
