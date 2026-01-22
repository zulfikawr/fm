package integration

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/files/remote"
	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
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

		opts.HostKeyCallback = hkcb
		fs, err := remote.NewRemoteFS(opts)
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

		if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") || strings.Contains(errStr, "i/o timeout") {
			return utils.LogError(m, msg.Err, "Remote connection failed")
		}

		m.UI.RemoteAuth = true
		m.StartInput(tui_context.InputAuth)

		label := "Password"
		if m.Inputs.AltMode {
			label = "PEM Path"
			m.Inputs.ActiveInput.EchoMode = ui.EchoNormal
		} else {
			m.Inputs.ActiveInput.EchoMode = ui.EchoPassword
		}
		m.Inputs.ActiveInput.SetPrompt(label + ": ")

		return tea.Batch(
			utils.SetMsg(m, "Remote Authentication Required"),
			m.Inputs.ActiveInput.FocusCmd(),
		)
	}

	m.UI.RemoteAuth = false
	m.FS = msg.FS

	authMethod := "password/key"
	if m.Inputs.AltMode {
		authMethod = "PEM file"
	} else if m.Inputs.ActiveInput.Value() == "" {
		authMethod = "agent/default keys"
	}

	utils.LogPush(m, tui_context.LogEntry{
		Type:    "Remote",
		Level:   tui_context.LogSuccess,
		Status:  tui_context.StatusSuccess,
		Message: fmt.Sprintf("Connection established to %s@%s", m.Remote.User, m.Remote.Host),
		Details: fmt.Sprintf("Authenticated via %s", authMethod),
	})

	return func() tea.Msg { return messages.NavigateMsg{Path: msg.Path} }
}

func handleHostConfirm(m *tui_context.Model, msg messages.HostConfirmMsg) tea.Cmd {
	m.UI.Loading = false
	m.UI.HostConfirm = true
	m.Remote.HostConfirmReq = msg.Request
	return nil
}

func handleHostConfirmKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		if m.Remote.HostConfirmReq != nil && m.Remote.HostConfirmReq.Resolve != nil {
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
			resolve := m.Remote.HostConfirmReq.Resolve
			go func() {
				resolve <- false
			}()
		}
		m.UI.HostConfirm = false
		m.Remote.HostConfirmReq = nil
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

	m.Remote.Host = host
	m.Remote.User = user
	m.UI.Loading = true
	m.UI.RemoteAuth = false
	m.Inputs.AltMode = false

	return tea.Batch(
		ConnectRemote(ssh.SSHConfig{
			Address: host,
			User:    user,
			KeyPath: keyPath,
		}, m.Remote.HostConfirmChan),
		ListenForHostConfirmation(m.Remote.HostConfirmChan),
	)
}

func HandleAuthFinalize(m *tui_context.Model, input string) tea.Cmd {
	m.UI.Loading = true
	password := ""
	keyPath := ""

	if m.Inputs.AltMode {
		keyPath = input
	} else {
		password = input
	}

	return tea.Batch(
		ConnectRemote(ssh.SSHConfig{
			Address:  m.Remote.Host,
			User:     m.Remote.User,
			Password: password,
			KeyPath:  keyPath,
		}, m.Remote.HostConfirmChan),
		ListenForHostConfirmation(m.Remote.HostConfirmChan),
	)
}
