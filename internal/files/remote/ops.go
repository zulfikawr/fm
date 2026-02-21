package remote

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
)

func (fs *RemoteFS) ReadDirEntries(ctx context.Context, p string) ([]os.DirEntry, error) {
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
	return entries, err
}

func (fs *RemoteFS) ReadDir(ctx context.Context, p string) ([]os.FileInfo, error) {
	var entries []os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		entries, err = fs.client.ReadDir(p)
		return err
	})
	return entries, err
}

func (fs *RemoteFS) Stat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		info, err = fs.client.Stat(p)
		return err
	})
	return info, err
}

func (fs *RemoteFS) Lstat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		info, err = fs.client.Lstat(p)
		return err
	})
	return info, err
}

func (fs *RemoteFS) RemoveAll(ctx context.Context, p string) error {
	return fs.runWithRetry(func() error {
		// Check context before starting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
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

		for i := range entries {
			// Check context in loop
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			
			entry := entries[i]
			childPath := path.Join(p, entry.Name())
			if err := fs.RemoveAll(ctx, childPath); err != nil {
				return err
			}
		}
		return fs.client.RemoveDirectory(p)
	})
}

func (fs *RemoteFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return fs.runWithRetry(func() error {
		return fs.client.Rename(oldPath, newPath)
	})
}

func (fs *RemoteFS) Create(ctx context.Context, p string) (io.WriteCloser, error) {
	var f *sftp.File
	err := fs.runWithRetry(func() error {
		fs.mu.RLock()
		client := fs.client
		fs.mu.RUnlock()
		
		if client == nil {
			return fmt.Errorf("sftp client is nil")
		}
		
		var err error
		f, err = client.Create(p)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &remoteFileWrapper{file: f, fs: fs, ctx: ctx}, err
}

func (fs *RemoteFS) Open(ctx context.Context, p string) (io.ReadCloser, error) {
	var f *sftp.File
	err := fs.runWithRetry(func() error {
		fs.mu.RLock()
		client := fs.client
		fs.mu.RUnlock()
		
		if client == nil {
			return fmt.Errorf("sftp client is nil")
		}
		
		var err error
		f, err = client.Open(p)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &remoteFileWrapper{file: f, fs: fs, ctx: ctx}, err
}

// remoteFileWrapper wraps sftp.File to ensure proper cleanup
type remoteFileWrapper struct {
	file *sftp.File
	fs   *RemoteFS
	ctx  context.Context
}

func (w *remoteFileWrapper) Read(p []byte) (n int, err error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
	}
	return w.file.Read(p)
}

func (w *remoteFileWrapper) Write(p []byte) (n int, err error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
	}
	return w.file.Write(p)
}

func (w *remoteFileWrapper) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (fs *RemoteFS) MkdirAll(ctx context.Context, p string, perm os.FileMode) error {
	return fs.runWithRetry(func() error {
		return fs.client.MkdirAll(p)
	})
}

func (fs *RemoteFS) Chmod(ctx context.Context, p string, mode os.FileMode) error {
	return fs.runWithRetry(func() error {
		return fs.client.Chmod(p, mode)
	})
}

func (fs *RemoteFS) Preallocate(ctx context.Context, path string, size int64) error {
	// SFTP doesn't support fallocate natively.
	// We could use StatVFS to check free space, but for now we'll keep it as a no-op.
	return nil
}

func (fs *RemoteFS) IsReadOnly(ctx context.Context, p string) (bool, error) {
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

	return isReadOnly, err
}
