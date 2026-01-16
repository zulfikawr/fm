package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"fm/internal/constants"
	"fm/internal/files/conflict"
	"fm/internal/files/core"
	"fm/internal/files/ops"
	tui_context "fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleFileOps handles file system and operation messages
func HandleFileOps(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.Confirming {
			return handleConfirmKeys(m, msg)
		}
		return handleFileKeys(m, msg)

	case ProgressMsg:
		return handleProgress(m, msg)

	case OperationFinishedMsg:
		return finalizeOperation(m, msg)

	case ConflictMsg:
		return handleConflict(m, msg)

	case WatchEventMsg:
		if m.Watcher.Watcher != nil {
			return Reload(m)
		}
	}
	return nil
}

func handleFileKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	if m.UI.InputActive || m.UI.SettingsOpen || m.UI.LogOpen || m.UI.ClipboardOpen {
		return nil
	}

	switch msg.String() {
	case "y", "c":
		return copySelected(m)
	case "x":
		return cutSelected(m)
	case "v":
		return performPaste(m)
	case "d":
		return performDelete(m)
	case "r":
		return startRename(m)
	case "z":
		return startZip(m)
	case "u":
		return startUnzip(m)
	}
	return nil
}

func handleConfirmKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	action := m.Operations.ActionType

	if action == constants.ActionConflict {
		switch msg.String() {
		case "y":
			return resolveConflict(m, "overwrite", false)
		case "Y":
			return resolveConflict(m, "overwrite", true)
		case "n":
			return resolveConflict(m, "skip", false)
		case "N":
			return resolveConflict(m, "skip", true)
		case "r":
			return resolveConflict(m, "rename", false)
		case "R":
			return resolveConflict(m, "rename", true)
		case "esc":
			m.UI.StopConfirming()
			m.Operations.ActionType = constants.ActionNone
			return nil
		}
		return nil
	}

	switch msg.String() {
	case "y", "Y":
		m.UI.StopConfirming()

		switch action {
		case constants.ActionDelete:
			return performDelete(m)
		case constants.ActionPaste:
			return performPaste(m)
		case constants.ActionResetSettings:
			return ConfirmSettingsReset(m)
		}
		m.Operations.ActionType = constants.ActionNone
	case "n", "N", "esc":
		m.UI.StopConfirming()
		m.Operations.ActionType = constants.ActionNone
	}
	return nil
}

func resolveConflict(m *tui_context.Model, choice string, applyToAll bool) tea.Cmd {
	var cmds []tea.Cmd

	ctx, cancel := context.WithCancel(m.Context)
	m.Operations.CancelFunc = cancel

	srcFS := m.Operations.Clipboard.SourceFS
	dstFS := m.FS
	logID := m.Operations.Conflict.LogID
	opType := m.Operations.Conflict.OpType
	dst := m.Operations.Conflict.Destination
	pending := m.Operations.Conflict.PendingItems
	isMove := m.Operations.Conflict.IsMove

	m.Operations.Conflict.ApplyToAll = applyToAll

	// Update policy based on choice
	policy := conflict.Ask
	switch choice {
	case "overwrite":
		policy = conflict.Overwrite
	case "skip":
		policy = conflict.Skip
	case "rename":
		policy = conflict.Rename
	}
	m.Operations.ConflictPolicy = policy

	// Resolve the immediate conflict using the centralized resolver
	resolver := conflict.NewResolver()
	// For Resolve, we need the source. In ConflictState, Source is the path.
	src := m.Operations.Conflict.Source
	resolvedPath, _, err := resolver.Resolve(m.Context, m.FS, src, dst, policy)
	if err == nil {
		if resolvedPath == "" && choice == "skip" {
			// Item skipped
			if len(pending) <= 1 && !applyToAll {
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				m.Operations.ConflictPolicy = conflict.Ask
				return Reload(m)
			}
			// For batch skip, we continue with the rest of the pending items
			pending = pending[1:]
		} else if resolvedPath != "" {
			dst = resolvedPath
		}
	}

	m.UI.StopConfirming()
	m.Operations.ActionType = constants.ActionNone

	// Trigger the operation again with the new policy
	switch opType {
	case "zip":
		zipName := m.FS.Base(dst)
		cmds = append(cmds, PerformZip(m, zipName))
	case "unzip":
		destName := m.FS.Base(dst)
		cmds = append(cmds, PerformUnzip(m, destName))
	case "move":
		m.UI.Loading = true
		cmds = append(cmds, moveItems(ctx, srcFS, dstFS, pending, m.Navigation.Path, m.Operations.ConflictPolicy, logID))
	case "copy":
		m.UI.Loading = true
		cmds = append(cmds, pasteItems(ctx, srcFS, dstFS, pending, m.Navigation.Path, m.Operations.ConflictPolicy, logID))
	default:
		// Fallback for older code that might not set opType correctly
		if isMove {
			cmds = append(cmds, moveItems(ctx, srcFS, dstFS, pending, m.Navigation.Path, m.Operations.ConflictPolicy, logID))
		} else {
			cmds = append(cmds, pasteItems(ctx, srcFS, dstFS, pending, m.Navigation.Path, m.Operations.ConflictPolicy, logID))
		}
	}

	// Reset policy to Ask for next time UNLESS ApplyToAll is set
	if !applyToAll {
		m.Operations.ConflictPolicy = conflict.Ask
	}

	return tea.Batch(cmds...)
}

