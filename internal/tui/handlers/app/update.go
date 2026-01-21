package app

import (
	"github.com/zulfikawr/fm/internal/constants"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleUpdateMessages handles update-related messages
func HandleUpdateMessages(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case messages.UpdateAvailableMsg:
		m.UI.UpdateAvailable = true
		m.UI.LatestVersion = msg.Version
		m.UI.StartConfirming()
		m.Operations.ActionType = constants.ActionUpdate
		return nil

	case messages.UpdateProgressMsg:
		m.Operations.Progress.Percent = float64(msg)
		m.Operations.Progress.Show("Updating...")
		return nil

	case messages.UpdateFinishedMsg:
		m.Operations.Progress.Hide()
		if msg.Err != nil {
			return utils.SetErrMsg(m, "Update failed: "+msg.Err.Error())
		}
		m.UI.UpdateAvailable = false
		return utils.SetMsg(m, "Successfully updated. Press [r] to restart")

	case tea.KeyMsg:
		if m.UI.UpdateAvailable && m.Operations.ActionType == constants.ActionUpdate {
			switch msg.String() {
			case "y", "Y":
				m.UI.StopConfirming()
				return StartUpdate(m)
			case "n", "N":
				m.UI.StopConfirming()
				m.UI.UpdateAvailable = false
				m.Operations.ActionType = constants.ActionNone
				return func() tea.Msg { return nil }
			case "esc":
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				return func() tea.Msg { return nil }
			}
		}

		// Handle restart after update
		if msg.String() == "r" && m.Message.Text == "Successfully updated. Press [r] to restart" {
			return RestartApp()
		}
	}

	return nil
}
