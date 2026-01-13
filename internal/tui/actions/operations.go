package actions

import (
	"context"
	"fmt"

	"fm/internal/constants"
	"fm/internal/files/ops"
	"fm/internal/tui/commands"
	tuierrors "fm/internal/tui/errors"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// UpdateProgress updates the progress state
func UpdateProgress(m *state.Model, label string, percent float64) {
	m.Operations.Progress.Show(label)
	m.Operations.Progress.Update(percent)
}

// FinalizeOperation cleans up state after an operation finishes
func FinalizeOperation(m *state.Model, paths []string) {
	m.Operations.Progress.Hide()
	for _, p := range paths {
		delete(m.Operations.ProcessingItems, p)
		m.Navigation.Deselect(p)
	}
	m.UI.SelectMode = m.Navigation.SelectedCount > 0
}

// PerformRename executes a rename operation
func PerformRename(m *state.Model, newName string) tea.Cmd {
	if newName == "" {
		return nil
	}

	// Validate filename
	if err := ops.ValidateFileName(newName); err != nil {
		userErr := tuierrors.UserError("rename", err.Error()).
			WithCode("INVALID_FILENAME").
			WithContext("filename", newName)
		return LogError(m, userErr, "Invalid filename")
	}

	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}

	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
	oldPath := selected.Path

	if m.Operations.ProcessingItems[oldPath] {
		return commands.SetMsg(m, "Error: Item is currently being processed")
	}

	newPath := m.FS.Join(m.Navigation.Path, newName)

	c, cancel := context.WithTimeout(context.Background(), constants.DirectoryLoadTimeout)
	defer cancel()

	if err := ops.Rename(c, m.FS, oldPath, newPath); err != nil {
		sysErr := tuierrors.SystemError("rename file", err).
			WithContext("from", oldPath).
			WithContext("to", newPath)
		return LogError(m, sysErr, "Rename")
	}

	return tea.Batch(Reload(m), LogInfo(m, fmt.Sprintf("Renamed %s to %s", selected.Name, newName)))
}

// ResolveConflict handles file conflict resolution
func ResolveConflict(choice string, m *state.Model) tea.Cmd {
	var cmds []tea.Cmd

	ctx, cancel := context.WithCancel(context.Background())
	m.Operations.CancelFunc = cancel

	srcFS := m.Operations.Clipboard.SourceFS
	dstFS := m.FS

	switch choice {
	case "overwrite":
		cmds = append(cmds, commands.OverwriteItem(ctx, srcFS, dstFS, m.Operations.Conflict.Source, m.Operations.Conflict.Destination, m.Operations.Clipboard.IsCut))
	case "skip":
		// Remove from processingItems since we're skipping it
		delete(m.Operations.ProcessingItems, m.Operations.Conflict.Source)
	case "rename":
		// Auto-rename
		c, cancelRename := context.WithTimeout(context.Background(), constants.SSHConnectionTimeout)
		defer cancelRename()

		ext := ""
		base := m.Operations.Conflict.Destination
		dst := m.Operations.Conflict.Destination
		for i := len(dst) - 1; i >= 0 && dst[i] != '/'; i-- {
			if dst[i] == '.' {
				ext = dst[i:]
				base = dst[:i]
				break
			}
		}

		newName := ""
		for i := 1; ; i++ {
			newName = fmt.Sprintf("%s (%d)%s", base, i, ext)
			if _, err := m.FS.Stat(c, newName); err != nil {
				break
			}
		}
		cmds = append(cmds, commands.OverwriteItem(ctx, srcFS, dstFS, m.Operations.Conflict.Source, newName, m.Operations.Clipboard.IsCut))
	}

	m.UI.Confirming = false
	m.Operations.ActionType = ""

	// Continue with pending items if any
	pending := m.Operations.Conflict.PendingItems
	if len(pending) > 0 {
		m.UI.Loading = true
		if m.Operations.Clipboard.IsCut {
			cmds = append(cmds, commands.MoveItems(ctx, srcFS, dstFS, pending, m.Navigation.Path))
		} else {
			cmds = append(cmds, commands.PasteItems(ctx, srcFS, dstFS, pending, m.Navigation.Path))
		}
	} else {
		if choice == "skip" {
			cmds = append(cmds, Reload(m))
		}
	}

	return tea.Batch(cmds...)
}

// PerformDelete initiates deletion
func PerformDelete(m *state.Model) []tea.Cmd {
	var targets []string
	for _, item := range m.Navigation.Items {
		if item.Selected {
			targets = append(targets, item.Path)
		}
	}
	if len(targets) == 0 && len(m.Navigation.FilteredItems) > 0 {
		cursor := m.Navigation.Cursor
		if cursor < len(m.Navigation.FilteredItems) {
			sel := m.Navigation.FilteredItems[cursor]
			if !sel.IsUp {
				targets = append(targets, sel.Path)
			}
		}
	}

	if len(targets) == 0 {
		return nil
	}

	if !ops.CheckAndMarkProcessing(m.Operations.ProcessingItems, targets) {
		return []tea.Cmd{commands.SetMsg(m, "Error: Some items are already being processed")}
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.Operations.CancelFunc = cancel

	m.UI.Loading = true
	return []tea.Cmd{
		commands.SetMsg(m, fmt.Sprintf("Deleting %d items...", len(targets))),
		commands.DeleteItems(ctx, m.FS, targets, m.Config.UseTrash),
	}
}

// PerformPaste initiates paste/move
func PerformPaste(m *state.Model) []tea.Cmd {
	clipboard := m.Operations.Clipboard.Paths
	if len(clipboard) == 0 {
		return nil
	}

	if !ops.CheckAndMarkProcessing(m.Operations.ProcessingItems, clipboard) {
		return []tea.Cmd{commands.SetMsg(m, "Error: Some items are already being processed")}
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.Operations.CancelFunc = cancel

	srcFS := m.Operations.Clipboard.SourceFS
	dstFS := m.FS

	m.UI.Loading = true
	if m.Operations.Clipboard.IsCut {
		cmds := []tea.Cmd{
			commands.SetMsg(m, fmt.Sprintf("Moving %d items...", len(clipboard))),
			commands.MoveItems(ctx, srcFS, dstFS, clipboard, m.Navigation.Path),
		}
		m.Operations.Clipboard.Paths = []string{}
		m.Operations.Clipboard.IsCut = false
		return cmds
	}
	return []tea.Cmd{
		commands.SetMsg(m, fmt.Sprintf("Pasting %d items...", len(clipboard))),
		commands.PasteItems(ctx, srcFS, dstFS, clipboard, m.Navigation.Path),
	}
}
