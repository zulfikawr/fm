package integration

import (
	"fmt"
	"os"
	"strings"

	"github.com/zulfikawr/fm/internal/files/remote"
	"github.com/zulfikawr/fm/internal/ssh"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleRemote handles remote-related messages
func HandleRemote(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.HostConfirm {
			return handleHostConfirmKeys(m, msg)
		}
	case messages.RemoteConnectMsg:
		return finalizeRemoteConnect(m, msg)
	case messages.HostConfirmMsg:
		return handleHostConfirm(m, msg)
	}
	return nil
}

func ConnectRemote(opts ssh.SSHConfig, askChan chan *ssh.HostConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		hkcb, err := ssh.GetHostKeyCallback(askChan)
		if err != nil {
			return messages.RemoteConnectMsg{Err: err}
		}

		// Create a new config with the callback instead of mutating
		newOpts := opts
		newOpts.HostKeyCallback = hkcb

		fs, err := remote.NewRemoteFS(newOpts)
		if err != nil {
			return messages.RemoteConnectMsg{Err: err}
		}

		home, _ := fs.GetHomeDir()
		return messages.RemoteConnectMsg{FS: fs, Path: home}
	}
}

func ListenForHostConfirmation(askChan chan *ssh.HostConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		req, ok := <-askChan
		if !ok {
			return nil
		}
		return messages.HostConfirmMsg{Request: req}
	}
}

func finalizeRemoteConnect(m *tui_context.Model, msg messages.RemoteConnectMsg) tea.Cmd {
	m.UI.Loading = false
	if msg.Err != nil {
		errStr := msg.Err.Error()

		utils.LogPush(m, tui_context.LogEntry{
			Type:    "Remote",
			Level:   tui_context.LogError,
			Status:  tui_context.StatusError,
			Message: "Connection failed",
			Details: errStr,
		})

		// Check if it's a fatal network/host error - don't offer password fallback
		if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") || strings.Contains(errStr, "i/o timeout") {
			return utils.LogError(m, msg.Err, "Remote connection failed")
		}

		// If we were trying key auth and it failed, show specific error
		if m.Navigation.Remote.TryKeyAuth {
			m.Navigation.Remote.TryKeyAuth = false
			if strings.Contains(errStr, "failed to read key file") || strings.Contains(errStr, "failed to parse key file") {
				return utils.LogError(m, msg.Err, "Key authentication failed")
			}
			if strings.Contains(errStr, "ssh: handshake failed") || strings.Contains(errStr, "unable to authenticate") {
				// Key auth failed - user can try password manually
				errMsg := "Key auth failed. Try: 'g' -> 'r' -> [p] for password"
				utils.LogPush(m, tui_context.LogEntry{
					Type:    "Remote",
					Level:   tui_context.LogError,
					Status:  tui_context.StatusError,
					Message: "Key authentication failed",
					Details: errMsg,
				})
				return utils.LogError(m, msg.Err, "Key authentication failed")
			}
		}

		// Only offer password auth if we weren't just trying key auth
		if !m.Navigation.Remote.TryKeyAuth {
			m.UI.RemoteAuth = true
			m.Operations.ActionType = "auth"
			m.UI.StartConfirming()

			return utils.SetMsg(m, "Remote Authentication Required")
		}

		return utils.LogError(m, msg.Err, "Remote connection failed")
	}

	m.UI.RemoteAuth = false
	m.FS = msg.FS
	m.Navigation.Remote.TryKeyAuth = false

	authMethod := "password/key"
	if m.Navigation.Remote.KeyPath != "" {
		authMethod = fmt.Sprintf("PEM: %s", m.Navigation.Remote.KeyPath)
	} else if m.Inputs.ActiveInput.Value() == "" {
		authMethod = "agent/default keys"
	}

	utils.LogPush(m, tui_context.LogEntry{
		Type:    "Remote",
		Level:   tui_context.LogSuccess,
		Status:  tui_context.StatusSuccess,
		Message: fmt.Sprintf("Connection established to %s@%s", m.Navigation.Remote.User, m.Navigation.Remote.Host),
		Details: fmt.Sprintf("Authenticated via %s", authMethod),
	})

	return func() tea.Msg { return messages.NavigateMsg{Path: msg.Path} }
}

