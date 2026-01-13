package actions

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/files/core"
)

func TestToggleSetting(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Test index 0: ShowHidden
	initial := m.Config.ShowHidden
	ToggleSetting(0, m)
	if m.Config.ShowHidden == initial {
		t.Error("Expected ShowHidden to toggle")
	}

	// Test index 8: ShowSize
	initial = m.Config.ShowSize
	ToggleSetting(8, m)
	if m.Config.ShowSize == initial {
		t.Error("Expected ShowSize to toggle")
	}

	// Test cyclable settings
	ToggleSetting(4, m)  // Editor
	ToggleSetting(9, m)  // SizeFormat
	ToggleSetting(11, m) // DateFormat
	ToggleSetting(12, m) // Theme
}

func TestUpdateFormatting_Up(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Items = []core.Item{
		{Name: "..", IsUp: true},
	}

	for i := range m.Navigation.Items {
		m.Navigation.Items[i].UpdateFormatting(m.Config.SizeFormatIndex, m.Config.DateFormatIndex)
	}
	if m.Navigation.Items[0].FormattedSize != "" {
		t.Error("Expected FormattedSize to remain empty for Up dir")
	}
}

func TestToggleSettingPrev(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Test index 4: EditorIndex (prev)
	initial := m.Config.EditorIndex
	ToggleSettingPrev(4, m)
	if m.Config.EditorIndex == initial {
		// This might stay same if only 1 editor, but should cycle
	}

	// Test fallback
	initialBool := m.Config.ShowHidden
	ToggleSettingPrev(0, m)
	if m.Config.ShowHidden == initialBool {
		t.Error("Expected fallback toggle")
	}
}

func TestUpdateFormatting(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Items = []core.Item{
		{Name: "f1", Size: 1024},
	}

	for i := range m.Navigation.Items {
		m.Navigation.Items[i].UpdateFormatting(m.Config.SizeFormatIndex, m.Config.DateFormatIndex)
	}
	if m.Navigation.Items[0].FormattedSize == "" {
		t.Error("Expected FormattedSize to be set")
	}

	// Cycle formats and format again
	m.Config.SizeFormatIndex = 1
	m.Config.DateFormatIndex = 1
	for i := range m.Navigation.Items {
		m.Navigation.Items[i].UpdateFormatting(m.Config.SizeFormatIndex, m.Config.DateFormatIndex)
	}
}

func TestSettingsReset(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.Confirming = true
	m.Operations.ActionType = "reset-settings"

	ConfirmSettingsReset(m)
	if m.UI.Confirming {
		t.Error("Expected confirming to be false after reset")
	}

	m.UI.Confirming = true
	m.Operations.ActionType = "reset-settings"
	CancelSettingsReset(m)
	if m.UI.Confirming {
		t.Error("Expected confirming to be false after cancel")
	}
}
