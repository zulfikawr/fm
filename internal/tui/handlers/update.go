package handlers

import (
	"os"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/tui/context"
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
		return UpdateAvailableMsg{Version: version}
	}
}

// StartUpdate starts the update process
func StartUpdate(m *context.Model) tea.Cmd {
	progress := make(chan float64)

	return tea.Batch(
		func() tea.Msg {
			defer close(progress)
			err := update.DownloadAndInstall(m.UI.LatestVersion, progress)
			return UpdateFinishedMsg{Err: err}
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
			func() tea.Msg { return UpdateProgressMsg(p) },
			listenForUpdateProgress(progress),
		)()
	}
}

// RestartApp restarts the application
func RestartApp() tea.Cmd {
	return func() tea.Msg {
		_ = update.Restart()
		os.Exit(0)
		return nil
	}
}
