package remote

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"fm/internal/files/errors"
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
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDirEntries", p)
	}
	return entries, nil
}

func (fs *RemoteFS) ReadDir(ctx context.Context, p string) ([]os.FileInfo, error) {
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

func (fs *RemoteFS) Stat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		info, err = fs.client.Stat(p)
		return err
	})
	return info, errors.WrapErrorWithPath(err, "Stat", p)
}

func (fs *RemoteFS) Lstat(ctx context.Context, p string) (os.FileInfo, error) {
	var info os.FileInfo
	err := fs.runWithRetry(func() error {
		var err error
		info, err = fs.client.Lstat(p)
		return err
	})
	return info, errors.WrapErrorWithPath(err, "Lstat", p)
}

func (fs *RemoteFS) RemoveAll(ctx context.Context, p string) error {
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

func (fs *RemoteFS) Rename(ctx context.Context, oldPath, newPath string) error {
	fs.cache.Invalidate(path.Dir(oldPath))
	fs.cache.Invalidate(path.Dir(newPath))
	err := fs.runWithRetry(func() error {
		return fs.client.Rename(oldPath, newPath)
	})
	return errors.WrapErrorWithPath(err, "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

func (fs *RemoteFS) Create(ctx context.Context, p string) (io.WriteCloser, error) {
	fs.cache.Invalidate(path.Dir(p))
	var f io.WriteCloser
	err := fs.runWithRetry(func() error {
		var err error
		f, err = fs.client.Create(p)
		return err
	})
	return f, errors.WrapErrorWithPath(err, "Create", p)
}

func (fs *RemoteFS) Open(ctx context.Context, p string) (io.ReadCloser, error) {
	var f io.ReadCloser
	err := fs.runWithRetry(func() error {
		var err error
		f, err = fs.client.Open(p)
		return err
	})
	return f, errors.WrapErrorWithPath(err, "Open", p)
}

func (fs *RemoteFS) MkdirAll(ctx context.Context, p string, perm os.FileMode) error {
	fs.cache.Invalidate(path.Dir(p))
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

	if err != nil {
		return false, errors.WrapErrorWithPath(err, "IsReadOnly", p)
	}
	return isReadOnly, nil
}
