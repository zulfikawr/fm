package commands

import (
	"fm/internal/files/core"
	"fm/internal/files/remote"
	"fm/internal/sshutil"

	tea "github.com/charmbracelet/bubbletea"
)

// RemoteConnectMsg is sent when a remote connection is established or fails.
type RemoteConnectMsg struct {
	FS   core.FileSystem
	Path string
	Err  error
}

// HostConfirmMsg is sent when a host needs confirmation.
type HostConfirmMsg struct {
	Request *sshutil.HostConfirmRequest
}

// ConnectRemote creates a command to connect to a remote SFTP server.
func ConnectRemote(address, user, password, keyPath string, askChan chan *sshutil.HostConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		// Create a wrapper channel because NewSftpFS expects chan<- *HostConfirmRequest (eventually)
		// but we want to pass it through our utilities.

		hkcb, err := sshutil.GetHostKeyCallback(askChan)
		if err != nil {
			return RemoteConnectMsg{Err: err}
		}

		fs, err := remote.NewSftpFS(address, user, password, keyPath, hkcb)
		if err != nil {
			return RemoteConnectMsg{Err: err}
		}

		home, _ := fs.GetHomeDir()
		return RemoteConnectMsg{FS: fs, Path: home}
	}
}

// ListenForHostConfirmation listens for host confirmation requests on a channel.
func ListenForHostConfirmation(askChan chan *sshutil.HostConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		req, ok := <-askChan
		if !ok {
			return nil
		}
		return HostConfirmMsg{Request: req}
	}
}