func overwriteItem(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string, isMove bool, logID string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			var err error
			if isMove {
				err = ops.CrossMove(ctx, srcFS, dstFS, src, dst, progChan, conflict.Overwrite)
			} else {
				err = ops.CrossCopy(ctx, srcFS, dstFS, src, dst, progChan, conflict.Overwrite)
			}

			if err != nil {
				return ErrorMsg{Err: err, LogID: logID}
			}
			return OperationFinishedMsg{Paths: []string{src, dst}, LogID: logID}
		},
	)
}

func hasTargets(m *tui_context.Model) bool {
	if m.Navigation.SelectedCount > 0 {
		return true
	}
	if len(m.Navigation.FilteredItems) > 0 {
		cursor := m.Navigation.Cursor
		if cursor < len(m.Navigation.FilteredItems) {
			return !m.Navigation.FilteredItems[cursor].IsUp
		}
	}
	return false
}

func getTargets(m *tui_context.Model) []string {
	var targets []string
	if m.Navigation.SelectedCount > 0 {
		for path, selected := range m.Navigation.SelectedPaths {
			if selected {
				targets = append(targets, path)
			}
		}
	} else if len(m.Navigation.FilteredItems) > 0 {
		cursor := m.Navigation.Cursor
		if cursor < len(m.Navigation.FilteredItems) {
			sel := m.Navigation.FilteredItems[cursor]
			if !sel.IsUp {
				targets = append(targets, sel.Path)
			}
		}
	}
	return targets
}

func copySelected(m *tui_context.Model) tea.Cmd {
	targets := getTargets(m)
	if len(targets) == 0 {
		return nil
	}
	m.Operations.Clipboard.SetCopy(m.FS, targets)
	m.ClearSelection()
	return SetMsg(m, fmt.Sprintf("Copied %d items to clipboard", len(targets)))
}

func cutSelected(m *tui_context.Model) tea.Cmd {
	targets := getTargets(m)
	if len(targets) == 0 {
		return nil
	}
	m.Operations.Clipboard.SetCut(m.FS, targets)
	m.ClearSelection()
	return SetMsg(m, fmt.Sprintf("Cut %d items to clipboard", len(targets)))
}

func formatDisplayPath(m *tui_context.Model, fs core.FileSystem, path string) string {
	if fs.IsLocal() {
		return path
	}
	host := m.Remote.Host
	user := m.Remote.User
	if host != "" {
		if user != "" {
			return fmt.Sprintf("%s@%s:%s", user, host, path)
		}
		return fmt.Sprintf("%s:%s", host, path)
	}
	return path
}

