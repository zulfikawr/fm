package handlers

import (
	"fmt"
	"strings"

	"fm/internal/tui/components/ui"
	tui_context "fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleRemote handles remote-related messages
func HandleRemote(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.HostConfirm {
			return handleHostConfirmKeys(m, msg)
		}
	case RemoteConnectMsg:
		return finalizeRemoteConnect(m, msg)
	case HostConfirmMsg:
		return handleHostConfirm(m, msg)
	}
	return nil
}

func finalizeRemoteConnect(m *tui_context.Model, msg RemoteConnectMsg) tea.Cmd {
	m.UI.Loading = false
	if msg.Err != nil {
		errStr := msg.Err.Error()

		// Log the error for visibility in Alt+L
		LogPush(m, "Remote", tui_context.LogError, tui_context.StatusError, "Connection failed", errStr)

		// If it's a known error that doesn't require authentication (like host down), log it
		if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") || strings.Contains(errStr, "i/o timeout") {
			return LogError(m, msg.Err, "Remote connection failed")
		}

		// Otherwise, assume authentication is needed or was wrong
		m.UI.RemoteAuth = true
		m.UI.StartInput()
		m.Inputs.Mode = tui_context.InputAuth
		m.Inputs.ActiveInput.SetValue("")

		// Set prompt label based on AltMode and enable masking for password
		label := "Password"
		if m.Inputs.AltMode {
			label = "PEM Path"
			m.Inputs.ActiveInput.EchoMode = ui.EchoNormal
		} else {
			m.Inputs.ActiveInput.EchoMode = ui.EchoPassword
		}
		m.Inputs.ActiveInput.SetPrompt(label + ": ")

		return tea.Batch(
			SetMsg(m, "Remote Authentication Required"),
			m.Inputs.ActiveInput.FocusCmd(),
		)
	}

	// Success!
	m.UI.RemoteAuth = false
	m.FS = msg.FS

	authMethod := "password/key"
	if m.Inputs.AltMode {
		authMethod = "PEM file"
	} else if m.Inputs.ActiveInput.Value() == "" {
		authMethod = "agent/default keys"
	}

	LogPush(m, "Remote", tui_context.LogSuccess, tui_context.StatusSuccess,
		fmt.Sprintf("Connection established to %s@%s", m.Remote.User, m.Remote.Host),
		fmt.Sprintf("Authenticated via %s", authMethod))

	return NavigateToPath(m, msg.Path)
}

func handleHostConfirm(m *tui_context.Model, msg HostConfirmMsg) tea.Cmd {
	m.UI.Loading = false
	m.UI.HostConfirm = true
	m.Remote.HostConfirmReq = msg.Request
	return nil
}

func handleHostConfirmKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		if m.Remote.HostConfirmReq != nil && m.Remote.HostConfirmReq.Resolve != nil {
			// Signal background goroutine to proceed
			resolve := m.Remote.HostConfirmReq.Resolve
			go func() {
				resolve <- true
			}()
		}
		m.UI.HostConfirm = false
		m.Remote.HostConfirmReq = nil
		m.UI.Loading = true
		return nil
	case "n", "N":
		if m.Remote.HostConfirmReq != nil && m.Remote.HostConfirmReq.Resolve != nil {
			// Signal background goroutine to fail
			resolve := m.Remote.HostConfirmReq.Resolve
			go func() {
				resolve <- false
			}()
		}
		m.UI.HostConfirm = false
		m.Remote.HostConfirmReq = nil
		return SetMsg(m, "Connection cancelled (host untrusted)")
	}
	return nil
}
