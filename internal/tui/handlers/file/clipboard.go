package file

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/conflict"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func CopySelected(m *tuictx.Model) tea.Cmd {
	targets := GetTargets(m)
	if len(targets) == 0 {
		return nil
	}
	m.Operations.Clipboard.SetCopy(m.FS, targets)
	m.ClearSelection()
	msg := fmt.Sprintf("Copied %d items to clipboard", len(targets))
	return func() tea.Msg { return messages.StatusMsg{Message: msg} }
}

func CutSelected(m *tuictx.Model) tea.Cmd {
	targets := GetTargets(m)
	if len(targets) == 0 {
		return nil
	}
	m.Operations.Clipboard.SetCut(m.FS, targets)
	m.ClearSelection()
	msg := fmt.Sprintf("Cut %d items to clipboard", len(targets))
	return func() tea.Msg { return messages.StatusMsg{Message: msg} }
}

func PerformPaste(m *tuictx.Model) tea.Cmd {
	paths := m.Operations.Clipboard.Paths
	if len(paths) == 0 {
		return func() tea.Msg { return messages.ErrorMsg{Err: fmt.Errorf("clipboard is empty")} }
	}

	if m.Config.Ops.ConfirmOperations && m.Operations.ActionType != constants.ActionPaste {
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
			FormatDisplayPath(m, srcFS, srcPath),
			FormatDisplayPath(m, dstFS, destDir))
	} else {
		msg = fmt.Sprintf("%s %d items from %s to %s",
			opVerb,
			len(paths),
			FormatDisplayPath(m, srcFS, srcPath),
			FormatDisplayPath(m, dstFS, destDir))
	}

	return func() tea.Msg {
		return messages.PerformPasteMsg{OpName: opName, Message: msg, Paths: paths, DestDir: destDir, IsCut: isCut}
	}
}
