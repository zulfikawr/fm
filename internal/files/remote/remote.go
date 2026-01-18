package remote

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"

	"github.com/pkg/sftp"
	sshx "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// RemoteFS implements FileSystem for SFTP.
type RemoteFS struct {
	mu     sync.RWMutex
	client *sftp.Client
	conn   *sshx.Client
	cache  *core.MetadataCache

	// Connection details for reconnection
	address string
	user    string
	config  *sshx.ClientConfig

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRemoteFS creates a new SFTP file system.
func NewRemoteFS(address, user, password, keyPath string, hostKeyCallback sshx.HostKeyCallback) (*RemoteFS, error) {
	auths := []sshx.AuthMethod{}

	// 1. Try SSH Agent
	if sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		agentClient := agent.NewClient(sock)
		auths = append(auths, sshx.PublicKeysCallback(agentClient.Signers))
	}

	// 2. Try Identity File (Key)
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file %s: %w", keyPath, err)
		}
		signer, err := sshx.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key file %s: %w", keyPath, err)
		}
		auths = append(auths, sshx.PublicKeys(signer))
	}

	// 3. Try Password
	if password != "" {
		auths = append(auths, sshx.Password(password))
	}

	// Setup HostKeyCallback - REQUIRE known_hosts for security by default
	if hostKeyCallback == nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.WrapError(err, "NewRemoteFS")
		}

		sshDir := filepath.Join(home, ".ssh")
		knownHostsPath := filepath.Join(sshDir, "known_hosts")

		// Ensure .ssh directory exists
		if _, err := os.Stat(sshDir); os.IsNotExist(err) {
			if err := os.MkdirAll(sshDir, 0o700); err != nil {
				return nil, fmt.Errorf("failed to create .ssh directory: %w", err)
			}
		}

		// Create empty known_hosts if it doesn't exist to prevent knownhosts.New from failing
		if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
			if err := os.WriteFile(knownHostsPath, []byte{}, 0o600); err != nil {
				return nil, fmt.Errorf("failed to create empty known_hosts: %w", err)
			}
		}

		hostKeyCallback, err = knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("load known_hosts failed: %s: %w\nRun 'ssh-keyscan %s >> ~/.ssh/known_hosts' first", knownHostsPath, err, address)
		}
	}

	config := &sshx.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         constants.SSHConnectionTimeout,
	}

	// Ensure port is present
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = address + ":22"
	}

	conn, err := sshx.Dial("tcp", address, config)
	if err != nil {
		return nil, errors.WrapError(err, "ssh dial failed: "+address)
	}

	client, err := sftp.NewClient(conn,
		sftp.UseConcurrentWrites(true),
	)
	if err != nil {
		conn.Close()
		return nil, errors.WrapError(err, "create sftp client failed: "+address)
	}

	ctx, cancel := context.WithCancel(context.Background())
	fs := &RemoteFS{
		client:  client,
		conn:    conn,
		cache:   core.NewMetadataCache(2 * time.Second),
		address: address,
		user:    user,
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Start background keep-alive
	go fs.keepAlive()

	return fs, nil
}

// Close releases the SFTP client and the underlying SSH connection.
func (fs *RemoteFS) Close() error {
	if fs.cancel != nil {
		fs.cancel()
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	var errs []string
	if fs.client != nil {
		if err := fs.client.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("sftp client: %v", err))
		}
	}
	if fs.conn != nil {
		if err := fs.conn.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("ssh connection: %v", err))
		}
	}
	if len(errs) > 0 {
		return errors.WrapError(fmt.Errorf("close failed: %s", strings.Join(errs, ", ")), "Close")
	}
	return nil
}

func (fs *RemoteFS) GetHomeDir() (string, error) {
	dir, err := fs.client.Getwd()
	return dir, errors.WrapError(err, "GetHomeDir")
}

func (fs *RemoteFS) Separator() string {
	return "/"
}

func (fs *RemoteFS) IsLocal() bool {
	return false
}

func (fs *RemoteFS) Address() string {
	return fs.address
}

func (fs *RemoteFS) User() string {
	return fs.user
}
