package nav

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	"github.com/zulfikawr/fm/internal/files/sorting"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func HandleNavKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	// Don't handle nav keys if an input is active or a modal view is open
	if m.UI.InputActive || m.UI.ActiveView != tui_context.ViewMain {
		return nil
	}

	key := msg.String()

	// Check for matching action using custom keybindings
	action := GetActionForKeyFromModel(m, key)

	// Tab management shortcuts
	if strings.HasPrefix(key, "alt+") {
		if len(key) == 5 && key[4] >= '1' && key[4] <= '9' {
			tabNum := int(key[4] - '0')
			return SwitchTab(m, tabNum)
		}
		// Check if alt+X combinations match any custom keybinding
		switch action {
		case "new_tab":
			return CreateTab(m)
		case "close_tab":
			return CloseTab(m)
		case "create":
			if !m.UI.InputActive {
				return func() tea.Msg { return messages.StartCreateMsg{} }
			}
		}
	}

	// Process actions based on custom keybindings
	switch action {
	case "move_up":
		MoveCursor(m, -1)
		m.Navigation.LastShiftIdx = -1
		m.Display.Mouse.InitialSelection = nil
	case "move_down":
		MoveCursor(m, 1)
		m.Navigation.LastShiftIdx = -1
		m.Display.Mouse.InitialSelection = nil
	case "page_up":
		// Implement page up navigation
		pageSize := m.Display.ViewportHeight - 3 // Account for header/footer
		if pageSize <= 0 {
			pageSize = 5 // Default to 5 if viewport is small
		}
		MoveCursor(m, -pageSize)
		m.Navigation.LastShiftIdx = -1
		m.Display.Mouse.InitialSelection = nil
	case "page_down":
		// Implement page down navigation
		pageSize := m.Display.ViewportHeight - 3 // Account for header/footer
		if pageSize <= 0 {
			pageSize = 5 // Default to 5 if viewport is small
		}
		MoveCursor(m, pageSize)
		m.Navigation.LastShiftIdx = -1
		m.Display.Mouse.InitialSelection = nil
	case "toggle_selection":
		ToggleSelection(m)
		return nil
	case "select_all":
		SelectAll(m)
		return nil
	case "clear_selection":
		m.ClearSelection()
		return nil
	case "open":
		return NavigateToSelected(m)
	case "go_parent":
		return NavigateToParent(m)
	case "filter":
		m.StartInput(tui_context.InputSearch)
		utils.UpdateSearchSuggestion(m)
		return m.Inputs.ActiveInput.FocusCmd()
	case "cycle_sort":
		m.Display.SortMode = m.Display.SortMode.Next()
		sorting.SortItems(m.Navigation.Items, m.Display.SortMode, true)
		ApplyFilter(m)
		return nil
	case "go_to_path":
		m.Operations.ActionType = constants.ActionGoto
		m.UI.StartConfirming()
		return nil
	case "fuzzy_search":
		m.StartInput(tui_context.InputFuzzySearch)
		return m.Inputs.ActiveInput.FocusCmd()
	case "history_back":
		return NavigateBack(m)
	case "history_forward":
		return NavigateForward(m)
	}

	// Handle shift+key combinations for range selection
	if strings.HasPrefix(key, "shift+") {
		switch {
		case strings.HasSuffix(key, "+up") || strings.HasSuffix(key, "+k"):
			if containsKeyForAction(m, key, "move_up") {
				return HandleShiftSelect(m, -1)
			}
		case strings.HasSuffix(key, "+down") || strings.HasSuffix(key, "+j"):
			if containsKeyForAction(m, key, "move_down") {
				return HandleShiftSelect(m, 1)
			}
		}
	}

	// Check if the key matches any of the configured keys for these actions
	// This handles the case where custom bindings might use different keys
	if containsKeyForAction(m, key, "move_up") || containsKeyForAction(m, key, "move_down") ||
		containsKeyForAction(m, key, "open") || containsKeyForAction(m, key, "go_parent") {
		// Already handled above, but this catches any edge cases
		return nil
	}

	return nil
}

// GetActionForKeyFromModel retrieves the action for a given key from the model's config
func GetActionForKeyFromModel(m *tui_context.Model, key string) string {
	for _, kb := range m.Config.Keybindings {
		for _, bindKey := range kb.Keys {
			if bindKey == key {
				return kb.Action
			}
		}
	}
	return ""
}

// containsKeyForAction checks if the given key is mapped to the specified action
func containsKeyForAction(m *tui_context.Model, key, action string) bool {
	for _, kb := range m.Config.Keybindings {
		if kb.Action == action {
			for _, bindKey := range kb.Keys {
				if bindKey == key {
					return true
				}
			}
		}
	}
	return false
}

func MoveCursor(m *tui_context.Model, delta int) {
	items := m.Navigation.FilteredItems
	if len(items) == 0 {
		return
	}

	newCursor := m.Navigation.Cursor + delta
	if newCursor < 0 {
		if m.Config.Ops.WrapNavigation {
			newCursor = len(items) - 1
		} else {
			newCursor = 0
		}
	} else if newCursor >= len(items) {
		if m.Config.Ops.WrapNavigation {
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
	if item.State.IsUp {
		return
	}

	item.State.Selected = !item.State.Selected
	m.Navigation.ToggleSelection(item.Path)

	m.Navigation.SelectMode = m.Navigation.SelectedCount() > 0
}

func SelectAll(m *tui_context.Model) {
	if len(m.Navigation.FilteredItems) == 0 {
		return
	}
	m.Navigation.SelectAll()
	m.Navigation.SelectMode = m.Navigation.SelectedCount() > 0
}

func NavigateToSelected(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}
	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]

	if selected.State.IsUp {
		return NavigateToParent(m)
	}

	if selected.IsDir {
		if err := ops.ValidatePath(m.FS, m.Navigation.Path, selected.Name); err != nil {
			return func() tea.Msg { return messages.ErrorMsg{Err: err} }
		}
		if !selected.Metadata.CanRead {
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
		m.Display.Mouse.InitialSelection = make(map[string]bool)
		for k, v := range m.Navigation.SelectedPaths {
			m.Display.Mouse.InitialSelection[k] = v
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
	if m.Display.Mouse.InitialSelection != nil {
		for path := range m.Display.Mouse.InitialSelection {
			m.Navigation.Select(path)
		}
	}

	for i := start; i <= end; i++ {
		item := items[i]
		if !item.State.IsUp {
			m.Navigation.Select(item.Path)
		}
	}
	m.Navigation.SelectMode = m.Navigation.SelectedCount() > 0
	return nil
}
