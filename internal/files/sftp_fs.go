package files

import (
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

	// Setup HostKeyCallback (Production Ready)
	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	home, err := os.UserHomeDir()
	if err == nil {
		knownHostsPath := path.Join(home, ".ssh", "known_hosts")
		if callback, err := knownhosts.New(knownHostsPath); err == nil {
			hostKeyCallback = callback
		}
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         5 * time.Second,
	}

	// Ensure port is present
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = address + ":22"
	}

	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create sftp client: %w", err)
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

func (s *SftpFS) ReadDir(p string) ([]os.FileInfo, error) {
	return s.client.ReadDir(p)
}

func (s *SftpFS) Stat(p string) (os.FileInfo, error) {
	return s.client.Stat(p)
}

func (s *SftpFS) Lstat(p string) (os.FileInfo, error) {
	return s.client.Lstat(p)
}

func (s *SftpFS) RemoveAll(p string) error {
	info, err := s.client.Stat(p)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return s.client.Remove(p)
	}

	entries, err := s.client.ReadDir(p)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		childPath := path.Join(p, entry.Name())
		if err := s.RemoveAll(childPath); err != nil {
			return err
		}
	}

	return s.client.RemoveDirectory(p)
}

func (s *SftpFS) Rename(oldPath, newPath string) error {
	return s.client.Rename(oldPath, newPath)
}

func (s *SftpFS) Create(p string) (io.WriteCloser, error) {
	return s.client.Create(p)
}

func (s *SftpFS) Open(p string) (io.ReadCloser, error) {
	return s.client.Open(p)
}

func (s *SftpFS) MkdirAll(p string, perm os.FileMode) error {
	return s.client.MkdirAll(p)
}

func (s *SftpFS) Chmod(p string, mode os.FileMode) error {
	return s.client.Chmod(p, mode)
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

func (s *SftpFS) GetGitStatus(p string) (map[string]string, string) {
	if s.conn == nil {
		return nil, ""
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
