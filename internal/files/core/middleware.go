package core

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/zulfikawr/fm/internal/files/errors"
)

// BaseFS is a helper struct that delegates all FileSystem methods to an underlying implementation.
type BaseFS struct {
	FS FileSystem
}

func (b *BaseFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	return b.FS.ReadDir(ctx, path)
}

func (b *BaseFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	return b.FS.ReadDirEntries(ctx, path)
}

func (b *BaseFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	return b.FS.Stat(ctx, path)
}

func (b *BaseFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return b.FS.Lstat(ctx, path)
}

func (b *BaseFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	return b.FS.Open(ctx, path)
}

func (b *BaseFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	return b.FS.Create(ctx, path)
}

func (b *BaseFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return b.FS.MkdirAll(ctx, path, perm)
}

func (b *BaseFS) RemoveAll(ctx context.Context, path string) error {
	return b.FS.RemoveAll(ctx, path)
}

func (b *BaseFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return b.FS.Rename(ctx, oldPath, newPath)
}

func (b *BaseFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return b.FS.Chmod(ctx, path, mode)
}

func (b *BaseFS) Preallocate(ctx context.Context, path string, size int64) error {
	return b.FS.Preallocate(ctx, path, size)
}

func (b *BaseFS) Separator() string                             { return b.FS.Separator() }
func (b *BaseFS) Join(elem ...string) string                    { return b.FS.Join(elem...) }
func (b *BaseFS) Abs(path string) (string, error)               { return b.FS.Abs(path) }
func (b *BaseFS) Rel(basepath, targpath string) (string, error) { return b.FS.Rel(basepath, targpath) }
func (b *BaseFS) Clean(path string) string                      { return b.FS.Clean(path) }
func (b *BaseFS) Dir(path string) string                        { return b.FS.Dir(path) }
func (b *BaseFS) Base(path string) string                       { return b.FS.Base(path) }
func (b *BaseFS) Ext(path string) string                        { return b.FS.Ext(path) }

func (b *BaseFS) GetHomeDir() (string, error) { return b.FS.GetHomeDir() }
func (b *BaseFS) IsLocal() bool               { return b.FS.IsLocal() }
func (b *BaseFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	return b.FS.IsReadOnly(ctx, path)
}
func (b *BaseFS) Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
	return b.FS.Walk(ctx, root, walkFn)
}
func (b *BaseFS) Address() string { return b.FS.Address() }
func (b *BaseFS) User() string    { return b.FS.User() }
func (b *BaseFS) Close() error    { return b.FS.Close() }

// ContextFS ensures context cancellation is checked before every operation.
type ContextFS struct {
	BaseFS
}

func NewContextFS(fs FileSystem) *ContextFS {
	return &ContextFS{BaseFS{FS: fs}}
}

func (c *ContextFS) check(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (c *ContextFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := c.check(ctx); err != nil {
		return nil, err
	}
	return c.FS.ReadDir(ctx, path)
}

func (c *ContextFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	if err := c.check(ctx); err != nil {
		return nil, err
	}
	return c.FS.ReadDirEntries(ctx, path)
}

func (c *ContextFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := c.check(ctx); err != nil {
		return nil, err
	}
	return c.FS.Stat(ctx, path)
}

func (c *ContextFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := c.check(ctx); err != nil {
		return nil, err
	}
	return c.FS.Lstat(ctx, path)
}

func (c *ContextFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	if err := c.check(ctx); err != nil {
		return nil, err
	}
	return c.FS.Open(ctx, path)
}

func (c *ContextFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	if err := c.check(ctx); err != nil {
		return nil, err
	}
	return c.FS.Create(ctx, path)
}

func (c *ContextFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	if err := c.check(ctx); err != nil {
		return err
	}
	return c.FS.MkdirAll(ctx, path, perm)
}

func (c *ContextFS) RemoveAll(ctx context.Context, path string) error {
	if err := c.check(ctx); err != nil {
		return err
	}
	return c.FS.RemoveAll(ctx, path)
}

func (c *ContextFS) Rename(ctx context.Context, oldPath, newPath string) error {
	if err := c.check(ctx); err != nil {
		return err
	}
	return c.FS.Rename(ctx, oldPath, newPath)
}

func (c *ContextFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	if err := c.check(ctx); err != nil {
		return err
	}
	return c.FS.Chmod(ctx, path, mode)
}

func (c *ContextFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	if err := c.check(ctx); err != nil {
		return false, err
	}
	return c.FS.IsReadOnly(ctx, path)
}

func (c *ContextFS) Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
	if err := c.check(ctx); err != nil {
		return err
	}
	return c.FS.Walk(ctx, root, walkFn)
}

