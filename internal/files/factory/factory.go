package factory

import (
	"errors"
	"fmt"
	"time"

	"github.com/zulfikawr/fm/internal/files/core"
	fileerrors "github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/logger"
	"github.com/zulfikawr/fm/internal/ssh"

	"golang.org/x/crypto/ssh/knownhosts"
)

// CreateFileSystem instantiates a LocalFS or RemoteFS based on the remote string.
func CreateFileSystem(remoteStr string, args []string) (core.FileSystem, *RemoteInfo, error) {
	fs, info, err := CreateFileSystemWithConnector(remoteStr, args, &DefaultConnector{})
	if err != nil {
		return nil, nil, err
	}

	// Wrap with decorators
	// Order: ContextFS (outer) -> CachedFS -> ErrorWrappedFS (inner)
	// ErrorWrappedFS should be inner so it wraps the raw errors from implementation.
	// CachedFS should be in middle to cache results.
	// ContextFS should be outer to check cancellation before anything else.
	fs = core.NewErrorWrappedFS(fs)
	fs = core.NewCachedFS(fs, 100, 2*time.Second)
	fs = core.NewContextFS(fs)

	return fs, info, nil
}

// CreateFileSystemWithConnector allows injecting a custom connector for testing.
func CreateFileSystemWithConnector(remoteStr string, args []string, conn FileSystemConnector) (core.FileSystem, *RemoteInfo, error) {
	if remoteStr == "" {
		return conn.NewLocalFS(), nil, nil
	}

	details := ssh.ResolveRemote(remoteStr)
	host := details.Host
	user := details.User
	keyPath := details.KeyPath
	startPath := details.StartPath

	// Check for key file as positional argument
	if len(args) > 0 {
		keyPath = args[0]
	}

	logger.Infof("Connecting to %s@%s...", user, host)

	// Create CLI host key callback (blocking)
	hkcb, err := conn.CreateHostKeyCallback()
	if err != nil {
		return nil, nil, fileerrors.WrapError(err, "setup host key verification")
	}

	// Try connecting with provided key or agent first
	fs, err := conn.NewRemoteFS(ssh.SSHConfig{
		Address:         host,
		User:            user,
		KeyPath:         keyPath,
		HostKeyCallback: hkcb,
	})
	if err != nil {
		// Check if it's a host key verification failure (user said no or mismatch)
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			return nil, nil, &fileerrors.PermissionError{Operation: "SSH", Path: host}
		}

		// If failed, prompt for password
		fmt.Printf("Connection attempt failed: %v\n", err)
		password, err := conn.ReadPassword()
		if err != nil {
			return nil, nil, fileerrors.WrapError(err, "reading password")
		}

		fs, err = conn.NewRemoteFS(ssh.SSHConfig{
			Address:         host,
			User:            user,
			Password:        password,
			KeyPath:         keyPath,
			HostKeyCallback: hkcb,
		})
		if err != nil {
			return nil, nil, fileerrors.WrapError(err, "connection failed")
		}
	}

	if startPath == "." || startPath == "" {
		cwd, err := fs.GetHomeDir()
		if err != nil {
			startPath = "/"
		} else {
			startPath = cwd
		}
	}

	return fs, &RemoteInfo{
			Host:      host,
			User:      user,
			StartPath: startPath,
		},
		nil
}
