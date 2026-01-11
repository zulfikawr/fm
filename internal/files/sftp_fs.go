package files

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConnectionTimeout is the timeout for SSH connections
const SSHConnectionTimeout = 5 * time.Second

// SftpFS implements FileSystem for SFTP.
type SftpFS struct {
	client *sftp.Client
	conn   *ssh.Client
}

// NewSftpFS creates a new SFTP file system.
func NewSftpFS(address, user, password, keyPath string) (*SftpFS, error) {
	auths := []ssh.AuthMethod{}

	// 1. Try SSH Agent
	if sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		agentClient := agent.NewClient(sock)
		auths = append(auths, ssh.PublicKeysCallback(agentClient.Signers))
	}

	// 2. Try Identity File (Key)
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				auths = append(auths, ssh.PublicKeys(signer))
			}
		}
	}

	// 3. Try Password
	if password != "" {
		auths = append(auths, ssh.Password(password))
	}

	// Setup HostKeyCallback - REQUIRE known_hosts for security
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory failed: %w", err)
	}

	knownHostsPath := path.Join(home, ".ssh", "known_hosts")
	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts failed: %s: %w\nRun 'ssh-keyscan %s >> ~/.ssh/known_hosts' first", knownHostsPath, err, address)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         SSHConnectionTimeout,
	}

	// Ensure port is present
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = address + ":22"
	}

	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed: %s: %w", address, err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create sftp client failed: %s: %w", address, err)
	}

	return &SftpFS{
		client: client,
		conn:   conn,
	}, nil
}

// Close releases the SFTP client and the underlying SSH connection.
func (s *SftpFS) Close() error {
	var errs []string
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("sftp client: %v", err))
		}
	}
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("ssh connection: %v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close failed: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (s *SftpFS) ReadDir(ctx context.Context, p string) ([]os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	infos, err := s.client.ReadDir(p)
	return infos, WrapError(err, "ReadDir")
}

func (s *SftpFS) Stat(ctx context.Context, p string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := s.client.Stat(p)
	return info, WrapError(err, "Stat")
}

func (s *SftpFS) Lstat(ctx context.Context, p string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := s.client.Lstat(p)
	return info, WrapError(err, "Lstat")
}

func (s *SftpFS) RemoveAll(ctx context.Context, p string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	info, err := s.client.Stat(p)
	if err != nil {
		return WrapError(err, "RemoveAll")
	}

	if !info.IsDir() {
		return WrapError(s.client.Remove(p), "RemoveAll")
	}

	entries, err := s.client.ReadDir(p)
	if err != nil {
		return WrapError(err, "RemoveAll")
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		childPath := path.Join(p, entry.Name())
		if err := s.RemoveAll(ctx, childPath); err != nil {
			return err
		}
	}

	return WrapError(s.client.RemoveDirectory(p), "RemoveAll")
}

func (s *SftpFS) Rename(ctx context.Context, oldPath, newPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return WrapError(s.client.Rename(oldPath, newPath), "Rename")
}

func (s *SftpFS) Create(ctx context.Context, p string) (io.WriteCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := s.client.Create(p)
	return f, WrapError(err, "Create")
}

func (s *SftpFS) Open(ctx context.Context, p string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := s.client.Open(p)
	return f, WrapError(err, "Open")
}

func (s *SftpFS) MkdirAll(ctx context.Context, p string, perm os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return WrapError(s.client.MkdirAll(p), "MkdirAll")
}

func (s *SftpFS) Chmod(ctx context.Context, p string, mode os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return WrapError(s.client.Chmod(p, mode), "Chmod")
}

func (s *SftpFS) GetHomeDir() (string, error) {
	return s.client.Getwd()
}

func (s *SftpFS) Separator() string {
	return "/"
}

func (s *SftpFS) IsLocal() bool {
	return false
}

func (s *SftpFS) Join(elem ...string) string {
	return path.Join(elem...)
}

func (s *SftpFS) Abs(p string) (string, error) {
	if path.IsAbs(p) {
		return path.Clean(p), nil
	}
	wd, err := s.client.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(wd, p), nil
}

func (s *SftpFS) Dir(p string) string {
	return path.Dir(p)
}

func (s *SftpFS) Base(p string) string {
	return path.Base(p)
}

func (s *SftpFS) GetGitStatus(ctx context.Context, p string) (map[string]string, string) {
	if s.conn == nil {
		return nil, ""
	}

	select {
	case <-ctx.Done():
		return nil, ""
	default:
	}

	rootCmd := fmt.Sprintf("git -C %s rev-parse --show-toplevel", p)
	session, err := s.conn.NewSession()
	if err != nil {
		return nil, ""
	}
	defer session.Close()
	out, err := session.Output(rootCmd)
	if err != nil {
		return nil, ""
	}
	repoRoot := strings.TrimSpace(string(out))

	session2, err := s.conn.NewSession()
	if err == nil {
		defer session2.Close()
		branchCmd := fmt.Sprintf("git -C %s rev-parse --abbrev-ref HEAD", p)
		branchOut, _ := session2.Output(branchCmd)
		branch := strings.TrimSpace(string(branchOut))

		session3, err := s.conn.NewSession()
		if err == nil {
			defer session3.Close()
			statusCmd := fmt.Sprintf("git -C %s status --porcelain --ignored", repoRoot)
			statusOut, _ := session3.Output(statusCmd)

			statuses := ParseGitStatusPorcelain(string(statusOut), repoRoot, p)
			return statuses, branch
		}
		return nil, branch
	}

	return nil, ""
}

func (s *SftpFS) IsReadOnly(ctx context.Context, p string) (bool, error) {
	// SFTP doesn't easily expose mount flags. Default to false.
	return false, nil
}

// GetDirSize calculates the total size of a directory recursively for SFTP.
func (s *SftpFS) GetDirSize(ctx context.Context, path string) int64 {
	var size int64
	visited := make(map[string]bool)

	var walk func(string, int) error
	walk = func(currPath string, depth int) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if depth > MaxDirectoryDepth {
			return nil
		}

		// Prevent symlink loops
		realPath, err := s.Abs(currPath)
		if err == nil {
			if visited[realPath] {
				return nil
			}
			visited[realPath] = true
		}

		entries, err := s.ReadDir(ctx, currPath)
		if err != nil {
			return nil
		}
		for _, e := range entries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if e.IsDir() {
				walk(s.Join(currPath, e.Name()), depth+1)
			} else {
				size += e.Size()
			}
		}
		return nil
	}

	walk(path, 0)
	return size
}
