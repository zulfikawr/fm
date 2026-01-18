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

func (fs *LocalFS) Close() error {
	return nil
}

func (fs *LocalFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	entries, err := os.ReadDir(path)
	return entries, errors.WrapErrorWithPath(err, "ReadDirEntries", path)
}

func (fs *LocalFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	// Check cache
	if entries, ok := fs.cache.Get(path); ok {
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
	fs.cache.Put(path, result)

	return result, nil
}

func (fs *LocalFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := os.Stat(path)
	return info, errors.WrapErrorWithPath(err, "Stat", path)
}

func (fs *LocalFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := os.Lstat(path)
	return info, errors.WrapErrorWithPath(err, "Lstat", path)
}

func (fs *LocalFS) RemoveAll(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fs.cache.Invalidate(filepath.Dir(path))
	return errors.WrapErrorWithPath(os.RemoveAll(path), "RemoveAll", path)
}

func (fs *LocalFS) Rename(ctx context.Context, oldPath, newPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fs.cache.Invalidate(filepath.Dir(oldPath))
	fs.cache.Invalidate(filepath.Dir(newPath))
	return errors.WrapErrorWithPath(os.Rename(oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

func (fs *LocalFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	fs.cache.Invalidate(filepath.Dir(path))
	f, err := os.Create(path)
	return f, errors.WrapErrorWithPath(err, "Create", path)
}

func (fs *LocalFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := os.Open(path)
	return f, errors.WrapErrorWithPath(err, "Open", path)
}

func (fs *LocalFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	fs.cache.Invalidate(filepath.Dir(path))
	return errors.WrapErrorWithPath(os.MkdirAll(path, perm), "MkdirAll", path)
}

func (fs *LocalFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return errors.WrapErrorWithPath(os.Chmod(path, mode), "Chmod", path)
}

func (fs *LocalFS) Preallocate(ctx context.Context, path string, size int64) error {
	return errors.WrapErrorWithPath(preallocate(path, size), "Preallocate", path)
}

func (fs *LocalFS) GetHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	return dir, errors.WrapError(err, "GetHomeDir")
}

func (fs *LocalFS) Separator() string {
	return string(os.PathSeparator)
}

func (fs *LocalFS) IsLocal() bool {
	return true
}

func (fs *LocalFS) Address() string {
	return ""
}

func (fs *LocalFS) User() string {
	return ""
}

func (fs *LocalFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (fs *LocalFS) Abs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	return abs, errors.WrapErrorWithPath(err, "Abs", path)
}

func (fs *LocalFS) Rel(basepath, targpath string) (string, error) {
	rel, err := filepath.Rel(basepath, targpath)
	return rel, errors.WrapError(err, "Rel")
}

func (fs *LocalFS) Clean(path string) string {
	return filepath.Clean(path)
}

func (fs *LocalFS) Dir(path string) string {
	return filepath.Dir(path)
}

func (fs *LocalFS) Base(path string) string {
	return filepath.Base(path)
}

func (fs *LocalFS) Ext(path string) string {
	return filepath.Ext(path)
}

func (fs *LocalFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	isRO, err := isReadOnly(path)
	return isRO, errors.WrapErrorWithPath(err, "IsReadOnly", path)
}

func (fs *LocalFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return walkFn(path, info, err)
	})
	return errors.WrapErrorWithPath(err, "Walk", root)
}
