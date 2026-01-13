package state

import (
	"testing"

	"fm/internal/files/core"
)

func TestNavigationState(t *testing.T) {
	n := &NavigationState{
		SelectedPaths: make(map[string]bool),
	}

	// Test Select
	n.Select("/file1")
	if n.SelectedCount != 1 || !n.SelectedPaths["/file1"] {
		t.Error("Expected /file1 to be selected")
	}

	// Test Deselect
	n.Deselect("/file1")
	if n.SelectedCount != 0 || n.SelectedPaths["/file1"] {
		t.Error("Expected /file1 to be deselected")
	}

	// Test Toggle
	n.ToggleSelection("/file2")
	if n.SelectedCount != 1 || !n.SelectedPaths["/file2"] {
		t.Error("Expected /file2 to be toggled on")
	}
	n.ToggleSelection("/file2")
	if n.SelectedCount != 0 || n.SelectedPaths["/file2"] {
		t.Error("Expected /file2 to be toggled off")
	}

	// Test Clear
	n.Select("/file3")
	n.Items = []core.Item{{Name: "f3", Selected: true}}
	n.ClearSelection()
	if n.SelectedCount != 0 || len(n.SelectedPaths) != 0 {
		t.Error("Expected selection to be cleared")
	}
}

func TestModelClearSelection(t *testing.T) {
	m := &Model{}
	m.UI.SelectMode = true
	m.Navigation.Select("/f1")

	m.ClearSelection()
	if m.UI.SelectMode {
		t.Error("Expected SelectMode false")
	}
	if m.Navigation.SelectedCount != 0 {
		t.Error("Expected SelectedCount 0")
	}
}
