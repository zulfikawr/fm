package factory

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"fm/internal/files/core"
	"fm/internal/files/local"
	remotefs "fm/internal/files/remote"
	"fm/internal/sshutil"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
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
	NewSftpFS(host, user, password, keyPath string, hkcb ssh.HostKeyCallback) (core.FileSystem, error)
	ReadPassword() (string, error)
	CreateHostKeyCallback() (ssh.HostKeyCallback, error)
}

// DefaultConnector is the production implementation of FileSystemConnector.
type DefaultConnector struct{}

func (c *DefaultConnector) NewLocalFS() core.FileSystem {
	return local.NewLocalFS()
}

func (c *DefaultConnector) NewSftpFS(host, user, password, keyPath string, hkcb ssh.HostKeyCallback) (core.FileSystem, error) {
	return remotefs.NewSftpFS(host, user, password, keyPath, hkcb)
}

func (c *DefaultConnector) ReadPassword() (string, error) {
	fmt.Print("Password: ")
	bytePw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytePw), nil
}

func (c *DefaultConnector) CreateHostKeyCallback() (ssh.HostKeyCallback, error) {
	return createHostKeyCallback()
}

// CreateFileSystem instantiates a LocalFS or SftpFS based on the remote string.
func CreateFileSystem(remoteStr string, args []string) (core.FileSystem, *RemoteInfo, error) {
	return CreateFileSystemWithConnector(remoteStr, args, &DefaultConnector{})
}

// CreateFileSystemWithConnector allows injecting a custom connector for testing.
func CreateFileSystemWithConnector(remoteStr string, args []string, conn FileSystemConnector) (core.FileSystem, *RemoteInfo, error) {
	if remoteStr == "" {
		return conn.NewLocalFS(), nil, nil
	}

	host := remoteStr
	user := ""
	keyPath := ""
	startPath := ""

	// Check SSH config first
	sshConfigs, _ := sshutil.ParseSSHConfig()
	if cfg, ok := sshConfigs[remoteStr]; ok {
		host = cfg.HostName
		if host == "" {
			host = remoteStr
		}
		user = cfg.User
		keyPath = cfg.IdentityFile
	} else if strings.Contains(remoteStr, "@") {
		parts := strings.SplitN(remoteStr, "@", 2)
		user = parts[0]
		rest := parts[1]

		if strings.Contains(rest, ":") {
			parts2 := strings.SplitN(rest, ":", 2)
			host = parts2[0]
			startPath = parts2[1]
		} else {
			host = rest
		}
	}

	// Check for key file as positional argument
	if len(args) > 0 {
		keyPath = args[0]
	}

	fmt.Printf("Connecting to %s@%s...\n", user, host)

	// Create CLI host key callback (blocking)
	hkcb, err := conn.CreateHostKeyCallback()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup host key verification: %w", err)
	}

	// Try connecting with provided key or agent first
	fs, err := conn.NewSftpFS(host, user, "", keyPath, hkcb)
	if err != nil {
		// Check if it's a host key verification failure (user said no or mismatch)
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			return nil, nil, fmt.Errorf("host key verification failed")
		}

		// If failed, prompt for password
		fmt.Printf("Connection attempt failed: %v\n", err)
		password, err := conn.ReadPassword()
		if err != nil {
			return nil, nil, fmt.Errorf("reading password: %w", err)
		}

		fs, err = conn.NewSftpFS(host, user, password, keyPath, hkcb)
		if err != nil {
			return nil, nil, fmt.Errorf("connection failed: %w", err)
		}
	}

	if startPath == "." || startPath == "" {
		startPath, _ = fs.GetHomeDir()
	}

	return fs, &RemoteInfo{
			Host:      host,
			User:      user,
			StartPath: startPath,
		},
		nil
}

func createHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Ensure directory exists
	_ = os.MkdirAll(sshDir, 0o700)
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		_ = os.WriteFile(knownHostsPath, []byte{}, 0o600)
	}

	cb, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			err := cb(hostname, remote, key)
			if err == nil {
				return nil
			}

			var keyErr *knownhosts.KeyError
			if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
				// Host not found in known_hosts
				fmt.Printf("The authenticity of host '%s' can't be established.\n", hostname)
				fmt.Printf("%s key fingerprint is %s.\n", key.Type(), ssh.FingerprintSHA256(key))
				fmt.Print("Are you sure you want to continue connecting (y/n)? ")

				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
					// Add to known_hosts
					return sshutil.AddToKnownHosts(hostname, remote, key)
				}
			}
			return err
		},
		nil
}
