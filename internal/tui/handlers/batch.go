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
	case messages.LogPushMsg:
		logID := utils.LogPush(m, tuictx.LogEntry{
			Type:    msg.Type,
			Level:   tuictx.LogInfo,
			Status:  tuictx.StatusRunning,
			Message: msg.Message,
		})
		return file.DeleteItems(ops.DeleteOptions{
			OpCtx: ops.OpContext{Context: m.Context, FS: m.FS},
			Paths: msg.Targets,
		}, logID), true

	case messages.PerformPasteMsg:
		logID := utils.LogPush(m, tuictx.LogEntry{
			Type:    msg.OpName,
			Level:   tuictx.LogInfo,
			Status:  tuictx.StatusRunning,
			Message: msg.Message,
		})
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel

		batchOpts := ops.BatchOptions{
			OpCtx:    ops.OpContext{Context: ctx, FS: m.FS},
			SrcFS:    m.Operations.Clipboard.SourceFS,
			Sources:  msg.Paths,
			DestDir:  msg.DestDir,
			Conflict: ops.ConflictOptions{Policy: m.Operations.ConflictPolicy},
		}

		if msg.IsCut {
			m.Operations.Clipboard.Clear()
			return file.MoveItems(batchOpts, logID), true
		}
		return file.PasteItems(batchOpts, logID), true

	case messages.PerformZipMsg:
		logID := utils.LogPush(m, tuictx.LogEntry{
			Type:    "Zip",
			Level:   tuictx.LogInfo,
			Status:  tuictx.StatusRunning,
			Message: msg.Message,
		})
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel
		progChan := make(chan core.Progress, 100)
		return tea.Batch(
			file.ListenToProgress(progChan),
			func() tea.Msg {
				defer close(progChan)
				err := ops.Zip(ops.ZipOptions{
					OpCtx:    ops.OpContext{Context: ctx, FS: m.FS, Progress: progChan},
					Srcs:     msg.Targets,
					Dst:      msg.Dst,
					Conflict: ops.ConflictOptions{Policy: m.Operations.ConflictPolicy},
				})
				if err != nil {
					return messages.ErrorMsg{Err: err, LogID: logID}
				}
				return messages.OperationFinishedMsg{Paths: []string{}, LogID: logID}
			},
		), true

	case messages.PerformUnzipMsg:
		logID := utils.LogPush(m, tuictx.LogEntry{
			Type:    "Unzip",
			Level:   tuictx.LogInfo,
			Status:  tuictx.StatusRunning,
			Message: msg.Message,
		})
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel
		progChan := make(chan core.Progress, 100)
		return tea.Batch(
			file.ListenToProgress(progChan),
			func() tea.Msg {
				defer close(progChan)
				err := ops.Unzip(ops.ZipOptions{
					OpCtx:    ops.OpContext{Context: ctx, FS: m.FS, Progress: progChan},
					Src:      msg.ZipPath,
					Dst:      msg.Dst,
					Conflict: ops.ConflictOptions{Policy: m.Operations.ConflictPolicy},
				})
				if err != nil {
					return messages.ErrorMsg{Err: err, LogID: logID}
				}
				return messages.OperationFinishedMsg{Paths: []string{}, LogID: logID}
			},
		), true

	case messages.PerformRenameMsg:
		logID := utils.LogPush(m, tuictx.LogEntry{
			Type:    "Rename",
			Level:   tuictx.LogInfo,
			Status:  tuictx.StatusRunning,
			Message: fmt.Sprintf("Renaming %s to %s", msg.Selected.Name, msg.NewName),
			Details: fmt.Sprintf("From: %s\nTo: %s", msg.OldPath, msg.NewPath),
		})

		ctx, cancel := context.WithCancel(m.Context)
		defer cancel()
		if err := ops.Rename(ops.RenameOptions{
			OpCtx:    ops.OpContext{Context: ctx, FS: m.FS},
			OldPath:  msg.OldPath,
			NewPath:  msg.NewPath,
			Conflict: ops.ConflictOptions{Policy: m.Operations.ConflictPolicy},
		}); err != nil {
			utils.LogUpdate(m, logID, tuictx.LogEntry{
				Status:  tuictx.StatusError,
				Level:   tuictx.LogError,
				Message: fmt.Sprintf("Failed to rename %s to %s", msg.Selected.Name, msg.NewName),
				Details: err.Error(),
			})
			return utils.LogError(m, err, "Rename"), true
		}

		utils.LogUpdate(m, logID, tuictx.LogEntry{
			Status:  tuictx.StatusSuccess,
			Level:   tuictx.LogSuccess,
			Message: fmt.Sprintf("Renamed %s to %s", msg.Selected.Name, msg.NewName),
		})
		return tea.Batch(
			utils.SetMsg(m, fmt.Sprintf("Renamed %s to %s", msg.Selected.Name, msg.NewName)),
			func() tea.Msg { return messages.ReloadMsg{} },
		), true
	}
	return nil, false
}
