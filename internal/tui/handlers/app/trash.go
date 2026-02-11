package app

import (
	"github.com/zulfikawr/fm/internal/files/trash"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleTrash handles trash-related messages
func HandleTrash(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.ActiveView == tui_context.ViewTrash {
			return handleTrashKeys(m, msg)
		}
	case messages.TrashLoadedMsg:
		// Store trash items in model
		m.Trash.Items = msg.Items
		m.Trash.Cursor = 0
		m.Trash.Offset = 0
		return nil
	case messages.TrashRestoreMsg:
		return func() tea.Msg {
			return performTrashRestore(m, msg.TrashedName)
		}
	case messages.TrashDeleteMsg:
		return func() tea.Msg {
			return performTrashDelete(m, msg.TrashedName)
		}
	case messages.TrashEmptyMsg:
		return func() tea.Msg {
			return performTrashEmpty(m)
		}
	case messages.TrashOperationFinishedMsg:
		if msg.Success {
			// Reload trash items
			return func() tea.Msg {
				manager, err := trash.NewManager(m.FS)
				if err != nil {
					return messages.ErrorMsg{Err: err}
				}
				items, err := manager.List()
				if err != nil {
					return messages.ErrorMsg{Err: err}
				}
				result := make([]interface{}, len(items))
				for i, item := range items {
					result[i] = item
				}
				return messages.TrashLoadedMsg{Items: result}
			}
		}
		return nil
	}
	return nil
}

func handleTrashKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "t":
		m.UI.ToggleTrash()
		return nil

	case "up", "k":
		if m.Trash.Cursor > 0 {
			m.Trash.Cursor--
		}
		m.Trash.Offset = ScrollTrash(m.Trash.Cursor, m.Trash.Offset, m.Display.ViewportHeight)

	case "down", "j":
		if m.Trash.Cursor < len(m.Trash.Items)-1 {
			m.Trash.Cursor++
		}
		m.Trash.Offset = ScrollTrash(m.Trash.Cursor, m.Trash.Offset, m.Display.ViewportHeight)

	case "r":
		// Restore selected item
		if len(m.Trash.Items) > 0 && m.Trash.Cursor < len(m.Trash.Items) {
			if item, ok := m.Trash.Items[m.Trash.Cursor].(trash.TrashItem); ok {
				return func() tea.Msg {
					return messages.TrashRestoreMsg{TrashedName: item.TrashedName}
				}
			}
		}

	case "d":
		// Delete selected item permanently
		if len(m.Trash.Items) > 0 && m.Trash.Cursor < len(m.Trash.Items) {
			if item, ok := m.Trash.Items[m.Trash.Cursor].(trash.TrashItem); ok {
				return func() tea.Msg {
					return messages.TrashDeleteMsg{TrashedName: item.TrashedName}
				}
			}
		}

	case "e":
		// Empty trash
		if len(m.Trash.Items) > 0 {
			return func() tea.Msg {
				return messages.TrashEmptyMsg{}
			}
		}
	}

	return nil
}

func performTrashRestore(m *tui_context.Model, trashedName string) tea.Msg {
	manager, err := trash.NewManager(m.FS)
	if err != nil {
		return messages.StatusMsg{Message: "Failed to restore: " + err.Error(), IsError: true}
	}

	if err := manager.Restore(m.Context, trashedName); err != nil {
		// Check if it's a conflict error
		if conflictErr, ok := err.(*trash.RestoreConflictError); ok {
			// TODO: Show conflict dialog - for now just use "Keep Both" strategy
			// Generate a new name
			info, err := manager.GetInfo(trashedName)
			if err != nil {
				return messages.StatusMsg{Message: "Restore conflict: " + conflictErr.Error(), IsError: true}
			}
			if info != nil {
				base := m.FS.Base(info.OriginalPath)
				ext := ""
				name := base
				if idx := len(base) - 1; idx >= 0 {
					for i := idx; i >= 0; i-- {
						if base[i] == '.' {
							ext = base[i:]
							name = base[:i]
							break
						}
					}
				}
				newName := name + " (restored)" + ext
				if err := manager.RestoreWithRename(m.Context, trashedName, newName); err != nil {
					return messages.StatusMsg{Message: "Failed to restore: " + err.Error(), IsError: true}
				}
				return messages.TrashOperationFinishedMsg{Success: true, Message: "Item restored as '" + newName + "'"}
			}
			return messages.StatusMsg{Message: "Restore conflict: " + conflictErr.Error(), IsError: true}
		}
		return messages.StatusMsg{Message: "Failed to restore: " + err.Error(), IsError: true}
	}

	return messages.TrashOperationFinishedMsg{Success: true, Message: "Item restored"}
}

func performTrashDelete(m *tui_context.Model, trashedName string) tea.Msg {
	manager, err := trash.NewManager(m.FS)
	if err != nil {
		return messages.StatusMsg{Message: "Failed to delete: " + err.Error(), IsError: true}
	}

	if err := manager.Delete(m.Context, trashedName); err != nil {
		return messages.StatusMsg{Message: "Failed to delete: " + err.Error(), IsError: true}
	}

	return messages.TrashOperationFinishedMsg{Success: true, Message: "Item deleted permanently"}
}

func performTrashEmpty(m *tui_context.Model) tea.Msg {
	manager, err := trash.NewManager(m.FS)
	if err != nil {
		return messages.StatusMsg{Message: "Failed to empty trash: " + err.Error(), IsError: true}
	}

	if err := manager.Empty(m.Context); err != nil {
		return messages.StatusMsg{Message: "Failed to empty trash: " + err.Error(), IsError: true}
	}

	return messages.TrashOperationFinishedMsg{Success: true, Message: "Trash emptied"}
}

// ScrollTrash recalculates the trash view offset
func ScrollTrash(cursor, offset, viewportHeight int) int {
	if viewportHeight <= 0 {
		return 0
	}
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+viewportHeight {
		return cursor - viewportHeight + 1
	}
	return offset
}
