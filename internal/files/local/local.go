package local

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/zulfikawr/fm/internal/constants"

	"golang.org/x/sync/errgroup"
)

// LocalFS implements FileSystem for the local disk.
type LocalFS struct{}

func NewLocalFS() *LocalFS {
	return &LocalFS{}
}

func (fs *LocalFS) Close() error {
	return nil
}

func (fs *LocalFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func (fs *LocalFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	infos := make([]os.FileInfo, len(entries))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(constants.MaxReadDirWorkers)

	for i, entry := range entries {
		g.Go(func() error {
			select {
			case <-gctx.Done():
				return gctx.Err()
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
		return nil, err
	}

	// Filter out nil entries (where Info() failed)
	result := make([]os.FileInfo, 0, len(infos))
	for _, info := range infos {
		if info != nil {
			result = append(result, info)
		}
	}

	return result, nil
}

func (fs *LocalFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (fs *LocalFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (fs *LocalFS) RemoveAll(ctx context.Context, path string) error {
	return os.RemoveAll(path)
}

func (fs *LocalFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (fs *LocalFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (fs *LocalFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (fs *LocalFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *LocalFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

func (fs *LocalFS) Preallocate(ctx context.Context, path string, size int64) error {
	return preallocate(path, size)
}

func (fs *LocalFS) GetHomeDir() (string, error) {
	return os.UserHomeDir()
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
	return filepath.Abs(path)
}

func (fs *LocalFS) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
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
	return isReadOnly(path)
}

func (fs *LocalFS) Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}
