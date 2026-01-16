package remote

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/files/errors"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/sync/errgroup"
)

// SftpFS implements FileSystem for SFTP.
type SftpFS struct {
	mu     sync.RWMutex
	client *sftp.Client
	conn   *ssh.Client
	cache  *core.MetadataCache

	// Connection details for reconnection
	address string
	user    string
	config  *ssh.ClientConfig

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSftpFS creates a new SFTP file system.
func NewSftpFS(address, user, password, keyPath string, hostKeyCallback ssh.HostKeyCallback) (*SftpFS, error) {
	auths := []ssh.AuthMethod{}

	// 1. Try SSH Agent
	if sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		agentClient := agent.NewClient(sock)
		auths = append(auths, ssh.PublicKeysCallback(agentClient.Signers))
	}

	// 2. Try Identity File (Key)
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file %s: %w", keyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key file %s: %w", keyPath, err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}

	// 3. Try Password
	if password != "" {
		auths = append(auths, ssh.Password(password))
	}

	// Setup HostKeyCallback - REQUIRE known_hosts for security by default
	if hostKeyCallback == nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.WrapError(err, "NewSftpFS")
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

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         constants.SSHConnectionTimeout,
	}

	// Ensure port is present
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = address + ":22"
	}

	conn, err := ssh.Dial("tcp", address, config)
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
	fs := &SftpFS{
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
func (s *SftpFS) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
		return errors.WrapError(fmt.Errorf("close failed: %s", strings.Join(errs, ", ")), "Close")
	}
	return nil
}

func (s *SftpFS) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			conn := s.conn
			s.mu.RUnlock()

			if conn != nil {
				// Send a global request as a keep-alive heartbeat
				_, _, err := conn.SendRequest("keepalive@openssh.com", true, nil)
				if err != nil {
					// Connection might be dead, but let runWithRetry handle the actual reconnection
					// when a real operation is attempted.
				}
			}
		}
	}
}

func (s *SftpFS) reconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close old connection if still open
	if s.client != nil {
		s.client.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}

	// Dial again
	conn, err := ssh.Dial("tcp", s.address, s.config)
	if err != nil {
		return fmt.Errorf("reconnect dial failed: %w", err)
	}

	client, err := sftp.NewClient(conn, sftp.UseConcurrentWrites(true))
	if err != nil {
		conn.Close()
		return fmt.Errorf("reconnect sftp failed: %w", err)
	}

	s.conn = conn
	s.client = client
	return nil
}

func (s *SftpFS) isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return err == io.EOF ||
		err == io.ErrUnexpectedEOF ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection lost") ||
		strings.Contains(errStr, "connection closed") ||
		strings.Contains(errStr, "use of closed network connection")
}

func (s *SftpFS) ReadDirEntries(ctx context.Context, p string) ([]os.DirEntry, error) {
	var entries []os.DirEntry
	err := s.runWithRetry(func() error {
		infos, err := s.client.ReadDir(p)
		if err != nil {
			return err
		}
		entries = make([]os.DirEntry, len(infos))
		for i, info := range infos {
			entries[i] = infoToDirEntry(info)
		}
		return nil
	})
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDirEntries", p)
	}
	return entries, nil
}

func (s *SftpFS) runWithRetry(fn func() error) error {
	// Try 1: Standard attempt
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil {
		if err := s.reconnect(); err != nil {
			return err
		}
	}

	err := fn()
	if s.isConnectionError(err) {
		// Try 2: Reconnect and retry once
		if recErr := s.reconnect(); recErr == nil {
			return fn()
		}
	}
	return err
}

type dirEntry struct {
	info os.FileInfo
}

func (d *dirEntry) Name() string               { return d.info.Name() }
func (d *dirEntry) IsDir() bool                { return d.info.IsDir() }
func (d *dirEntry) Type() os.FileMode          { return d.info.Mode().Type() }
func (d *dirEntry) Info() (os.FileInfo, error) { return d.info, nil }

func infoToDirEntry(info os.FileInfo) os.DirEntry {
	return &dirEntry{info: info}
}

func (s *SftpFS) ReadDir(ctx context.Context, p string) ([]os.FileInfo, error) {
	if entries, ok := s.cache.Get(p); ok {
		return entries, nil
	}

	var entries []os.FileInfo
	err := s.runWithRetry(func() error {
		var err error
		entries, err = s.client.ReadDir(p)
		return err
	})

	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDir", p)
	}

	s.cache.Put(p, entries)
	return entries, nil
}

func (s *SftpFS) Stat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := s.runWithRetry(func() error {
		var err error
		info, err = s.client.Stat(p)
		return err
	})
	return info, errors.WrapErrorWithPath(err, "Stat", p)
}

func (s *SftpFS) Lstat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := s.runWithRetry(func() error {
		var err error
		info, err = s.client.Lstat(p)
		return err
	})
	return info, errors.WrapErrorWithPath(err, "Lstat", p)
}

func (s *SftpFS) RemoveAll(ctx context.Context, p string) error {
	s.cache.Invalidate(path.Dir(p))
	return s.runWithRetry(func() error {
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
			if err := s.RemoveAll(ctx, childPath); err != nil {
				return err
			}
		}
		return s.client.RemoveDirectory(p)
	})
}

