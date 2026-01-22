package file

import (
	"context"
	"fmt"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func StartCreate(m *tui_context.Model) tea.Cmd {
	m.StartInput(tui_context.InputCreate)
	m.Inputs.AltMode = false // false = File, true = Folder
	m.Inputs.ActiveInput.SetValue("")
	return m.Inputs.ActiveInput.FocusCmd()
}

func PerformCreate(m *tui_context.Model, name string) tea.Cmd {
	if name == "" {
		return nil
	}

	if err := ops.ValidateFileName(name); err != nil {
		return func() tea.Msg { return messages.ErrorMsg{Err: err} }
	}

	isFolder := m.Inputs.AltMode
	path := m.FS.Join(m.Navigation.Path, name)

	ctx, cancel := context.WithTimeout(m.Context, constants.DirectoryLoadTimeout)
	defer cancel()

	resolvedPath, err := ops.CreateAtomic(ops.CreateOptions{
		OpCtx:    ops.OpContext{Context: ctx, FS: m.FS},
		Path:     path,
		IsDir:    isFolder,
		Conflict: ops.ConflictOptions{Policy: m.Operations.ConflictPolicy},
	})
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			m.UI.Loading = false
			m.Operations.Conflict.Set(tui_context.ConflictParams{
				Source:       "",
				Destination:  cerr.Destination,
				PendingItems: []string{cerr.Destination},
				IsMove:       false,
				OpType:       "create",
			})
			m.Operations.ActionType = constants.ActionConflict
			m.UI.StartConfirming()
			return nil
		}
		return func() tea.Msg { return messages.ErrorMsg{Err: err} }
	}

	if resolvedPath == "" {
		return nil // Skip
	}

	msg := "File created"
	if isFolder {
		msg = "Folder created"
	}
	return func() tea.Msg { return messages.StatusMsg{Message: msg} }
}

func PerformDelete(m *tui_context.Model) tea.Cmd {
	targets := GetTargets(m)
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
			FormatDisplayPath(m, m.FS, m.Navigation.Path))
	} else {
		msg = fmt.Sprintf("Deleting %d items from %s",
			len(targets),
			FormatDisplayPath(m, m.FS, m.Navigation.Path))
	}

	return func() tea.Msg { return messages.LogPushMsg{Type: "Delete", Message: msg, Targets: targets} }
}

func StartRename(m *tui_context.Model) tea.Cmd {
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

func PerformRename(m *tui_context.Model, newName string) tea.Cmd {
	if newName == "" {
		return nil
	}

	if err := ops.ValidateFileName(newName); err != nil {
		return func() tea.Msg { return messages.ErrorMsg{Err: err} }
	}

	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}

	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
	oldPath := selected.Path
	newPath := m.FS.Join(m.Navigation.Path, newName)

	return func() tea.Msg {
		return messages.PerformRenameMsg{Selected: selected, OldPath: oldPath, NewPath: newPath, NewName: newName}
	}
}

func DeleteItems(opts ops.DeleteOptions, logID string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	opts.OpCtx.Progress = progChan
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.DeleteMultiple(opts)
			if err != nil {
				return messages.ErrorMsg{Err: err, LogID: logID}
			}
			return messages.OperationFinishedMsg{Paths: opts.Paths, LogID: logID}
		},
	)
}

func PerformConflictRename(m *tui_context.Model, newName string) tea.Cmd {
	if newName == "" {
		return nil
	}

	if err := ops.ValidateFileName(newName); err != nil {
		return func() tea.Msg { return messages.ErrorMsg{Err: err} }
	}

	oldDst := m.Operations.Conflict.Destination
	newDst := m.FS.Join(m.FS.Dir(oldDst), newName)
	m.Operations.Conflict.Destination = newDst

	return ResolveConflict(m, "rename", false)
}