// ErrorWrappedFS wraps errors with operation names and paths.
type ErrorWrappedFS struct {
	BaseFS
}

func NewErrorWrappedFS(fs FileSystem) *ErrorWrappedFS {
	return &ErrorWrappedFS{BaseFS{FS: fs}}
}

func (e *ErrorWrappedFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	res, err := e.FS.ReadDir(ctx, path)
	return res, errors.WrapErrorWithPath(err, "ReadDir", path)
}

func (e *ErrorWrappedFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	res, err := e.FS.ReadDirEntries(ctx, path)
	return res, errors.WrapErrorWithPath(err, "ReadDirEntries", path)
}

func (e *ErrorWrappedFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	res, err := e.FS.Stat(ctx, path)
	return res, errors.WrapErrorWithPath(err, "Stat", path)
}

func (e *ErrorWrappedFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	res, err := e.FS.Lstat(ctx, path)
	return res, errors.WrapErrorWithPath(err, "Lstat", path)
}

func (e *ErrorWrappedFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	res, err := e.FS.Open(ctx, path)
	return res, errors.WrapErrorWithPath(err, "Open", path)
}

func (e *ErrorWrappedFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	res, err := e.FS.Create(ctx, path)
	return res, errors.WrapErrorWithPath(err, "Create", path)
}

func (e *ErrorWrappedFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	err := e.FS.MkdirAll(ctx, path, perm)
	return errors.WrapErrorWithPath(err, "MkdirAll", path)
}

func (e *ErrorWrappedFS) RemoveAll(ctx context.Context, path string) error {
	err := e.FS.RemoveAll(ctx, path)
	return errors.WrapErrorWithPath(err, "RemoveAll", path)
}

func (e *ErrorWrappedFS) Rename(ctx context.Context, oldPath, newPath string) error {
	err := e.FS.Rename(ctx, oldPath, newPath)
	return errors.WrapErrorWithPath(err, "Rename", oldPath+" -> "+newPath)
}

func (e *ErrorWrappedFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	err := e.FS.Chmod(ctx, path, mode)
	return errors.WrapErrorWithPath(err, "Chmod", path)
}

func (e *ErrorWrappedFS) Abs(path string) (string, error) {
	res, err := e.FS.Abs(path)
	return res, errors.WrapErrorWithPath(err, "Abs", path)
}

func (e *ErrorWrappedFS) Rel(basepath, targpath string) (string, error) {
	res, err := e.FS.Rel(basepath, targpath)
	return res, errors.WrapError(err, "Rel")
}

func (e *ErrorWrappedFS) GetHomeDir() (string, error) {
	res, err := e.FS.GetHomeDir()
	return res, errors.WrapError(err, "GetHomeDir")
}

func (e *ErrorWrappedFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	res, err := e.FS.IsReadOnly(ctx, path)
	return res, errors.WrapErrorWithPath(err, "IsReadOnly", path)
}

func (e *ErrorWrappedFS) Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
	err := e.FS.Walk(ctx, root, walkFn)
	return errors.WrapErrorWithPath(err, "Walk", root)
}

// CachedFS adds a caching layer for ReadDir and ensures invalidation on writes.
type CachedFS struct {
	BaseFS
	cache *SimpleCache[string, []os.FileInfo]
}

func NewCachedFS(fs FileSystem, capacity int, ttl time.Duration) *CachedFS {
	return &CachedFS{
		BaseFS: BaseFS{FS: fs},
		cache:  NewSimpleCache[string, []os.FileInfo](capacity, ttl),
	}
}

func (c *CachedFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if entries, ok := c.cache.Get(path); ok {
		return entries, nil
	}
	entries, err := c.FS.ReadDir(ctx, path)
	if err == nil {
		c.cache.Put(path, entries)
	}
	return entries, err
}

func (c *CachedFS) invalidate(path string) {
	c.cache.Invalidate(c.FS.Dir(path))
}

func (c *CachedFS) RemoveAll(ctx context.Context, path string) error {
	c.invalidate(path)
	return c.FS.RemoveAll(ctx, path)
}

func (c *CachedFS) Rename(ctx context.Context, oldPath, newPath string) error {
	c.invalidate(oldPath)
	c.invalidate(newPath)
	return c.FS.Rename(ctx, oldPath, newPath)
}

func (c *CachedFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	c.invalidate(path)
	return c.FS.Create(ctx, path)
}

func (c *CachedFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	c.invalidate(path)
	return c.FS.MkdirAll(ctx, path, perm)
}

func (c *CachedFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	c.invalidate(path)
	return c.FS.Chmod(ctx, path, mode)
}