func (s *SftpFS) Rename(ctx context.Context, oldPath, newPath string) error {
	s.cache.Invalidate(path.Dir(oldPath))
	s.cache.Invalidate(path.Dir(newPath))
	err := s.runWithRetry(func() error {
		return s.client.Rename(oldPath, newPath)
	})
	return errors.WrapErrorWithPath(err, "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

func (s *SftpFS) Create(ctx context.Context, p string) (io.WriteCloser, error) {
	s.cache.Invalidate(path.Dir(p))
	var f io.WriteCloser
	err := s.runWithRetry(func() error {
		var err error
		f, err = s.client.Create(p)
		return err
	})
	return f, errors.WrapErrorWithPath(err, "Create", p)
}

func (s *SftpFS) Open(ctx context.Context, p string) (io.ReadCloser, error) {
	var f io.ReadCloser
	err := s.runWithRetry(func() error {
		var err error
		f, err = s.client.Open(p)
		return err
	})
	return f, errors.WrapErrorWithPath(err, "Open", p)
}

func (s *SftpFS) MkdirAll(ctx context.Context, p string, perm os.FileMode) error {
	s.cache.Invalidate(path.Dir(p))
	return s.runWithRetry(func() error {
		return s.client.MkdirAll(p)
	})
}

func (s *SftpFS) Chmod(ctx context.Context, p string, mode os.FileMode) error {
	return s.runWithRetry(func() error {
		return s.client.Chmod(p, mode)
	})
}

func (s *SftpFS) Preallocate(ctx context.Context, path string, size int64) error {
	// SFTP doesn't support fallocate natively.
	// We could use StatVFS to check free space, but for now we'll keep it as a no-op.
	return nil
}

func (s *SftpFS) GetHomeDir() (string, error) {
	dir, err := s.client.Getwd()
	return dir, errors.WrapError(err, "GetHomeDir")
}

func (s *SftpFS) Separator() string {
	return "/"
}

func (s *SftpFS) IsLocal() bool {
	return false
}

func (s *SftpFS) Address() string {
	return s.address
}

func (s *SftpFS) User() string {
	return s.user
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
		return "", errors.WrapError(err, "Abs")
	}
	return path.Join(wd, p), nil
}

func (s *SftpFS) Rel(basepath, targpath string) (string, error) {
	base := path.Clean(basepath)
	targ := path.Clean(targpath)

	if base == targ {
		return ".", nil
	}

	// Simple implementation for slash-based paths
	baseElems := strings.Split(strings.Trim(base, "/"), "/")
	targElems := strings.Split(strings.Trim(targ, "/"), "/")

	if base == "/" {
		baseElems = []string{}
	}
	if targ == "/" {
		targElems = []string{}
	}

	i := 0
	for i < len(baseElems) && i < len(targElems) && baseElems[i] == targElems[i] {
		i++
	}

	var rel []string
	for j := i; j < len(baseElems); j++ {
		rel = append(rel, "..")
	}
	rel = append(rel, targElems[i:]...)

	if len(rel) == 0 {
		return ".", nil
	}

	return strings.Join(rel, "/"), nil
}

func (s *SftpFS) Clean(p string) string {
	return path.Clean(p)
}

func (s *SftpFS) Dir(p string) string {
	return path.Dir(p)
}

func (s *SftpFS) Base(p string) string {
	return path.Base(p)
}

func (s *SftpFS) Ext(p string) string {
	return path.Ext(p)
}

func (s *SftpFS) IsReadOnly(ctx context.Context, p string) (bool, error) {
	var isReadOnly bool
	err := s.runWithRetry(func() error {
		// 1. Try StatVFS extension (OpenSSH) to check mount flags
		if vfs, err := s.client.StatVFS(p); err == nil {
			// 1 is ST_RDONLY on most systems
			if vfs.Flag&1 != 0 {
				isReadOnly = true
				return nil
			}
		}

		// 2. Fallback to checking permission bits of the directory/file itself
		info, err := s.client.Stat(p)
		if err != nil {
			return err
		}

		// Check if current user (we assume they own the session) has write bit
		isReadOnly = info.Mode().Perm()&0o200 == 0
		return nil
	})

	if err != nil {
		return false, errors.WrapErrorWithPath(err, "IsReadOnly", p)
	}
	return isReadOnly, nil
}

func (s *SftpFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(16) // Concurrency limit for parallel walking

	// Start with the root
	info, err := s.Stat(ctx, root)
	if err := walkFn(root, info, err); err != nil {
		if err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if info == nil || !info.IsDir() {
		return nil
	}

	return s.parallelWalk(ctx, g, root, walkFn)
}

func (s *SftpFS) parallelWalk(ctx context.Context, g *errgroup.Group, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	// Read current directory entries
	entries, err := s.ReadDir(ctx, root)
	if err != nil {
		return walkFn(root, nil, err)
	}

	for _, entry := range entries {
		entry := entry // capture
		p := path.Join(root, entry.Name())

		// Report the entry
		if err := walkFn(p, entry, nil); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}

		// Recursively walk subdirectories in parallel
		if entry.IsDir() {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-s.ctx.Done():
					return fmt.Errorf("filesystem closed")
				default:
					return s.parallelWalk(ctx, g, p, walkFn)
				}
			})
		}
	}

	return nil
}
