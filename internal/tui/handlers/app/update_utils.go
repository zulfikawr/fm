package app

import (
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"
	"github.com/zulfikawr/fm/internal/update"

	tea "github.com/charmbracelet/bubbletea"
)

// CheckForUpdates triggers an update check
func CheckForUpdates() tea.Cmd {
	return func() tea.Msg {
		if strings.Contains(constants.AppVersion, "-dev") {
			return nil
		}
		version, err := update.CheckForUpdate()
		if err != nil || version == "" {
			return nil
		}
		return messages.UpdateAvailableMsg{Version: version}
	}
}

// StartUpdate starts the update process
func StartUpdate(m *tui_context.Model) tea.Cmd {
	progress := make(chan float64)

	return tea.Batch(
		func() tea.Msg {
			defer close(progress)
			err := update.DownloadAndInstall(m.UI.LatestVersion, progress)
			return messages.UpdateFinishedMsg{Err: err}
		},
		listenForUpdateProgress(progress),
	)
}

func listenForUpdateProgress(progress chan float64) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-progress
		if !ok {
			return nil
		}
		return tea.Batch(
			func() tea.Msg { return messages.UpdateProgressMsg(p) },
			listenForUpdateProgress(progress),
		)()
	}
}

