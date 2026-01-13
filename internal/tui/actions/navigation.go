package actions

import (
	"context"
	"fm/internal/files/local"
	"fm/internal/files/ops"
	"fm/internal/tui/commands"
	"fm/internal/tui/filter"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// ClearSelection clears all selected items in the current model
func ClearSelection(m *state.Model) bool {
	hasSelection := m.Navigation.SelectedCount > 0
	m.ClearSelection()

	if hasSelection {
		filter.Apply(m)
		return true
	}
	m.Message.Text = ""
	return true
}

// NavigateToPath handles navigation to a specific directory path
func NavigateToPath(m *state.Model, path string) tea.Cmd {
	// Clean and validate path
	info, err := m.FS.Stat(context.TODO(), path)
	if err != nil {
		return commands.SetMsg(m, "Error: "+err.Error())
	}

	if !info.IsDir() {
		return commands.SetMsg(m, "Error: Not a directory")
	}

	// Save current state
	m.Cache.CursorMemory.Put(m.Navigation.Path, m.Navigation.Cursor)
	m.Cache.OffsetMemory.Put(m.Navigation.Path, m.Navigation.Offset)
	SaveTabState(m)

	// Update path
	m.Navigation.Path = path
	m.Navigation.PathGen++

	// Reset view state for new directory
	m.Navigation.Cursor = 0
	m.Navigation.Offset = 0
	m.Navigation.Items = nil
	m.Navigation.FilteredItems = nil

	return Reload(m)
}

// SwitchToLocal switches the current filesystem back to local
func SwitchToLocal(m *state.Model, path string) tea.Cmd {
	if m.FS.IsLocal() {
		return NavigateToPath(m, path)
	}

	// Close remote FS
	m.FS.Close()

	// Initialize local FS
	m.FS = local.NewLocalFS()

	targetPath := path
	if targetPath == "" {
		home, err := m.FS.GetHomeDir()
		if err == nil {
			targetPath = home
		} else {
			targetPath = "/"
		}
	}

	// Reset view state
	m.Navigation.Path = targetPath
	m.Navigation.PathGen++
	m.Navigation.Cursor = 0
	m.Navigation.Offset = 0
	m.Navigation.Items = nil
	m.Navigation.FilteredItems = nil

	return Reload(m)
}

// NavigateToParent handles navigation to the parent directory
func NavigateToParent(m *state.Model) tea.Cmd {
	parent := m.FS.Dir(m.Navigation.Path)
	return NavigateToPath(m, parent)
}

// NavigateToSelected handles navigation into the currently selected directory
func NavigateToSelected(m *state.Model) tea.Cmd {
	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}
	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]

	if selected.IsUp {
		return NavigateToParent(m)
	}

	if selected.IsDir {
		// Validate path to prevent traversal attacks
		if err := ops.ValidatePath(m.Navigation.Path, selected.Name); err != nil {
			return commands.SetMsg(m, "Security: "+err.Error())
		}

		if !selected.CanRead {
			return commands.SetMsg(m, "Access Denied: You do not have permission to read this directory")
		}

		return NavigateToPath(m, m.FS.Join(m.Navigation.Path, selected.Name))
	}

	return nil
}
