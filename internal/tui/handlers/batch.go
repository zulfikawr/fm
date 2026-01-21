package handlers

import (
	"context"
	"fmt"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/file"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// handleBatch handles complex, long-running operations
func handleBatch(m *tuictx.Model, msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case messages.PerformPasteMsg:
		logID := utils.LogPush(m, msg.OpName, tuictx.LogInfo, tuictx.StatusRunning, msg.Message, "")
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel

		if msg.IsCut {
			m.Operations.Clipboard.Clear()
			return file.MoveItems(ctx, m.Operations.Clipboard.SourceFS, m.FS, msg.Paths, msg.DestDir, m.Operations.ConflictPolicy, false, logID), true
		}
		return file.PasteItems(ctx, m.Operations.Clipboard.SourceFS, m.FS, msg.Paths, msg.DestDir, m.Operations.ConflictPolicy, false, logID), true

	case messages.PerformZipMsg:
		logID := utils.LogPush(m, "Zip", tuictx.LogInfo, tuictx.StatusRunning, msg.Message, "")
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel
		progChan := make(chan core.Progress, 100)
		return tea.Batch(
			file.ListenToProgress(progChan),
			func() tea.Msg {
				defer close(progChan)
				err := ops.Zip(ctx, m.FS, msg.Targets, msg.Dst, progChan, m.Operations.ConflictPolicy)
				if err != nil {
					return messages.ErrorMsg{Err: err, LogID: logID}
				}
				return messages.OperationFinishedMsg{Paths: []string{}, LogID: logID}
			},
		), true

	case messages.PerformUnzipMsg:
		logID := utils.LogPush(m, "Unzip", tuictx.LogInfo, tuictx.StatusRunning, msg.Message, "")
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel
		progChan := make(chan core.Progress, 100)
		return tea.Batch(
			file.ListenToProgress(progChan),
			func() tea.Msg {
				defer close(progChan)
				err := ops.Unzip(ctx, m.FS, msg.ZipPath, msg.Dst, progChan, m.Operations.ConflictPolicy)
				if err != nil {
					return messages.ErrorMsg{Err: err, LogID: logID}
				}
				return messages.OperationFinishedMsg{Paths: []string{}, LogID: logID}
			},
		), true

	case messages.PerformRenameMsg:
		logID := utils.LogPush(m, "Rename", tuictx.LogInfo, tuictx.StatusRunning,
			fmt.Sprintf("Renaming %s to %s", msg.Selected.Name, msg.NewName),
			fmt.Sprintf("From: %s\nTo: %s", msg.OldPath, msg.NewPath))

		ctx, cancel := context.WithCancel(m.Context)
		defer cancel()

		if err := ops.Rename(ctx, m.FS, msg.OldPath, msg.NewPath, m.Operations.ConflictPolicy); err != nil {
			utils.LogUpdate(m, logID, tuictx.StatusError, tuictx.LogError,
				fmt.Sprintf("Failed to rename %s to %s", msg.Selected.Name, msg.NewName), err.Error())
			return utils.LogError(m, err, "Rename"), true
		}

		utils.LogUpdate(m, logID, tuictx.StatusSuccess, tuictx.LogSuccess,
			fmt.Sprintf("Renamed %s to %s", msg.Selected.Name, msg.NewName), "")
		return tea.Batch(
			utils.SetMsg(m, fmt.Sprintf("Renamed %s to %s", msg.Selected.Name, msg.NewName)),
			func() tea.Msg { return messages.ReloadMsg{} },
		), true
	}
	return nil, false
}