func performPaste(m *tui_context.Model) tea.Cmd {
	paths := m.Operations.Clipboard.Paths
	if len(paths) == 0 {
		return SetMsg(m, "Clipboard is empty")
	}

	if m.Config.ConfirmOperations && m.Operations.ActionType != constants.ActionPaste {
		m.UI.StartConfirming()
		m.Operations.ActionType = constants.ActionPaste
		return nil
	}

	m.Operations.ActionType = constants.ActionNone
	m.UI.Loading = true
	m.Operations.ConflictPolicy = conflict.Ask

	srcFS := m.Operations.Clipboard.SourceFS
	dstFS := m.FS
	destDir := m.Navigation.Path
	isCut := m.Operations.Clipboard.IsCut

	opName := "Paste"
	opVerb := "Pasting"
	if isCut {
		opName = "Move"
		opVerb = "Moving"
	}

	srcPath := srcFS.Dir(paths[0])
	var msg string
	if len(paths) == 1 {
		msg = fmt.Sprintf("%s %s from %s to %s",
			opVerb,
			srcFS.Base(paths[0]),
			formatDisplayPath(m, srcFS, srcPath),
			formatDisplayPath(m, dstFS, destDir))
	} else {
		msg = fmt.Sprintf("%s %d items from %s to %s",
			opVerb,
			len(paths),
			formatDisplayPath(m, srcFS, srcPath),
			formatDisplayPath(m, dstFS, destDir))
	}

	logID := LogPush(m, opName, tui_context.LogInfo, tui_context.StatusRunning, msg, "")

	ctx, cancel := context.WithCancel(m.Context)
	m.Operations.CancelFunc = cancel

	if isCut {
		m.Operations.Clipboard.Clear()
		return moveItems(ctx, srcFS, dstFS, paths, destDir, m.Operations.ConflictPolicy, logID)
	}
	return pasteItems(ctx, srcFS, dstFS, paths, destDir, m.Operations.ConflictPolicy, logID)
}

func performDelete(m *tui_context.Model) tea.Cmd {
	targets := getTargets(m)
	if len(targets) == 0 {
		return nil
	}

	if m.Config.ConfirmOperations && m.Operations.ActionType != constants.ActionDelete {
		m.UI.StartConfirming()
		m.Operations.ActionType = constants.ActionDelete
		return nil
	}

	m.Operations.ActionType = constants.ActionNone
	m.UI.Loading = true

	var msg string
	if len(targets) == 1 {
		msg = fmt.Sprintf("Deleting %s from %s",
			m.FS.Base(targets[0]),
			formatDisplayPath(m, m.FS, m.Navigation.Path))
	} else {
		msg = fmt.Sprintf("Deleting %d items from %s",
			len(targets),
			formatDisplayPath(m, m.FS, m.Navigation.Path))
	}

	logID := LogPush(m, "Delete", tui_context.LogInfo, tui_context.StatusRunning, msg, "")

	ctx, cancel := context.WithCancel(m.Context)
	m.Operations.CancelFunc = cancel

	return deleteItems(ctx, m.FS, targets, m.Config.UseTrash, logID)
}

func startRename(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}
	cursor := m.Navigation.Cursor
	if cursor >= len(m.Navigation.FilteredItems) {
		return nil
	}
	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
	if selected.IsUp {
		return nil
	}

	m.StartInput(tui_context.InputRename)
	m.Inputs.ActiveInput.SetValue(selected.Name)
	return m.Inputs.ActiveInput.FocusCmd()
}

func startZip(m *tui_context.Model) tea.Cmd {
	targets := getTargets(m)
	if len(targets) == 0 {
		return nil
	}

	m.StartInput(tui_context.InputZip)
	m.Inputs.ActiveInput.SetValue("Archive.zip")
	m.Inputs.ActiveInput.SetCursor(len("Archive"))
	return m.Inputs.ActiveInput.FocusCmd()
}

// PerformZip executes a zip operation
func PerformZip(m *tui_context.Model, zipName string) tea.Cmd {
	if zipName == "" {
		return nil
	}

	targets := getTargets(m)
	if len(targets) == 0 {
		return nil
	}

	dst := m.FS.Join(m.Navigation.Path, zipName)

	// Check for conflict
	if m.Operations.ConflictPolicy == conflict.Ask {
		resolver := conflict.NewResolver()
		resolvedPath, _, err := resolver.Resolve(m.Context, m.FS, targets[0], dst, m.Operations.ConflictPolicy)
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				m.UI.Loading = false
				m.Operations.Conflict.Set(cerr.Source, cerr.Destination, targets, false, "zip", "")
				m.Operations.ActionType = constants.ActionConflict
				m.UI.StartConfirming()
				return nil
			}
		}
		if resolvedPath == "" {
			return nil // Skip
		}
		dst = resolvedPath
	}

	m.ClearSelection()
	m.UI.Loading = true

	logID := LogPush(m, "Zip", tui_context.LogInfo, tui_context.StatusRunning,
		fmt.Sprintf("Zipping %d items into %s", len(targets), zipName), "")

	ctx, cancel := context.WithCancel(m.Context)
	m.Operations.CancelFunc = cancel

	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.Zip(ctx, m.FS, targets, dst, progChan, m.Operations.ConflictPolicy)
			if err != nil {
				return ErrorMsg{Err: err, LogID: logID}
			}
			return OperationFinishedMsg{Paths: []string{}, LogID: logID}
		},
	)
}

