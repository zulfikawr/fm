package file

import (
	"context"
	"errors"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleFileOps handles file system and operation messages
func HandleFileOps(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.Confirming {
			return HandleConfirmKeys(m, msg)
		}
		return HandleFileKeys(m, msg)

	case messages.ProgressMsg:
		return HandleProgress(m, msg)

	case messages.OperationFinishedMsg:
		return FinalizeOperation(m, msg)

	case messages.ConflictMsg:
		return HandleConflict(m, msg)
	}
	return nil
}

func HandleFileKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	if m.UI.InputActive || m.UI.SettingsOpen || m.UI.LogOpen || m.UI.ClipboardOpen || m.UI.TrashOpen {
		return nil
	}

	key := msg.String()
	action := GetActionForKeyFromModel(m, key)

	switch action {
	case "copy":
		return CopySelected(m)
	case "cut":
		return CutSelected(m)
	case "paste":
		return PerformPaste(m)
	case "delete":
		return PerformDelete(m)
	case "rename":
		return StartRename(m)
	case "zip":
		return StartZip(m)
	case "unzip":
		return StartUnzip(m)
	}
	return nil
}

func HandleConfirmKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	action := m.Operations.ActionType

	if action == constants.ActionConflict {
		switch msg.String() {
		case "y":
			return ResolveConflict(m, "overwrite", false)
		case "Y":
			return ResolveConflict(m, "overwrite", true)
		case "n":
			return ResolveConflict(m, "skip", false)
		case "N":
			return ResolveConflict(m, "skip", true)
		case "r":
			m.UI.StopConfirming()
			return func() tea.Msg { return messages.StartConflictRenameMsg{} }
		case "R":
			return ResolveConflict(m, "rename", true)
		case "esc":
			m.UI.StopConfirming()
			m.Operations.ActionType = constants.ActionNone
			return nil
		}
		return nil
	}

	switch msg.String() {
	case "l", "L":
		if action == constants.ActionGoto {
			m.UI.StopConfirming()
			m.Operations.ActionType = constants.ActionNone
			m.StartInput(tui_context.InputGoto)
			m.Inputs.AltMode = false
			m.Inputs.ActiveInput.SetValue(m.Navigation.Path)
			return m.Inputs.ActiveInput.FocusCmd()
		}
	case "r", "R":
		if action == constants.ActionGoto {
			m.UI.StopConfirming()
			m.Operations.ActionType = constants.ActionNone
			m.StartInput(tui_context.InputGoto)
			m.Inputs.AltMode = true
			m.Inputs.ActiveInput.SetValue("")
			return m.Inputs.ActiveInput.FocusCmd()
		}
	case "p", "P":
		if action == constants.ActionAuth {
			m.UI.StopConfirming()
			m.Operations.ActionType = constants.ActionNone
			m.StartInput(tui_context.InputAuth)
			m.Inputs.AltMode = false
			m.Inputs.ActiveInput.EchoMode = ui.EchoPassword
			m.Inputs.ActiveInput.SetPrompt("Password: ")
			return m.Inputs.ActiveInput.FocusCmd()
		}
	case "k", "K":
		if action == constants.ActionAuth {
			m.UI.StopConfirming()
			m.Operations.ActionType = constants.ActionNone
			m.StartInput(tui_context.InputAuth)
			m.Inputs.AltMode = true
			m.Inputs.ActiveInput.EchoMode = ui.EchoNormal
			m.Inputs.ActiveInput.SetPrompt("PEM Path: ")
			return m.Inputs.ActiveInput.FocusCmd()
		}
	case "y", "Y":
		m.UI.StopConfirming()

		switch action {
		case constants.ActionDelete:
			return PerformDelete(m)
		case constants.ActionPaste:
			return PerformPaste(m)
		case constants.ActionResetSettings:
			return func() tea.Msg { return messages.ResetSettingsMsg{} }
		case constants.ActionTestIcons:
			return func() tea.Msg { return messages.IconTestMsg{Success: true} }
		}
		m.Operations.ActionType = constants.ActionNone
	case "n", "N", "esc":
		m.UI.StopConfirming()
		if action == constants.ActionTestIcons {
			m.Operations.ActionType = constants.ActionNone
			return func() tea.Msg { return messages.IconTestMsg{Success: false} }
		}
		m.Operations.ActionType = constants.ActionNone
	}
	return nil
}

