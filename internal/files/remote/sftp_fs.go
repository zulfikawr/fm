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
	"time"

	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/files/errors"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SftpFS implements FileSystem for SFTP.
type SftpFS struct {
	client *sftp.Client
	conn   *ssh.Client
	cache  *core.MetadataCache
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

	return &SftpFS{
			client: client,
			conn:   conn,
			cache:  core.NewMetadataCache(2 * time.Second),
		},
		nil
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
		return errors.WrapError(fmt.Errorf("close failed: %s", strings.Join(errs, ", ")), "Close")
	}
	return nil
}

func (s *SftpFS) ReadDirEntries(ctx context.Context, p string) ([]os.DirEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// sftp.Client.ReadDir returns []os.FileInfo, which also satisfies os.DirEntry
	infos, err := s.client.ReadDir(p)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDirEntries", p)
	}

	entries := make([]os.DirEntry, len(infos))
	for i, info := range infos {
		entries[i] = infoToDirEntry(info)
	}
	return entries, nil
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

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := s.client.ReadDir(p)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDir", p)
	}

	s.cache.Put(p, entries)

	return entries, nil
}

func (s *SftpFS) Stat(ctx context.Context, p string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := s.client.Stat(p)
	return info, errors.WrapErrorWithPath(err, "Stat", p)
}

func (s *SftpFS) Lstat(ctx context.Context, p string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := s.client.Lstat(p)
	return info, errors.WrapErrorWithPath(err, "Lstat", p)
}

func (s *SftpFS) RemoveAll(ctx context.Context, p string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.cache.Invalidate(path.Dir(p))
	info, err := s.client.Stat(p)
	if !info.IsDir() {
		return errors.WrapErrorWithPath(s.client.Remove(p), "RemoveAll", p)
	}

	entries, err := s.client.ReadDir(p)
	if err != nil {
		return errors.WrapErrorWithPath(err, "RemoveAll", p)
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

	return errors.WrapErrorWithPath(s.client.RemoveDirectory(p), "RemoveAll", p)
}

func (s *SftpFS) Rename(ctx context.Context, oldPath, newPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.cache.Invalidate(path.Dir(oldPath))
	s.cache.Invalidate(path.Dir(newPath))
	return errors.WrapErrorWithPath(s.client.Rename(oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

func (s *SftpFS) Create(ctx context.Context, p string) (io.WriteCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.cache.Invalidate(path.Dir(p))
	f, err := s.client.Create(p)
	return f, errors.WrapErrorWithPath(err, "Create", p)
}

func (s *SftpFS) Open(ctx context.Context, p string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := s.client.Open(p)
	return f, errors.WrapErrorWithPath(err, "Open", p)
}

func (s *SftpFS) MkdirAll(ctx context.Context, p string, perm os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.cache.Invalidate(path.Dir(p))
	return errors.WrapErrorWithPath(s.client.MkdirAll(p), "MkdirAll", p)
}

func (s *SftpFS) Chmod(ctx context.Context, p string, mode os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return errors.WrapErrorWithPath(s.client.Chmod(p, mode), "Chmod", p)
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
	// 1. Try StatVFS extension (OpenSSH) to check mount flags
	if vfs, err := s.client.StatVFS(p); err == nil {
		// 1 is ST_RDONLY on most systems
		if vfs.Flag&1 != 0 {
			return true, nil
		}
	}

	// 2. Fallback to checking permission bits of the directory/file itself
	info, err := s.client.Stat(p)
	if err != nil {
		return false, errors.WrapErrorWithPath(err, "IsReadOnly", p)
	}

	// Check if current user (we assume they own the session) has write bit
	// This is a simplified check as we don't know the remote UID/GID match
	return info.Mode().Perm()&0o200 == 0, nil
}

func (s *SftpFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	walker := s.client.Walk(root)
	for walker.Step() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := walker.Err(); err != nil {
			if err := walkFn(walker.Path(), nil, err); err != nil {
				if err == filepath.SkipDir {
					walker.SkipDir()
					continue
				}
				return err
			}
			continue
		}
		if err := walkFn(walker.Path(), walker.Stat(), nil); err != nil {
			if err == filepath.SkipDir {
				walker.SkipDir()
				continue
			}
			return err
		}
	}
	return nil
}
