package handlers

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/trash"
	"github.com/zulfikawr/fm/internal/logger"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

// handleGlobal handles system-level messages like resizing, ticks, and global shortcuts
func handleGlobal(m *tuictx.Model, msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case app.TickMsg:
		return nil, true

	case app.RAMUpdateMsg:
		m.Display.RAMUsageMB = msg.RAMUsageMB
		return app.StartRAMTicker(), true // Reschedule the RAM ticker

	case tea.WindowSizeMsg:
		m.Display.Width = msg.Width
		m.Display.Height = msg.Height
		m.SyncViewportHeight()
		return nil, true

	case tea.KeyMsg:
		// Handle keybindings using custom configuration
		action := config.GetActionForKey(msg.String(), m.Config.Keybindings)

		switch action {
		case "quit":
			// If we are recording a keybinding, don't trigger quit
			if m.UI.InputActive && m.Inputs.Mode == tuictx.InputKeybinding {
				return nil, false
			}

			// Get current quit key for dynamic message
			quitKey := "ctrl+c"
			for i := range m.Config.Keybindings {
				kb := m.Config.Keybindings[i]
				if kb.Action == "quit" && len(kb.Keys) > 0 {
					quitKey = kb.Keys[0]
					if quitKey == " " {
						quitKey = "space"
					}
					break
				}
			}
			msg := "Press [" + quitKey + "] again to close"

			if m.Message.Text == msg {
				if m.FS.IsLocal() && m.Watcher.Watcher != nil {
					logger.CloseAndLog(m.Watcher.Watcher, "local filesystem watcher during quit")
				}
				return tea.Quit, true
			}
			// Set exit message and clear it after 2 seconds if not confirmed
			m.Message.Push(msg, false)
			return func() tea.Msg {
				time.Sleep(2 * time.Second)
				return messages.ClearStatusMsg{}
			}, true
		case "analyze":
			if !m.UI.InputActive && m.UI.ActiveView == tuictx.ViewMain {
				return func() tea.Msg { return messages.StartAnalyzeMsg{} }, true
			} else if m.UI.ActiveView == tuictx.ViewAnalyze {
				m.UI.ActiveView = tuictx.ViewMain
				return nil, true
			}
			return nil, false
		case "toggle_regex_search":
			if !m.UI.InputActive && m.UI.ActiveView == tuictx.ViewMain {
				m.Config.Ops.EnableRegexSearch = !m.Config.Ops.EnableRegexSearch
				if err := m.Config.Save(); err != nil {
					m.HandleError("Failed to save configuration", err)
				}
				msg := "Regex Search enabled"
				if !m.Config.Ops.EnableRegexSearch {
					msg = "Regex Search disabled"
				}
				return utils.SetMsg(m, msg), true
			}
			return nil, false
		case "logs_view":
			m.UI.ToggleLogs()
			return nil, true
		case "clipboard_view":
			m.UI.ToggleClipboard()
			return nil, true
		case "trash_view":
			m.UI.ToggleTrash()
			if m.UI.ActiveView == tuictx.ViewTrash {
				// Load trash items when opening
				return func() tea.Msg {
					return loadTrashItems(m)
				}, true
			}
			return nil, true
		case "settings":
			if m.UI.InputActive || m.UI.ActiveView == tuictx.ViewAnalyze {
				return nil, false
			}
			m.UI.ToggleSettings()
			return nil, true
		case "help":
			if m.UI.InputActive || m.UI.ActiveView == tuictx.ViewAnalyze {
				return nil, false
			}
			m.UI.ToggleHelp()
			return nil, true
		case "clear_selection", "cancel_input":
			// Handle escape key for multiple purposes
			// 1. High Priority: Cancel active input/prompt
			if m.UI.InputActive {
				// inputs.go already handles 'esc' for actual inputs,
				// but this ensures we don't fall through to closing screens
				return nil, false
			}
			if m.UI.Confirming {
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				return nil, true
			}
			if m.UI.HostConfirm {
				m.UI.HostConfirm = false
				m.Navigation.Remote.HostConfirmReq = nil
				return nil, true
			}

			// 2. Handle closing modals next
			if m.UI.ActiveView != tuictx.ViewMain {
				m.UI.ActiveView = tuictx.ViewMain
				return nil, true
			}
		}
	}
	return nil, false
}

func loadTrashItems(m *tuictx.Model) tea.Msg {
	manager, err := trash.NewManager(m.FS)
	if err != nil {
		return messages.ErrorMsg{Err: err}
	}

	items, err := manager.List()
	if err != nil {
		return messages.ErrorMsg{Err: err}
	}

	// Convert to interface{} slice
	result := make([]interface{}, len(items))
	for i, item := range items {
		result[i] = item
	}

	return messages.TrashLoadedMsg{Items: result}
}
