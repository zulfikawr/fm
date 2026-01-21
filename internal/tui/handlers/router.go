package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
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

// HandleUpdate is the main message dispatcher (Reducer)
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

	// 1. Priority/Global Handlers
	switch msg := msg.(type) {
	case app.TickMsg:
		return tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.Display.Width = msg.Width
		m.Display.Height = msg.Height
		m.SyncViewportHeight()
		return tea.Batch(cmds...)

	case messages.WatchEventMsg:
		if m.Watcher.DebounceTimer != nil {
			m.Watcher.DebounceTimer.Stop()
		}
		m.Watcher.DebounceTimer = time.NewTimer(150 * time.Millisecond)
		return func() tea.Msg {
			<-m.Watcher.DebounceTimer.C
			return messages.DebounceWatchMsg{}
		}

	case messages.DebounceWatchMsg:
		m.Watcher.IsListening = false
		return nav.Reload(m, false)

	case messages.WatcherErrorMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds,
			utils.SetErrMsg(m, "Watcher error: restarting..."),
			utils.RestartWatcherAction(m),
		)
		return tea.Batch(cmds...)

	case messages.WatcherClosedMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds,
			utils.SetMsg(m, "Watcher closed: restarting..."),
			utils.RestartWatcherAction(m),
		)
		return tea.Batch(cmds...)

	case messages.RemotePollMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds, nav.Reload(m, true))
		return tea.Batch(cmds...)

	case messages.DebounceFilterMsg:
		if msg.Generation == m.Navigation.FilterGen {
			nav.ApplyFilter(m)
		}
		return nil

	case messages.ClearMsg:
		m.Operations.Progress.Hide()
		m.Message.Pop()
		return tea.Batch(cmds...)

	case messages.ErrorMsg:
		// Update original log entry with failure message
		for i := range m.Logs.Entries {
			if m.Logs.Entries[i].ID == msg.LogID {
				msgText := m.Logs.Entries[i].Message
				if strings.HasPrefix(msgText, "Pasting ") {
					msgText = "Failed to paste " + msgText[8:]
				} else if strings.HasPrefix(msgText, "Moving ") {
					msgText = "Failed to move " + msgText[7:]
				} else if strings.HasPrefix(msgText, "Deleting ") {
					msgText = "Failed to delete " + msgText[9:]
				} else {
					msgText = "Failed: " + msgText
				}
				utils.LogUpdate(m, msg.LogID, tuictx.StatusError, tuictx.LogError, msgText, msg.Err.Error())
				break
			}
		}

		m.UI.Loading = false
		m.Operations.ProcessingItems = make(map[string]bool)

		cmds = append(cmds,
			utils.LogError(m, msg.Err, "Operation failed"),
			nav.Reload(m, false),
			tea.Tick(constants.ProgressDisplayDuration, func(time.Time) tea.Msg {
				return messages.ClearMsg{}
			}),
		)
		return tea.Batch(cmds...)

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

	case messages.PerformPasteMsg:
		logID := utils.LogPush(m, msg.OpName, tuictx.LogInfo, tuictx.StatusRunning, msg.Message, "")
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel

		if msg.IsCut {
			m.Operations.Clipboard.Clear()
			return file.MoveItems(ctx, m.Operations.Clipboard.SourceFS, m.FS, msg.Paths, msg.DestDir, m.Operations.ConflictPolicy, false, logID)
		}
		return file.PasteItems(ctx, m.Operations.Clipboard.SourceFS, m.FS, msg.Paths, msg.DestDir, m.Operations.ConflictPolicy, false, logID)

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
		)

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
		)

	case messages.LogPushMsg:
		logID := utils.LogPush(m, msg.Type, tuictx.LogInfo, tuictx.StatusRunning, msg.Message, "")
		ctx, cancel := context.WithCancel(m.Context)
		m.Operations.CancelFunc = cancel
		return file.DeleteItems(ctx, m.FS, msg.Targets, m.Config.UseTrash, logID)

	case messages.PerformRenameMsg:
		logID := utils.LogPush(m, "Rename", tuictx.LogInfo, tuictx.StatusRunning,
			fmt.Sprintf("Renaming %s to %s", msg.Selected.Name, msg.NewName),
			fmt.Sprintf("From: %s\nTo: %s", msg.OldPath, msg.NewPath))

		ctx, cancel := context.WithTimeout(m.Context, constants.DirectoryLoadTimeout)
		defer cancel()

		if err := ops.Rename(ctx, m.FS, msg.OldPath, msg.NewPath, conflict.Ask); err != nil {
			utils.LogUpdate(m, logID, tuictx.StatusError, tuictx.LogError,
				fmt.Sprintf("Failed to rename %s to %s", msg.Selected.Name, msg.NewName), err.Error())
			return utils.LogError(m, err, "Rename")
		}

		utils.LogUpdate(m, logID, tuictx.StatusSuccess, tuictx.LogSuccess,
			fmt.Sprintf("Renamed %s to %s", msg.Selected.Name, msg.NewName), "")
		return tea.Batch(
			utils.SetMsg(m, fmt.Sprintf("Renamed %s to %s", msg.Selected.Name, msg.NewName)),
			nav.Reload(m, false),
		)

	case messages.OperationFinishedEventMsg:
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
				utils.LogUpdate(m, msg.LogID, tuictx.StatusSuccess, tuictx.LogSuccess, msgText, "")
				return tea.Batch(
					utils.SetMsg(m, msgText),
					nav.Reload(m, false),
					tea.Tick(constants.ProgressDisplayDuration, func(time.Time) tea.Msg {
						return messages.ClearMsg{}
					}),
				)
			}
		}

	case tea.KeyMsg:
		// 1. Text Input Handling (Highest Priority)
		if m.UI.InputActive {
			// Special case: fuzzy search navigation
			isFuzzyNavKey := false
			if m.Inputs.Mode == tuictx.InputFuzzySearch {
				switch msg.String() {
				case "up", "down", "tab", "alt+j", "alt+k", "alt+n", "alt+m":
					isFuzzyNavKey = true
				}
			}

			if isFuzzyNavKey {
				if cmd := integration.HandleSearch(m, msg); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return tea.Batch(cmds...)
			}

			if m.Inputs.Mode != tuictx.InputFuzzySearch {
				switch msg.String() {
				case "tab":
					if m.Inputs.Mode == tuictx.InputGoto || m.Inputs.Mode == tuictx.InputAuth || m.Inputs.Mode == tuictx.InputCreate {
						m.Inputs.AltMode = !m.Inputs.AltMode
						return nil
					}
				}
			}

			var cmd tea.Cmd
			m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

			if m.Inputs.Mode == tuictx.InputFuzzySearch {
				// Trigger search on change
				if msg.String() != "enter" && msg.String() != "esc" {
					query := m.Inputs.ActiveInput.Value()
					if query != m.Search.Query {
						cmds = append(cmds, integration.TriggerSearch(m, query))
					}
				}
			}

			if m.Inputs.Mode == tuictx.InputSearch {
				if m.Navigation.FilterTimer != nil {
					m.Navigation.FilterTimer.Stop()
				}
				m.Navigation.FilterGen++
				gen := m.Navigation.FilterGen
				m.Navigation.FilterTimer = time.NewTimer(50 * time.Millisecond)
				cmds = append(cmds, func() tea.Msg {
					<-m.Navigation.FilterTimer.C
					return messages.DebounceFilterMsg{Generation: gen}
				})
			}

			// Handle Enter/Esc for inputs
			switch msg.String() {
			case "enter":
				if cmd := finalizeInput(m); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return tea.Batch(cmds...)
			case "esc":
				mode := m.Inputs.Mode
				m.StopInput(true)
				if mode == tuictx.InputSearch {
					m.Navigation.FilterQuery = ""
					nav.ApplyFilter(m)
				}
				if mode == tuictx.InputFuzzySearch {
					integration.StopSearch(m)
				}
				return tea.Batch(cmds...)
			}
			return tea.Batch(cmds...)
		}

		// Handle global keys (Quit, etc)
		switch msg.String() {
		case "ctrl+c":
			if m.Message.Text == "Press [Ctrl+C] again to close" {
				if m.FS.IsLocal() && m.Watcher.Watcher != nil {
					m.Watcher.Watcher.Close()
				}
				return tea.Quit
			}
			return utils.SetMsg(m, "Press [Ctrl+C] again to close")
		case "alt+l":
			m.UI.ToggleLogs()
			return tea.Batch(cmds...)
		case "alt+c":
			m.UI.ToggleClipboard()
			return tea.Batch(cmds...)
		case ".":
			m.UI.ToggleSettings()
			return tea.Batch(cmds...)
		case "esc":
			// 1. High Priority: Cancel active confirmation or prompt
			if m.UI.Confirming {
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				return tea.Batch(cmds...)
			}
			if m.UI.HostConfirm {
				m.UI.HostConfirm = false
				m.Remote.HostConfirmReq = nil
				return tea.Batch(cmds...)
			}

			// 2. Handle closing modals next
			if m.UI.SettingsOpen {
				m.UI.ToggleSettings()
				return tea.Batch(cmds...)
			}
			if m.UI.LogOpen {
				m.UI.ToggleLogs()
				return tea.Batch(cmds...)
			}
			if m.UI.ClipboardOpen {
				m.UI.ToggleClipboard()
				return tea.Batch(cmds...)
			}

			// 3. Global esc handling for clearing selection or filter
			if !m.UI.InputActive && !m.UI.Confirming &&
				!m.UI.RemoteAuth && !m.UI.HostConfirm {
				if m.Navigation.FilterQuery != "" {
					m.Navigation.FilterQuery = ""
					nav.ApplyFilter(m)
					return tea.Batch(cmds...)
				}
				m.ClearSelection()
				return tea.Batch(cmds...)
			}
		}
	}

	if cmd := app.HandleUpdateMessages(m, msg); cmd != nil {
		return cmd
	}

	// 2. Delegate to domain handlers based on message type or UI state
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

