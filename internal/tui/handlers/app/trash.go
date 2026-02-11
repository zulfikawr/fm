package app

import (
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/trash"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleTrash handles trash-related messages
func HandleTrash(m *tuictx.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.ActiveView == tuictx.ViewTrash {
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
	case messages.TrashRestoreConflictMsg:
		// Set up conflict state for user to resolve
		destPath := msg.OriginalPath
		m.Trash.RestoreConflict = &tuictx.TrashRestoreConflict{
			TrashedName:    msg.TrashedName,
			OriginalPath:   msg.OriginalPath,
			DestPath:       destPath,
			ConflictReason: msg.ConflictReason,
		}
		m.Operations.ActionType = constants.ActionTrashRestore
		m.Operations.ConflictPolicy = conflict.Ask
		return nil
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

func handleTrashKeys(m *tuictx.Model, msg tea.KeyMsg) tea.Cmd {
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

func performTrashRestore(m *tuictx.Model, trashedName string) tea.Msg {
	manager, err := trash.NewManager(m.FS)
	if err != nil {
		return messages.StatusMsg{Message: "Failed to restore: " + err.Error(), IsError: true}
	}

	if err := manager.Restore(m.Context, trashedName); err != nil {
		// Check if it's a conflict error
		if conflictErr, ok := err.(*trash.RestoreConflictError); ok {
			// Get trash item info for proper conflict dialog
			info, infoErr := manager.GetInfo(trashedName)
			if infoErr != nil {
				return messages.StatusMsg{Message: "Restore conflict: " + conflictErr.Error(), IsError: true}
			}
			if info == nil {
				return messages.StatusMsg{Message: "Restore conflict: " + conflictErr.Error(), IsError: true}
			}

			// Emit conflict message so UI can show dialog
			return messages.TrashRestoreConflictMsg{
				TrashedName:    trashedName,
				OriginalPath:   info.OriginalPath,
				ConflictReason: conflictErr.Error(),
			}
		}
		return messages.StatusMsg{Message: "Failed to restore: " + err.Error(), IsError: true}
	}

	return messages.TrashOperationFinishedMsg{Success: true, Message: "Item restored"}
}

func performTrashDelete(m *tuictx.Model, trashedName string) tea.Msg {
	manager, err := trash.NewManager(m.FS)
	if err != nil {
		return messages.StatusMsg{Message: "Failed to delete: " + err.Error(), IsError: true}
	}

	if err := manager.Delete(m.Context, trashedName); err != nil {
		return messages.StatusMsg{Message: "Failed to delete: " + err.Error(), IsError: true}
	}

	return messages.TrashOperationFinishedMsg{Success: true, Message: "Item deleted permanently"}
}

func performTrashEmpty(m *tuictx.Model) tea.Msg {
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

// ResolveTrashRestoreConflict handles user input for trash restore conflict resolution
func ResolveTrashRestoreConflict(m *tuictx.Model, choice string) tea.Msg {
	if m.Trash.RestoreConflict == nil {
		return messages.StatusMsg{Message: "No active trash restore conflict", IsError: true}
	}

	manager, err := trash.NewManager(m.FS)
	if err != nil {
		m.Trash.RestoreConflict = nil
		m.Operations.ActionType = constants.ActionNone
		return messages.StatusMsg{Message: "Failed to resolve conflict: " + err.Error(), IsError: true}
	}

	trashedName := m.Trash.RestoreConflict.TrashedName
	originalPath := m.Trash.RestoreConflict.OriginalPath

	switch choice {
	case "y", "Y":
		// Overwrite: restore to original path, replacing existing file
		if err := manager.Restore(m.Context, trashedName); err != nil {
			m.Trash.RestoreConflict = nil
			m.Operations.ActionType = constants.ActionNone
			return messages.StatusMsg{Message: "Failed to restore: " + err.Error(), IsError: true}
		}
		m.Trash.RestoreConflict = nil
		m.Operations.ActionType = constants.ActionNone
		return messages.TrashOperationFinishedMsg{Success: true, Message: "Item restored (overwritten)"}

	case "n", "N":
		// Skip: don't restore this item
		m.Trash.RestoreConflict = nil
		m.Operations.ActionType = constants.ActionNone
		return messages.TrashOperationFinishedMsg{Success: false, Message: "Restore skipped"}

	case "r", "R":
		// Rename: restore with a new name
		base := m.FS.Base(originalPath)
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
			m.Trash.RestoreConflict = nil
			m.Operations.ActionType = constants.ActionNone
			return messages.StatusMsg{Message: "Failed to restore: " + err.Error(), IsError: true}
		}
		m.Trash.RestoreConflict = nil
		m.Operations.ActionType = constants.ActionNone
		return messages.TrashOperationFinishedMsg{Success: true, Message: "Item restored as '" + newName + "'"}

	case "c", "C":
		// Cancel: abort restore operation
		m.Trash.RestoreConflict = nil
		m.Operations.ActionType = constants.ActionNone
		return messages.StatusMsg{Message: "Restore cancelled"}

	default:
		return messages.StatusMsg{Message: "Invalid choice. Press [y]es, [n]o, [r]ename, or [c]ancel", IsError: false}
	}
}
