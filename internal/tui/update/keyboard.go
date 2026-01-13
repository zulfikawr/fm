package update

import (
	"fm/internal/constants"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleKeyMsg delegates keyboard events to specialized handlers
func HandleKeyMsg(m *state.Model, msg tea.KeyMsg) tea.Cmd {
	// Handle escape key for clearing selections and inputs
	if handled, cmd := HandleEscape(m, msg); handled {
		return cmd
	}

	// Mode-specific updates
	if m.UI.SettingsOpen {
		return HandleSettingsUpdate(msg, m)
	}
	if m.UI.InputActive {
		switch m.Inputs.Mode {
		case state.InputSearch:
			return HandleSearching(msg, m)
		case state.InputRename:
			return HandleRenaming(msg, m)
		case state.InputGoto:
			return HandleGoto(msg, m)
		case state.InputAuth:
			return HandleRemoteAuth(msg, m)
		}
	}

	if m.UI.HostConfirm {
		switch msg.String() {
		case "y", "Y":
			m.UI.HostConfirm = false
			if m.Remote.HostConfirmReq != nil && m.Remote.HostConfirmReq.Resolve != nil {
				m.Remote.HostConfirmReq.Resolve <- true
			}
			m.UI.Loading = true
			return commands.ListenForHostConfirmation(m.Remote.HostConfirmChan)
		case "n", "N", "esc":
			m.UI.HostConfirm = false
			if m.Remote.HostConfirmReq != nil && m.Remote.HostConfirmReq.Resolve != nil {
				m.Remote.HostConfirmReq.Resolve <- false
			}
			return nil
		}
		return nil
	}

	if m.UI.Confirming {
		return HandleConfirming(msg, m)
	}

	// Main navigation and actions
	switch msg.String() {
	case "alt+t":
		cmd, handled := HandleCreateTab(m)
		if handled {
			return cmd
		}
	case "alt+1", "alt+2", "alt+3", "alt+4", "alt+5", "alt+6", "alt+7", "alt+8", "alt+9":
		tabNum := ParseTabNumber(msg.String())
		cmd, handled := HandleSwitchTab(m, tabNum)
		if handled {
			return cmd
		}
	case "alt+w":
		cmd, handled := HandleCloseTab(m)
		if handled {
			return cmd
		}
	case "ctrl+c":
		if m.Operations.Progress.Visible {
			m.UI.Confirming = true
			m.Operations.ActionType = constants.ActionCancel
			return nil
		}
		if m.FS.IsLocal() && m.Watcher.Watcher != nil {
			m.Watcher.Watcher.Close()
		}
		return tea.Quit
	case "q":
		if m.FS.IsLocal() && m.Watcher.Watcher != nil {
			m.Watcher.Watcher.Close()
		}
		return tea.Quit
	}

	var cmds []tea.Cmd
	cmds = append(cmds, HandleNavigation(msg, m)...)
	cmds = append(cmds, HandleAction(msg, m)...)

	return tea.Batch(cmds...)
}

// HandleEscape handles the escape key for clearing selections and messages
func HandleEscape(m *state.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
	if msg.String() != "esc" {
		return false, nil
	}

	// Only handle escape if not in any special mode
	if m.UI.InputActive || m.UI.Confirming || m.UI.SettingsOpen ||
		m.UI.RemoteAuth || m.UI.HostConfirm {
		return false, nil
	}

	actions.ClearSelection(m)
	return true, nil
}

// ParseTabNumber extracts the tab number from alt+N key combinations
func ParseTabNumber(key string) int {
	if len(key) >= 5 && key[:4] == "alt+" {
		return int(key[4] - '0') // Extract number from "alt+N"
	}
	return 0
}
