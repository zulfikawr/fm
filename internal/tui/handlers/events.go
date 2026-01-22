package handlers

import (
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// handleEvents handles external signals and operation lifecycle events
func handleEvents(m *tuictx.Model, msg tea.Msg) (tea.Cmd, bool) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.WatchEventMsg:
		if m.Watcher.DebounceTimer != nil {
			m.Watcher.DebounceTimer.Stop()
		}
		m.Watcher.DebounceTimer = time.NewTimer(150 * time.Millisecond)
		return func() tea.Msg {
			<-m.Watcher.DebounceTimer.C
			return messages.DebounceWatchMsg{}
		}, true

	case messages.DebounceWatchMsg:
		m.Watcher.IsListening = false
		return nav.Reload(m, false), true

	case messages.WatcherErrorMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds,
			utils.SetErrMsg(m, "Watcher error: restarting..."),
			utils.RestartWatcherAction(m),
		)
		return tea.Batch(cmds...), true

	case messages.WatcherClosedMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds,
			utils.SetMsg(m, "Watcher closed: restarting..."),
			utils.RestartWatcherAction(m),
		)
		return tea.Batch(cmds...), true

	case messages.RemotePollMsg:
		m.Watcher.IsListening = false
		return nav.Reload(m, true), true

	case messages.DebounceFilterMsg:
		if msg.Generation == m.Navigation.FilterGen {
			nav.ApplyFilter(m)
		}
		return nil, true

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
				utils.LogUpdate(m, msg.LogID, tuictx.LogEntry{
					Status:  tuictx.StatusSuccess,
					Level:   tuictx.LogSuccess,
					Message: msgText,
				})
				return tea.Batch(
					utils.SetMsg(m, msgText),
					nav.Reload(m, false),
					tea.Tick(constants.ProgressDisplayDuration, func(time.Time) tea.Msg {
						return messages.ClearMsg{}
					}),
				), true
			}
		}

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
				utils.LogUpdate(m, msg.LogID, tuictx.LogEntry{
					Status:  tuictx.StatusError,
					Level:   tuictx.LogError,
					Message: msgText,
					Details: msg.Err.Error(),
				})
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
		return tea.Batch(cmds...), true
	}

	return nil, false
}