func handleProgress(m *tui_context.Model, msg ProgressMsg) tea.Cmd {
	// Throttle updates to ~30Hz (33ms) to prevent UI flooding
	now := time.Now()
	if now.Sub(m.Operations.Progress.LastProgressUpdate) < 33*time.Millisecond && msg.Percent < 1.0 {
		return listenToProgress(msg.Channel)
	}

	m.Operations.Progress.Show(msg.Label)
	m.Operations.Progress.Update(msg.Percent)
	m.Operations.Progress.LastProgressUpdate = now
	return listenToProgress(msg.Channel)
}

func finalizeOperation(m *tui_context.Model, msg OperationFinishedMsg) tea.Cmd {
	m.UI.Loading = false
	m.Operations.Progress.Update(1.0)
	m.Operations.ConflictPolicy = conflict.Ask
	m.Operations.Conflict.Clear()
	for i := range m.Logs.Entries {
		if m.Logs.Entries[i].ID == msg.LogID {
			msgText := m.Logs.Entries[i].Message
			if strings.HasPrefix(msgText, "Pasting ") {
				msgText = "Pasted " + msgText[8:]
			} else if strings.HasPrefix(msgText, "Moving ") {
				msgText = "Moved " + msgText[7:]
			} else if strings.HasPrefix(msgText, "Deleting ") {
				msgText = "Deleted " + msgText[9:]
			} else if strings.HasPrefix(msgText, "Zipping ") {
				msgText = "Zipped " + msgText[8:]
			} else if strings.HasPrefix(msgText, "Extracting ") {
				msgText = "Unzipped " + msgText[11:]
			}
			LogUpdate(m, msg.LogID, tui_context.StatusSuccess, tui_context.LogSuccess, msgText, "")
			break
		}
	}

	for _, p := range msg.Paths {
		m.Navigation.Deselect(p)
	}
	m.UI.SelectMode = m.Navigation.SelectedCount > 0

	return tea.Batch(
		Reload(m),
		tea.Tick(constants.ProgressDisplayDuration, func(time.Time) tea.Msg {
			return ClearMsg{}
		}),
	)
}

func handleConflict(m *tui_context.Model, msg ConflictMsg) tea.Cmd {
	m.UI.Loading = false
	m.Operations.Conflict.Set(msg.Src, msg.Dst, msg.PendingItems, msg.IsMove, msg.OpType, msg.LogID)
	m.Operations.ActionType = constants.ActionConflict
	m.UI.StartConfirming()
	return nil
}

// PerformRename executes a rename operation
func PerformRename(m *tui_context.Model, newName string) tea.Cmd {
	if newName == "" {
		return nil
	}

	// Validate filename
	if err := ops.ValidateFileName(newName); err != nil {
		return SetErrMsg(m, "Invalid filename: "+err.Error())
	}

	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}

	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
	oldPath := selected.Path
	newPath := m.FS.Join(m.Navigation.Path, newName)

	logID := LogPush(m, "Rename", tui_context.LogInfo, tui_context.StatusRunning,
		fmt.Sprintf("Renaming %s to %s", selected.Name, newName),
		fmt.Sprintf("From: %s\nTo: %s", oldPath, newPath))

	c, cancel := context.WithTimeout(m.Context, constants.DirectoryLoadTimeout)
	defer cancel()

	if err := ops.Rename(c, m.FS, oldPath, newPath, conflict.Ask); err != nil {
		LogUpdate(m, logID, tui_context.StatusError, tui_context.LogError,
			fmt.Sprintf("Failed to rename %s to %s", selected.Name, newName), err.Error())
		return LogError(m, err, "Rename")
	}

	LogUpdate(m, logID, tui_context.StatusSuccess, tui_context.LogSuccess,
		fmt.Sprintf("Renamed %s to %s", selected.Name, newName), "")
	return Reload(m)
}

// --- Commands ---

func listenToProgress(progChan chan core.Progress) tea.Cmd {
	return func() tea.Msg {
		prog, ok := <-progChan
		if !ok {
			return nil
		}
		return ProgressMsg{
			Percent: prog.Percent,
			Label:   prog.Label,
			Channel: progChan,
		}
	}
}

