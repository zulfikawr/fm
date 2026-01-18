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
func (fs *SftpFS) Close() error {
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

func (fs *SftpFS) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-ticker.C:
			fs.mu.RLock()
			conn := fs.conn
			fs.mu.RUnlock()

			if conn != nil {
				// Send a global request as a keep-alive heartbeat
				_, _, _ = conn.SendRequest("keepalive@openssh.com", true, nil)
			}
		}
	}
}

func (fs *SftpFS) reconnect() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Close old connection if still open
	if fs.client != nil {
		fs.client.Close()
	}
	if fs.conn != nil {
		fs.conn.Close()
	}

	// Dial again
	conn, err := ssh.Dial("tcp", fs.address, fs.config)
	if err != nil {
		return fmt.Errorf("reconnect dial failed: %w", err)
	}

	client, err := sftp.NewClient(conn, sftp.UseConcurrentWrites(true))
	if err != nil {
		conn.Close()
		return fmt.Errorf("reconnect sftp failed: %w", err)
	}

	fs.conn = conn
	fs.client = client
	return nil
}

func (fs *SftpFS) isConnectionError(err error) bool {
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

func (fs *SftpFS) ReadDirEntries(ctx context.Context, p string) ([]os.DirEntry, error) {
	var entries []os.DirEntry
	err := fs.runWithRetry(func() error {
		infos, err := fs.client.ReadDir(p)
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

func (fs *SftpFS) runWithRetry(fn func() error) error {
	// Try 1: Standard attempt
	fs.mu.RLock()
	client := fs.client
	fs.mu.RUnlock()

	if client == nil {
		if err := fs.reconnect(); err != nil {
			return err
		}
	}

	err := fn()
	if fs.isConnectionError(err) {
		// Try 2: Reconnect and retry once
		if recErr := fs.reconnect(); recErr == nil {
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

func (fs *SftpFS) ReadDir(ctx context.Context, p string) ([]os.FileInfo, error) {
	if entries, ok := fs.cache.Get(p); ok {
		return entries, nil
	}

	var entries []os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		entries, err = fs.client.ReadDir(p)
		return err
	})

	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDir", p)
	}

	fs.cache.Put(p, entries)
	return entries, nil
}

func (fs *SftpFS) Stat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		info, err = fs.client.Stat(p)
		return err
	})
	return info, errors.WrapErrorWithPath(err, "Stat", p)
}

func (fs *SftpFS) Lstat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		info, err = fs.client.Lstat(p)
		return err
	})
	return info, errors.WrapErrorWithPath(err, "Lstat", p)
}

func (fs *SftpFS) RemoveAll(ctx context.Context, p string) error {
	fs.cache.Invalidate(path.Dir(p))
	return fs.runWithRetry(func() error {
		info, err := fs.client.Stat(p)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fs.client.Remove(p)
		}

		entries, err := fs.client.ReadDir(p)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			childPath := path.Join(p, entry.Name())
			if err := fs.RemoveAll(ctx, childPath); err != nil {
				return err
			}
		}
		return fs.client.RemoveDirectory(p)
	})
}

func (fs *SftpFS) Rename(ctx context.Context, oldPath, newPath string) error {
	fs.cache.Invalidate(path.Dir(oldPath))
	fs.cache.Invalidate(path.Dir(newPath))
	err := fs.runWithRetry(func() error {
		return fs.client.Rename(oldPath, newPath)
	})
	return errors.WrapErrorWithPath(err, "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

func (fs *SftpFS) Create(ctx context.Context, p string) (io.WriteCloser, error) {
	fs.cache.Invalidate(path.Dir(p))
	var f io.WriteCloser
	err := fs.runWithRetry(func() error {
		var err error
		f, err = fs.client.Create(p)
		return err
	})
	return f, errors.WrapErrorWithPath(err, "Create", p)
}

func (fs *SftpFS) Open(ctx context.Context, p string) (io.ReadCloser, error) {
	var f io.ReadCloser
	err := fs.runWithRetry(func() error {
		var err error
		f, err = fs.client.Open(p)
		return err
	})
	return f, errors.WrapErrorWithPath(err, "Open", p)
}

func (fs *SftpFS) MkdirAll(ctx context.Context, p string, perm os.FileMode) error {
	fs.cache.Invalidate(path.Dir(p))
	return fs.runWithRetry(func() error {
		return fs.client.MkdirAll(p)
	})
}

func (fs *SftpFS) Chmod(ctx context.Context, p string, mode os.FileMode) error {
	return fs.runWithRetry(func() error {
		return fs.client.Chmod(p, mode)
	})
}

func (fs *SftpFS) Preallocate(ctx context.Context, path string, size int64) error {
	// SFTP doesn't support fallocate natively.
	// We could use StatVFS to check free space, but for now we'll keep it as a no-op.
	return nil
}

func (fs *SftpFS) GetHomeDir() (string, error) {
	dir, err := fs.client.Getwd()
	return dir, errors.WrapError(err, "GetHomeDir")
}

func (fs *SftpFS) Separator() string {
	return "/"
}

func (fs *SftpFS) IsLocal() bool {
	return false
}

func (fs *SftpFS) Address() string {
	return fs.address
}

func (fs *SftpFS) User() string {
	return fs.user
}

func (fs *SftpFS) Join(elem ...string) string {
	return path.Join(elem...)
}

func (fs *SftpFS) Abs(p string) (string, error) {
	if path.IsAbs(p) {
		return path.Clean(p), nil
	}
	wd, err := fs.client.Getwd()
	if err != nil {
		return "", errors.WrapError(err, "Abs")
	}
	return path.Join(wd, p), nil
}

func (fs *SftpFS) Rel(basepath, targpath string) (string, error) {
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

func (fs *SftpFS) Clean(p string) string {
	return path.Clean(p)
}

func (fs *SftpFS) Dir(p string) string {
	return path.Dir(p)
}

func (fs *SftpFS) Base(p string) string {
	return path.Base(p)
}

func (fs *SftpFS) Ext(p string) string {
	return path.Ext(p)
}

func (fs *SftpFS) IsReadOnly(ctx context.Context, p string) (bool, error) {
	var isReadOnly bool
	err := fs.runWithRetry(func() error {
		// 1. Try StatVFS extension (OpenSSH) to check mount flags
		if vfs, err := fs.client.StatVFS(p); err == nil {
			// 1 is ST_RDONLY on most systems
			if vfs.Flag&1 != 0 {
				isReadOnly = true
				return nil
			}
		}

		// 2. Fallback to checking permission bits of the directory/file itself
		info, err := fs.client.Stat(p)
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

func (fs *SftpFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(16) // Concurrency limit for parallel walking

	// Start with the root
	info, err := fs.Stat(ctx, root)
	if err := walkFn(root, info, err); err != nil {
		if err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if info == nil || !info.IsDir() {
		return nil
	}

	if err := fs.parallelWalk(ctx, g, root, walkFn); err != nil {
		return err
	}

	return g.Wait()
}

func (fs *SftpFS) parallelWalk(ctx context.Context, g *errgroup.Group, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	// Read current directory entries
	entries, err := fs.ReadDir(ctx, root)
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
				case <-fs.ctx.Done():
					return fmt.Errorf("filesystem closed")
				default:
					return fs.parallelWalk(ctx, g, p, walkFn)
				}
			})
		}
	}

	return nil
}
