package update

import (
	"context"
	"fmt"

	"fm/internal/constants"
	"fm/internal/files/ops"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tuierrors "fm/internal/tui/errors"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleOperationsMsg delegates operation-related messages to specialized handlers
func HandleOperationsMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.ProgressMsg:
		return HandleProgress(m, msg)

	case commands.OperationFinishedMsg:
		HandleOperationFinished(m, msg)
		return nil

	case commands.ConflictMsg:
		HandleConflict(m, msg)
		return nil
	}
	return nil
}

// HandleProgress handles operation progress messages
func HandleProgress(m *state.Model, msg commands.ProgressMsg) tea.Cmd {
	m.Operations.Progress.Show(msg.Label)
	m.Operations.Progress.Update(msg.Percent)
	return commands.ListenToProgress(msg.Channel)
}

// HandleOperationFinished handles operation completion messages
func HandleOperationFinished(m *state.Model, msg commands.OperationFinishedMsg) {
	m.Operations.Progress.Hide()
	for _, p := range msg.Paths {
		delete(m.Operations.ProcessingItems, p)
		if m.Operations.SelectedPaths[p] {
			delete(m.Operations.SelectedPaths, p)
			m.Navigation.SelectedCount--
		}
	}
	m.UI.SelectMode = len(m.Operations.SelectedPaths) > 0
}

// HandleRenaming handles renaming events
func HandleRenaming(msg tea.Msg, m *state.Model) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			actions.ClosePrompt(m)
			return nil
		case "enter":
			newName := m.Inputs.ActiveInput.Value()
			if newName != "" {
				// Validate filename
				if err := ops.ValidateFileName(newName); err != nil {
					// User-facing validation error
					userErr := tuierrors.UserError("rename", err.Error()).
						WithCode("INVALID_FILENAME").
						WithContext("filename", newName)
					cmd := actions.LogError(m, userErr, "Invalid filename")
					actions.ClosePrompt(m)
					return cmd
				}

				selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
				oldPath := selected.Path

				if m.Operations.ProcessingItems[oldPath] {
					actions.ClosePrompt(m)
					return commands.SetMsg(m, "Error: Item is currently being processed")
				}

				newPath := m.FS.Join(m.Navigation.Path, newName)

				c, cancel := context.WithTimeout(context.Background(), constants.DirectoryLoadTimeout)
				defer cancel()

				if err := ops.Rename(c, m.FS, oldPath, newPath); err != nil {
					// System error with context
					sysErr := tuierrors.SystemError("rename file", err).
						WithContext("from", oldPath).
						WithContext("to", newPath)
					cmd = actions.LogError(m, sysErr, "Rename")
				} else {
					cmd = tea.Batch(actions.Reload(m), actions.LogInfo(m, fmt.Sprintf("Renamed %s to %s", selected.Name, newName)))
				}
			}
			actions.ClosePrompt(m)
			return cmd
		}
	}
	m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
	return cmd
}

// HandleConfirming handles confirmation events
func HandleConfirming(msg tea.Msg, m *state.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			var cmds []tea.Cmd
			switch m.Operations.ActionType {
			case constants.ActionDelete:
				cmds = append(cmds, PerformDelete(m)...)
			case constants.ActionPaste:
				cmds = append(cmds, PerformPaste(m)...)
			case constants.ActionConflict:
				return ResolveConflict("overwrite", m)
			}
			m.UI.Confirming = false
			return tea.Batch(cmds...)
		case "n", "N", "esc":
			if m.Operations.ActionType == constants.ActionConflict {
				return ResolveConflict("skip", m)
			}
			m.UI.Confirming = false
			return nil
		case "r", "R":
			if m.Operations.ActionType == constants.ActionConflict {
				return ResolveConflict("rename", m)
			}
		}
	}
	return nil
}

// ResolveConflict handles file conflict resolution
func ResolveConflict(choice string, m *state.Model) tea.Cmd {
	var cmds []tea.Cmd

	switch choice {
	case "overwrite":
		cmds = append(cmds, commands.OverwriteItem(m.FS, m.Operations.Conflict.Source, m.Operations.Conflict.Destination, m.Operations.Clipboard.IsCut))
	case "skip":
		// Remove from processingItems since we're skipping it
		delete(m.Operations.ProcessingItems, m.Operations.Conflict.Source)
	case "rename":
		// Auto-rename
		c, cancel := context.WithTimeout(context.Background(), constants.SSHConnectionTimeout)
		defer cancel()

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
		cmds = append(cmds, commands.OverwriteItem(m.FS, m.Operations.Conflict.Source, newName, m.Operations.Clipboard.IsCut))
	}

	m.UI.Confirming = false
	m.Operations.ActionType = ""

	// Continue with pending items if any
	pending := m.Operations.Conflict.PendingItems
	if len(pending) > 0 {
		m.UI.Loading = true
		if m.Operations.Clipboard.IsCut {
			cmds = append(cmds, commands.MoveItems(m.FS, pending, m.Navigation.Path))
		} else {
			cmds = append(cmds, commands.PasteItems(m.FS, pending, m.Navigation.Path))
		}
	} else {
		if choice == "skip" {
			cmds = append(cmds, actions.Reload(m))
		}
	}

	return tea.Batch(cmds...)
}

func checkAndMarkProcessing(m *state.Model, paths []string) bool {
	processing := m.Operations.ProcessingItems
	for _, p := range paths {
		if processing[p] {
			return false
		}
	}
	for _, p := range paths {
		processing[p] = true
	}
	return true
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

	if !checkAndMarkProcessing(m, targets) {
		return []tea.Cmd{commands.SetMsg(m, "Error: Some items are already being processed")}
	}

	m.UI.Loading = true
	return []tea.Cmd{
		commands.SetMsg(m, fmt.Sprintf("Deleting %d items...", len(targets))),
		commands.DeleteItems(m.FS, targets, m.Config.UseTrash),
		actions.Reload(m),
	}
}

// PerformPaste initiates paste/move
func PerformPaste(m *state.Model) []tea.Cmd {
	clipboard := m.Operations.Clipboard.Paths
	if len(clipboard) == 0 {
		return nil
	}

	if !checkAndMarkProcessing(m, clipboard) {
		return []tea.Cmd{commands.SetMsg(m, "Error: Some items are already being processed")}
	}

	m.UI.Loading = true
	if m.Operations.Clipboard.IsCut {
		cmds := []tea.Cmd{
			commands.SetMsg(m, fmt.Sprintf("Moving %d items...", len(clipboard))),
			commands.MoveItems(m.FS, clipboard, m.Navigation.Path),
			actions.Reload(m),
		}
		m.Operations.Clipboard.Paths = []string{}
		m.Operations.Clipboard.IsCut = false
		return cmds
	}
	return []tea.Cmd{
		commands.SetMsg(m, fmt.Sprintf("Pasting %d items...", len(clipboard))),
		commands.PasteItems(m.FS, clipboard, m.Navigation.Path),
		actions.Reload(m),
	}
}