func deleteItems(ctx context.Context, fs core.FileSystem, targets []string, useTrash bool, logID string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.DeleteMultiple(ctx, fs, targets, useTrash, progChan)
			if err != nil {
				return ErrorMsg{Err: err, LogID: logID}
			}
			return OperationFinishedMsg{Paths: targets, LogID: logID}
		},
	)
}

func pasteItems(ctx context.Context, srcFS, dstFS core.FileSystem, sources []string, destDir string, policy conflict.Policy, logID string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.CopyMultiple(ctx, srcFS, sources, destDir, progChan, policy)
			if err != nil {
				var conflict *conflict.ConflictError
				if errors.As(err, &conflict) {
					return ConflictMsg{
						Src:          conflict.Source,
						Dst:          conflict.Destination,
						PendingItems: conflict.PendingItems,
						IsMove:       false,
						OpType:       "copy",
						LogID:        logID,
					}
				}
				return ErrorMsg{Err: err, LogID: logID}
			}
			return OperationFinishedMsg{Paths: sources, LogID: logID}
		},
	)
}

func moveItems(ctx context.Context, srcFS, dstFS core.FileSystem, sources []string, destDir string, policy conflict.Policy, logID string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.MoveMultiple(ctx, srcFS, sources, destDir, progChan, policy)
			if err != nil {
				var conflict *conflict.ConflictError
				if errors.As(err, &conflict) {
					return ConflictMsg{
						Src:          conflict.Source,
						Dst:          conflict.Destination,
						PendingItems: conflict.PendingItems,
						IsMove:       true,
						OpType:       "move",
						LogID:        logID,
					}
				}
				return ErrorMsg{Err: err, LogID: logID}
			}
			return OperationFinishedMsg{Paths: sources, LogID: logID}
		},
	)
}

func startUnzip(m *tui_context.Model) tea.Cmd {
	targets := getTargets(m)
	var zipPath string
	if len(targets) > 0 {
		// Only unzip the first selected item if it's a zip
		if strings.HasSuffix(strings.ToLower(targets[0]), ".zip") {
			zipPath = targets[0]
		}
	} else if len(m.Navigation.FilteredItems) > 0 {
		selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
		if !selected.IsUp && strings.HasSuffix(strings.ToLower(selected.Name), ".zip") {
			zipPath = selected.Path
		}
	}

	if zipPath == "" {
		return SetMsg(m, "Please select a .zip file to unzip")
	}

	m.StartInput(tui_context.InputUnzip)

	// Default destination is a folder with same name as zip
	baseName := m.FS.Base(zipPath)
	folderName := strings.TrimSuffix(baseName, m.FS.Ext(baseName))
	m.Inputs.ActiveInput.SetValue(folderName)
	return m.Inputs.ActiveInput.FocusCmd()
}

// PerformUnzip executes an unzip operation
func PerformUnzip(m *tui_context.Model, destName string) tea.Cmd {
	if destName == "" {
		return nil
	}

	// We need the source zip path again
	targets := getTargets(m)
	var zipPath string
	if len(targets) > 0 {
		zipPath = targets[0]
	} else {
		selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
		zipPath = selected.Path
	}

	dst := m.FS.Join(m.Navigation.Path, destName)

	// Check for conflict
	if m.Operations.ConflictPolicy == conflict.Ask {
		resolver := conflict.NewResolver()
		resolvedPath, _, err := resolver.Resolve(m.Context, m.FS, zipPath, dst, m.Operations.ConflictPolicy)
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				m.UI.Loading = false
				m.Operations.Conflict.Set(cerr.Source, cerr.Destination, []string{zipPath}, false, "unzip", "")
				m.Operations.ActionType = constants.ActionConflict
				m.UI.StartConfirming()
				return nil
			}
		}
		if resolvedPath == "" {
			return nil // Skip
		}
		dst = resolvedPath
	}

	m.ClearSelection()
	m.UI.Loading = true

	logID := LogPush(m, "Unzip", tui_context.LogInfo, tui_context.StatusRunning,
		fmt.Sprintf("Extracting %s into %s", m.FS.Base(zipPath), destName), "")

	ctx, cancel := context.WithCancel(m.Context)
	m.Operations.CancelFunc = cancel

	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.Unzip(ctx, m.FS, zipPath, dst, progChan, m.Operations.ConflictPolicy)
			if err != nil {
				return ErrorMsg{Err: err, LogID: logID}
			}
			return OperationFinishedMsg{Paths: []string{}, LogID: logID}
		},
	)
}
