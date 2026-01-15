package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/files/errors"

	"golang.org/x/sync/errgroup"
)

// LocalFS implements FileSystem for the local disk.
type LocalFS struct {
	cache *core.MetadataCache
}

func NewLocalFS() *LocalFS {
	return &LocalFS{
		cache: core.NewMetadataCache(2 * time.Second),
	}
}

func (l *LocalFS) Close() error {
	return nil
}

func (l *LocalFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	entries, err := os.ReadDir(path)
	return entries, errors.WrapErrorWithPath(err, "ReadDirEntries", path)
}

func (l *LocalFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	// Check cache
	if entries, ok := l.cache.Get(path); ok {
		return entries, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDir", path)
	}

	infos := make([]os.FileInfo, len(entries))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(constants.MaxReadDirWorkers)

	for i, entry := range entries {
		i, entry := i, entry // capture for closure
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			info, err := entry.Info()
			if err == nil {
				infos[i] = info
			}
			return nil // We ignore individual stat errors to show as much as possible
		})
	}

	if err := g.Wait(); err != nil {
		return nil, errors.WrapErrorWithPath(err, "ReadDir", path)
	}

	// Filter out nil entries (where Info() failed)
	result := make([]os.FileInfo, 0, len(infos))
	for _, info := range infos {
		if info != nil {
			result = append(result, info)
		}
	}

	// Store in cache
	l.cache.Put(path, result)

	return result, nil
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
	l.cache.Invalidate(filepath.Dir(path))
	return errors.WrapErrorWithPath(os.RemoveAll(path), "RemoveAll", path)
}

func (l *LocalFS) Rename(ctx context.Context, oldPath, newPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	l.cache.Invalidate(filepath.Dir(oldPath))
	l.cache.Invalidate(filepath.Dir(newPath))
	return errors.WrapErrorWithPath(os.Rename(oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

func (l *LocalFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	l.cache.Invalidate(filepath.Dir(path))
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
	l.cache.Invalidate(filepath.Dir(path))
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

func (l *LocalFS) Preallocate(ctx context.Context, path string, size int64) error {
	return preallocate(path, size)
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

func (l *LocalFS) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

func (l *LocalFS) Clean(path string) string {
	return filepath.Clean(path)
}

func (l *LocalFS) Dir(path string) string {
	return filepath.Dir(path)
}

func (l *LocalFS) Base(path string) string {
	return filepath.Base(path)
}

func (l *LocalFS) Ext(path string) string {
	return filepath.Ext(path)
}

func (l *LocalFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	return isReadOnly(path)
}

func (l *LocalFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return walkFn(path, info, err)
	})
}