func finalizeInput(m *tuictx.Model) tea.Cmd {
	val := m.Inputs.ActiveInput.Value()
	mode := m.Inputs.Mode

	switch mode {
	case tuictx.InputSearch:
		m.StopInput(false)
		return nil
	case tuictx.InputRename:
		m.StopInput(true)
		return file.PerformRename(m, val)
	case tuictx.InputConflictRename:
		m.StopInput(true)
		return file.PerformConflictRename(m, val)
	case tuictx.InputCreate:
		m.StopInput(true)
		return file.PerformCreate(m, val)
	case tuictx.InputZip:
		m.StopInput(true)
		return file.PerformZip(m, val)
	case tuictx.InputUnzip:
		m.StopInput(true)
		return file.PerformUnzip(m, val)
	case tuictx.InputGoto:
		m.StopInput(true)
		return nav.HandleGotoFinalize(m, val)
	case tuictx.InputAuth:
		m.StopInput(true)
		return integration.HandleAuthFinalize(m, val)
	case tuictx.InputFuzzySearch:
		if len(m.Search.Results) > 0 {
			res := m.Search.Results[m.Search.CursorFile]
			line := 1
			if m.Search.CursorMatch >= 0 && m.Search.CursorMatch < len(res.Matches) {
				line = res.Matches[m.Search.CursorMatch].Line
			}

			m.StopInput(true)
			integration.StopSearch(m)

			return OpenFileAtLineAction(m, res.Path, line)
		}
		m.StopInput(true)
		integration.StopSearch(m)
	}
	return nil
}
