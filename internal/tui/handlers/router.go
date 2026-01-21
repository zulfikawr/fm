package handlers

import (
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/file"
	"github.com/zulfikawr/fm/internal/tui/handlers/integration"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

var OpenFileAction = file.OpenFile
var OpenFileAtLineAction = file.OpenFileAtLine

// HandleUpdate is the main message dispatcher (The Hub)
func HandleUpdate(m *tuictx.Model, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Synchronize viewport height to ensure correct scrolling calculations
	m.SyncViewportHeight()

	// 0. Update UI components (like spinners) - Always run this first
	var spinnerCmd tea.Cmd
	m.Display.LoadingSpinner, spinnerCmd = m.Display.LoadingSpinner.Update(msg)
	if spinnerCmd != nil {
		cmds = append(cmds, spinnerCmd)
	}

	// 1. Global/System Handlers (Resize, Quit, Tick, Toggles)
	if cmd, handled := handleGlobal(m, msg); handled {
		return cmd
	}

	// 2. Input Handlers (Highest Priority when active)
	if cmd, handled := handleInputs(m, msg); handled {
		return cmd
	}

	// 3. Batch/Complex Operation Handlers
	if cmd, handled := handleBatch(m, msg); handled {
		return cmd
	}

	// 4. Signal/Event Handlers
	if cmd, handled := handleEvents(m, msg); handled {
		return cmd
	}

	// 5. Domain-Specific Handlers
	switch msg := msg.(type) {
	case messages.ReloadMsg:
		return nav.Reload(m, msg.Silent)

	case messages.NavigateMsg:
		return nav.NavigateToPath(m, msg.Path)

	case messages.RemoteGotoMsg:
		return integration.HandleRemoteGoto(m, msg.Input)

	case messages.StatusMsg:
		if msg.IsError {
			return utils.SetErrMsg(m, msg.Message)
		}
		return utils.SetMsg(m, msg.Message)

	case messages.StartCreateMsg:
		return file.StartCreate(m)

	case messages.StartConflictRenameMsg:
		m.StartInput(tuictx.InputConflictRename)
		m.Inputs.ActiveInput.SetValue(m.FS.Base(m.Operations.Conflict.Destination))
		return m.Inputs.ActiveInput.FocusCmd()

	case messages.ResetSettingsMsg:
		return app.ConfirmSettingsReset(m)

	case messages.TabLimitMsg:
		return utils.SetMsg(m, "Tab limit reached (max 9 tabs)")

	case messages.OpenFileMsg:
		return OpenFileAction(m, msg.Item)

	case messages.WatchDirMsg:
		return nav.WatchDirAction(m)

	case messages.ClearMsg:
		m.Operations.Progress.Hide()
		m.Message.Pop()
		return nil
	}

	// Priority update for specialized messages
	if cmd := app.HandleUpdateMessages(m, msg); cmd != nil {
		return cmd
	}

	// Delegate to sub-package handlers
	if cmd := nav.HandleNavigation(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := file.HandleFileOps(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := integration.HandleGit(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := integration.HandleRemote(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := integration.HandleSearch(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := app.HandleSettings(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := app.HandleLogs(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}
