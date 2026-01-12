package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"fm/internal/files/errors"
)

// LocalFS implements FileSystem for the local disk.
type LocalFS struct{}

func (l *LocalFS) Close() error {
	return nil
}

func (l *LocalFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDir", path)
	}
	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (l *LocalFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := os.Stat(path)
	return info, errors.WrapErrorWithPath(err, "Stat", path)
}

func (l *LocalFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := os.Lstat(path)
	return info, errors.WrapErrorWithPath(err, "Lstat", path)
}

func (l *LocalFS) RemoveAll(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return errors.WrapErrorWithPath(os.RemoveAll(path), "RemoveAll", path)
}

func (l *LocalFS) Rename(ctx context.Context, oldPath, newPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return errors.WrapErrorWithPath(os.Rename(oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

func (l *LocalFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := os.Create(path)
	return f, errors.WrapErrorWithPath(err, "Create", path)
}

func (l *LocalFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := os.Open(path)
	return f, errors.WrapErrorWithPath(err, "Open", path)
}

func (l *LocalFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return errors.WrapErrorWithPath(os.MkdirAll(path, perm), "MkdirAll", path)
}

func (l *LocalFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return errors.WrapErrorWithPath(os.Chmod(path, mode), "Chmod", path)
}

func (l *LocalFS) GetHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	return dir, errors.WrapError(err, "GetHomeDir")
}

func (l *LocalFS) Separator() string {
	return string(os.PathSeparator)
}

func (l *LocalFS) IsLocal() bool {
	return true
}

func (l *LocalFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (l *LocalFS) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

func (l *LocalFS) Dir(path string) string {
	return filepath.Dir(path)
}

func (l *LocalFS) Base(path string) string {
	return filepath.Base(path)
}

func (l *LocalFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	// Try to detect read-only filesystem by attempting to create a test file
	tmpFile := filepath.Join(path, ".fm-readonly-test-"+fmt.Sprintf("%d", os.Getpid()))

	// Try to create a temporary file
	f, err := os.Create(tmpFile)
	if err != nil {
		// If we get a permission denied error, it's likely read-only
		return os.IsPermission(err), nil
	}

	// Clean up the test file
	f.Close()
	os.Remove(tmpFile)
	return false, nil
}
