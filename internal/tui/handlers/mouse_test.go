package handlers

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/context"
)

func TestHandleMouse_DragSelect(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/")
	m.Config.UI.EnableMouse = true
	m.Display.Height = 20
	m.Display.Width = 80
	m.Navigation.FilteredItems = []core.Item{
		{Name: "file1", Path: "/file1"},
		{Name: "file2", Path: "/file2"},
		{Name: "file3", Path: "/file3"},
	}

	m.Config.UI.ShowHeader = true

	// Let's test drag-to-select starting from empty area
	m.Display.Mouse.IsDragging = false
	HandleMouse(m, tea.MouseMsg{
		X:      10,
		Y:      10, // Empty area
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	HandleMouse(m, tea.MouseMsg{
		X:      10,
		Y:      4, // Drag up to file1
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionMotion,
	})

	if m.Navigation.SelectedCount() == 0 {
		t.Error("Expected items to be selected via drag")
	}
}

func TestHandleMouse_DoubleClick(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/")
	m.Config.UI.EnableMouse = true
	m.Display.Height = 20
	m.Config.UI.ShowHeader = true
	m.Navigation.FilteredItems = []core.Item{
		{Name: "file1", Path: "/file1", IsDir: false},
	}

	// First click
	HandleMouse(m, tea.MouseMsg{
		X:      10,
		Y:      4,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	// Second click (Double click)
	cmd := HandleMouse(m, tea.MouseMsg{
		X:      10,
		Y:      4,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	if cmd == nil {
		t.Error("Expected command from double click")
	}
}

func TestHandleMouse_ShiftClick(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/")
	m.Config.UI.EnableMouse = true
	m.Display.Height = 20
	m.Config.UI.ShowHeader = true
	m.Navigation.FilteredItems = []core.Item{
		{Name: "file1", Path: "/file1"},
	}
	m.Navigation.Select("/file1")

	if m.Navigation.SelectedCount() != 1 {
		t.Fatalf("Expected 1 item selected, got %d", m.Navigation.SelectedCount())
	}

	HandleMouse(m, tea.MouseMsg{
		X:      10,
		Y:      4,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		Shift:  true,
	})

	if m.Navigation.SelectedCount() != 0 {
		t.Error("Expected Shift+Click to deselect")
	}
}
