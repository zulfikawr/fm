package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/constants"
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

	case tea.WindowSizeMsg:
		m.Display.Width = msg.Width
		m.Display.Height = msg.Height
		m.SyncViewportHeight()
		return nil, true

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.Message.Text == "Press [Ctrl+C] again to close" {
				if m.FS.IsLocal() && m.Watcher.Watcher != nil {
					_ = m.Watcher.Watcher.Close()
				}
				return tea.Quit, true
			}
			return utils.SetMsg(m, "Press [Ctrl+C] again to close"), true
		case "alt+u":
			if !m.UI.InputActive && !m.UI.SettingsOpen && !m.UI.HelpOpen {
				if m.UI.AnalyzeOpen {
					m.UI.AnalyzeOpen = false
					return nil, true
				}
				return func() tea.Msg { return messages.StartAnalyzeMsg{} }, true
			}
			return nil, false
		case "alt+l":
			m.UI.ToggleLogs()
			return nil, true
		case "alt+c":
			m.UI.ToggleClipboard()
			return nil, true
		case ".":
			if m.UI.InputActive {
				return nil, false
			}
			m.UI.ToggleSettings()
			return nil, true
		case "?":
			if m.UI.InputActive {
				return nil, false
			}
			m.UI.ToggleHelp()
			return nil, true
		case "esc":
			// 1. High Priority: Cancel active confirmation or prompt
			if m.UI.Confirming {
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				return nil, true
			}
			if m.UI.HostConfirm {
				m.UI.HostConfirm = false
				m.Remote.HostConfirmReq = nil
				return nil, true
			}

			// 2. Handle closing modals next
			if m.UI.SettingsOpen {
				m.UI.ToggleSettings()
				return nil, true
			}
			if m.UI.HelpOpen {
				m.UI.ToggleHelp()
				return nil, true
			}
			if m.UI.LogOpen {
				m.UI.ToggleLogs()
				return nil, true
			}
			if m.UI.ClipboardOpen {
				m.UI.ToggleClipboard()
				return nil, true
			}
			if m.UI.AnalyzeOpen {
				m.UI.AnalyzeOpen = false
				return nil, true
			}
		}
	}
	return nil, false
}
