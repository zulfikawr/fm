package nav

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	"github.com/zulfikawr/fm/internal/files/sorting"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func HandleNavKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	// Don't handle nav keys if an input is active or a modal view is open
	if m.UI.InputActive || m.UI.SettingsOpen || m.UI.LogOpen || m.UI.ClipboardOpen {
		return nil
	}

	key := msg.String()

	// Tab management shortcuts
	if strings.HasPrefix(key, "alt+") {
		if len(key) == 5 && key[4] >= '1' && key[4] <= '9' {
			tabNum := int(key[4] - '0')
			return SwitchTab(m, tabNum)
		}
		switch key {
		case "alt+t":
			return CreateTab(m)
		case "alt+w":
			return CloseTab(m)
		case "alt+n":
			if !m.UI.InputActive {
				return func() tea.Msg { return messages.StartCreateMsg{} }
			}
		case "alt+m":
			// reserved
		}
	}

	switch key {
	case "up", "k":
		MoveCursor(m, -1)
		m.Navigation.LastShiftIdx = -1
		m.Display.InitialSelectedPaths = nil
	case "down", "j":
		MoveCursor(m, 1)
		m.Navigation.LastShiftIdx = -1
		m.Display.InitialSelectedPaths = nil
	case "shift+up", "shift+k":
		return HandleShiftSelect(m, -1)
	case "shift+down", "shift+j":
		return HandleShiftSelect(m, 1)
	case "enter", "right", "l":
		return NavigateToSelected(m)
	case "backspace", "left", "h":
		return NavigateToParent(m)
	case "esc":
		m.ClearSelection()
		return nil
	case " ":
		ToggleSelection(m)
		return nil
	case "alt+a":
		SelectAll(m)
		return nil
	case "/":
		m.StartInput(tui_context.InputSearch)
		return m.Inputs.ActiveInput.FocusCmd()
	case "s":
		m.Display.SortMode = m.Display.SortMode.Next()
		sorting.SortItems(m.Navigation.Items, m.Display.SortMode, true)
		ApplyFilter(m)
		return nil
	case "g":
		m.StartInput(tui_context.InputGoto)
		m.Inputs.ActiveInput.SetValue(m.Navigation.Path)
		return m.Inputs.ActiveInput.FocusCmd()
	case "alt+/":
		m.StartInput(tui_context.InputFuzzySearch)
		return m.Inputs.ActiveInput.FocusCmd()
	case "[":
		return NavigateBack(m)
	case "]":
		return NavigateForward(m)
	}
	return nil
}

func MoveCursor(m *tui_context.Model, delta int) {
	items := m.Navigation.FilteredItems
	if len(items) == 0 {
		return
	}

	newCursor := m.Navigation.Cursor + delta
	if newCursor < 0 {
		if m.Config.WrapNavigation {
			newCursor = len(items) - 1
		} else {
			newCursor = 0
		}
	} else if newCursor >= len(items) {
		if m.Config.WrapNavigation {
			newCursor = 0
		} else {
			newCursor = len(items) - 1
		}
	}

	m.Navigation.Cursor = newCursor
	SyncOffset(m)

	m.Cache.CursorMemory.Put(m.Navigation.Path, m.Navigation.Cursor)
	m.Cache.OffsetMemory.Put(m.Navigation.Path, m.Navigation.Offset)
}

func ToggleSelection(m *tui_context.Model) {
	if len(m.Navigation.FilteredItems) == 0 {
		return
	}
	cursor := m.Navigation.Cursor
	if cursor >= len(m.Navigation.FilteredItems) {
		return
	}
	item := &m.Navigation.FilteredItems[cursor]
	if item.IsUp {
		return
	}

	item.Selected = !item.Selected
	m.Navigation.ToggleSelection(item.Path)

	m.UI.SelectMode = m.Navigation.SelectedCount > 0
}

func SelectAll(m *tui_context.Model) {
	if len(m.Navigation.FilteredItems) == 0 {
		return
	}
	m.Navigation.SelectAll()
	m.UI.SelectMode = m.Navigation.SelectedCount > 0
}

func NavigateToSelected(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}
	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]

	if selected.IsUp {
		return NavigateToParent(m)
	}

	if selected.IsDir {
		if err := ops.ValidatePath(m.FS, m.Navigation.Path, selected.Name); err != nil {
			return func() tea.Msg { return messages.ErrorMsg{Err: err} }
		}
		if !selected.CanRead {
			return func() tea.Msg { return messages.ErrorMsg{Err: fmt.Errorf("access denied")} }
		}
		return NavigateToPath(m, m.FS.Join(m.Navigation.Path, selected.Name))
	}

	if selected.IsArchive() {
		return EnterArchive(m, selected)
	}

	return func() tea.Msg { return messages.OpenFileMsg{Item: selected} }
}

func NavigateToParent(m *tui_context.Model) tea.Cmd {
	if m.Navigation.ParentFS != nil && core.IsRoot(m.FS, m.Navigation.Path) {
		return ExitArchive(m)
	}

	if core.IsRoot(m.FS, m.Navigation.Path) {
		return nil
	}

	return NavigateToPath(m, core.GetParent(m.FS, m.Navigation.Path))
}

func HandleShiftSelect(m *tui_context.Model, delta int) tea.Cmd {
	items := m.Navigation.FilteredItems

	if len(items) == 0 {
		return nil
	}

	if m.Navigation.LastShiftIdx == -1 {
		m.Navigation.LastShiftIdx = m.Navigation.Cursor

		// Store initial state
		m.Display.InitialSelectedPaths = make(map[string]bool)
		for k, v := range m.Navigation.SelectedPaths {
			m.Display.InitialSelectedPaths[k] = v
		}
	}

	MoveCursor(m, delta)
	start := m.Navigation.LastShiftIdx
	end := m.Navigation.Cursor
	if start > end {
		start, end = end, start
	}

	// Reset to initial state before applying current range
	m.Navigation.ClearSelection()
	if m.Display.InitialSelectedPaths != nil {
		for path := range m.Display.InitialSelectedPaths {
			m.Navigation.Select(path)
		}
	}

	for i := start; i <= end; i++ {
		item := items[i]
		if !item.IsUp {
			m.Navigation.Select(item.Path)
		}
	}
	m.UI.SelectMode = m.Navigation.SelectedCount > 0
	return nil
}