func ResolveConflict(m *tui_context.Model, choice string, applyToAll bool) tea.Cmd {
	var cmds []tea.Cmd

	ctx, cancel := context.WithCancel(m.Context)
	m.Operations.CancelFunc = cancel

	srcFS := m.Operations.Clipboard.SourceFS
	dstFS := m.FS
	logID := m.Operations.Conflict.LogID
	opType := m.Operations.Conflict.OpType
	dst := m.Operations.Conflict.Destination

	pending := m.Operations.Conflict.PendingItems

	m.Operations.Conflict.ApplyToAll = applyToAll

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

	resolver := conflict.NewResolver()
	src := m.Operations.Conflict.Source
	resolvedPath, _, err := resolver.Resolve(m.Context, m.FS, conflict.ResolveOptions{
		Src:    src,
		Dst:    dst,
		Policy: policy,
	})
	if err == nil {
		if resolvedPath == "" && choice == "skip" {
			if len(pending) <= 1 && !applyToAll {
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				m.Operations.ConflictPolicy = conflict.Ask
				return func() tea.Msg { return messages.ReloadMsg{} }
			}
			pending = pending[1:]
		} else if resolvedPath != "" {
			dst = resolvedPath
		}
	}

	m.UI.StopConfirming()
	m.Operations.ActionType = constants.ActionNone

	switch opType {
	case "create":
		name := m.FS.Base(dst)
		cmds = append(cmds, PerformCreate(m, name))
	case "zip":
		zipName := m.FS.Base(dst)
		cmds = append(cmds, PerformZip(m, zipName))
	case "unzip":
		destName := m.FS.Base(dst)
		cmds = append(cmds, PerformUnzip(m, destName))
	case "move":
		m.UI.Loading = true
		cmds = append(cmds, MoveItems(ops.BatchOptions{
			OpCtx:    ops.OpContext{Context: ctx, FS: dstFS},
			SrcFS:    srcFS,
			Sources:  pending,
			DestDir:  m.Navigation.Path,
			Conflict: ops.ConflictOptions{Policy: m.Operations.ConflictPolicy, ApplyToAll: applyToAll},
		}, logID))
	case "copy":
		m.UI.Loading = true
		cmds = append(cmds, PasteItems(ops.BatchOptions{
			OpCtx:    ops.OpContext{Context: ctx, FS: dstFS},
			SrcFS:    srcFS,
			Sources:  pending,
			DestDir:  m.Navigation.Path,
			Conflict: ops.ConflictOptions{Policy: m.Operations.ConflictPolicy, ApplyToAll: applyToAll},
		}, logID))
	}

	if !applyToAll {
		m.Operations.ConflictPolicy = conflict.Ask
	}

	return tea.Batch(cmds...)
}

func HandleProgress(m *tui_context.Model, msg messages.ProgressMsg) tea.Cmd {
	now := time.Now()
	if now.Sub(m.Operations.Progress.LastProgressUpdate) < 33*time.Millisecond && msg.Percent < 1.0 {
		return ListenToProgress(msg.Channel)
	}

	m.Operations.Progress.Show(msg.Label)
	m.Operations.Progress.Update(msg.Percent)
	m.Operations.Progress.LastProgressUpdate = now
	return ListenToProgress(msg.Channel)
}

func FinalizeOperation(m *tui_context.Model, msg messages.OperationFinishedMsg) tea.Cmd {
	m.UI.Loading = false
	m.Operations.Progress.Update(1.0)
	m.Operations.ConflictPolicy = conflict.Ask
	m.Operations.Conflict.Clear()

	for _, p := range msg.Paths {
		m.Navigation.Deselect(p)
	}
	m.Navigation.SelectMode = m.Navigation.SelectedCount() > 0

	return func() tea.Msg { return messages.OperationFinishedEventMsg{LogID: msg.LogID, Paths: msg.Paths} }
}

func HandleConflict(m *tui_context.Model, msg messages.ConflictMsg) tea.Cmd {
	m.UI.Loading = false
	m.Operations.Conflict.Set(tui_context.ConflictParams{
		Source:       msg.Src,
		Destination:  msg.Dst,
		PendingItems: msg.PendingItems,
		IsMove:       msg.IsMove,
		OpType:       msg.OpType,
		LogID:        msg.LogID,
	})
	m.Operations.ActionType = constants.ActionConflict
	m.UI.StartConfirming()
	return nil
}

func ListenToProgress(progChan chan core.Progress) tea.Cmd {
	return func() tea.Msg {
		prog, ok := <-progChan
		if !ok {
			return nil
		}
		return messages.ProgressMsg{
			Percent: prog.Percent,
			Label:   prog.Label,
			Channel: progChan,
		}
	}
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

func PasteItems(opts ops.BatchOptions, logID string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	newOpts := opts
	newOpts.OpCtx.Progress = progChan
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.CopyMultiple(newOpts)
			if err != nil {
				var conflict *conflict.ConflictError
				if errors.As(err, &conflict) {
					return messages.ConflictMsg{
						Src:          conflict.Source,
						Dst:          conflict.Destination,
						PendingItems: conflict.PendingItems,
						IsMove:       false,
						OpType:       "copy",
						LogID:        logID,
					}
				}
				return messages.ErrorMsg{Err: err, LogID: logID}
			}
			return messages.OperationFinishedMsg{Paths: newOpts.Sources, LogID: logID}
		},
	)
}

func MoveItems(opts ops.BatchOptions, logID string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	newOpts := opts
	newOpts.OpCtx.Progress = progChan
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)
			err := ops.MoveMultiple(newOpts)
			if err != nil {
				var conflict *conflict.ConflictError
				if errors.As(err, &conflict) {
					return messages.ConflictMsg{
						Src:          conflict.Source,
						Dst:          conflict.Destination,
						PendingItems: conflict.PendingItems,
						IsMove:       true,
						OpType:       "move",
						LogID:        logID,
					}
				}
				return messages.ErrorMsg{Err: err, LogID: logID}
			}
			return messages.OperationFinishedMsg{Paths: newOpts.Sources, LogID: logID}
		},
	)
}
