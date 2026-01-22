package factory

import (
	"fmt"
	"syscall"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/files/local"
	remotefs "github.com/zulfikawr/fm/internal/files/remote"
	"github.com/zulfikawr/fm/internal/ssh"

	sshx "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// RemoteInfo contains information about a remote connection
type RemoteInfo struct {
	Host      string
	User      string
	StartPath string
}

// FileSystemConnector defines the interface for creating file systems.
type FileSystemConnector interface {
	NewLocalFS() core.FileSystem
	NewRemoteFS(opts ssh.SSHConfig) (core.FileSystem, error)
	ReadPassword() (string, error)
	CreateHostKeyCallback() (sshx.HostKeyCallback, error)
}

// DefaultConnector is the production implementation of FileSystemConnector.
type DefaultConnector struct{}

func (c *DefaultConnector) NewLocalFS() core.FileSystem {
	return local.NewLocalFS()
}

func (c *DefaultConnector) NewRemoteFS(opts ssh.SSHConfig) (core.FileSystem, error) {
	return remotefs.NewRemoteFS(opts)
}

func (c *DefaultConnector) ReadPassword() (string, error) {
	fmt.Print("Password: ")
	bytePw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", errors.WrapError(err, "read password")
	}
	return string(bytePw), nil
}

func (c *DefaultConnector) CreateHostKeyCallback() (sshx.HostKeyCallback, error) {
	cb, err := ssh.CreateCLIHostKeyCallback()
	return cb, errors.WrapError(err, "create host key callback")
}