func handleHostConfirm(m *tui_context.Model, msg messages.HostConfirmMsg) tea.Cmd {
	m.UI.Loading = false
	m.UI.HostConfirm = true
	m.Navigation.Remote.HostConfirmReq = msg.Request
	return nil
}

func handleHostConfirmKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		if m.Navigation.Remote.HostConfirmReq != nil && m.Navigation.Remote.HostConfirmReq.Resolve != nil {
			resolve := m.Navigation.Remote.HostConfirmReq.Resolve
			go func() {
				resolve <- true
			}()
		}
		m.UI.HostConfirm = false
		m.Navigation.Remote.HostConfirmReq = nil
		m.UI.Loading = true
		return nil
	case "n", "N":
		if m.Navigation.Remote.HostConfirmReq != nil && m.Navigation.Remote.HostConfirmReq.Resolve != nil {
			resolve := m.Navigation.Remote.HostConfirmReq.Resolve
			go func() {
				resolve <- false
			}()
		}
		m.UI.HostConfirm = false
		m.Navigation.Remote.HostConfirmReq = nil
		return utils.SetMsg(m, "Connection cancelled (host untrusted)")
	}
	return nil
}

func HandleRemoteGoto(m *tui_context.Model, input string) tea.Cmd {
	host := input
	user := ""
	keyPath := ""

	// 1. Resolve alias from ~/.ssh/config
	sshConfigs, _ := ssh.ParseSSHConfig()
	if cfg, ok := sshConfigs[input]; ok {
		host = cfg.HostName
		if host == "" {
			host = input
		}
		user = cfg.User
		keyPath = cfg.IdentityFile
	} else if strings.Contains(input, "@") {
		// 2. Parse user@host
		parts := strings.SplitN(input, "@", 2)
		user = parts[0]
		host = parts[1]
	}

	m.Navigation.Remote.Host = host
	m.Navigation.Remote.User = user
	m.Navigation.Remote.KeyPath = ""
	m.Navigation.Remote.TryKeyAuth = false
	m.UI.Loading = true
	m.UI.RemoteAuth = false
	m.Inputs.AltMode = false

	return tea.Batch(
		ConnectRemote(ssh.SSHConfig{
			Address: host,
			User:    user,
			KeyPath: keyPath,
		}, m.Navigation.Remote.HostConfirmChan),
		ListenForHostConfirmation(m.Navigation.Remote.HostConfirmChan),
	)
}

func HandleAuthFinalize(m *tui_context.Model, input string) tea.Cmd {
	password := ""
	keyPath := ""
	tryKeyAuth := false

	if m.Inputs.AltMode {
		keyPath = input
		tryKeyAuth = true

		// Check if key file exists first
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			errMsg := fmt.Sprintf("Key file not found: %s", keyPath)
			utils.LogPush(m, tui_context.LogEntry{
				Type:    "Remote",
				Level:   tui_context.LogError,
				Status:  tui_context.StatusError,
				Message: "Key authentication failed",
				Details: errMsg,
			})
			utils.LogError(m, fmt.Errorf("%s", errMsg), "Key file not found")
			return nil
		}
	} else {
		password = input
	}

	m.UI.Loading = true
	m.Navigation.Remote.KeyPath = keyPath
	m.Navigation.Remote.TryKeyAuth = tryKeyAuth

	return tea.Batch(
		ConnectRemote(ssh.SSHConfig{
			Address:  m.Navigation.Remote.Host,
			User:     m.Navigation.Remote.User,
			Password: password,
			KeyPath:  keyPath,
		}, m.Navigation.Remote.HostConfirmChan),
		ListenForHostConfirmation(m.Navigation.Remote.HostConfirmChan),
	)
}
