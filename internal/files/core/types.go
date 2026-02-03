package core

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// Reader provides read-only filesystem operations
type Reader interface {
	ReadDir(ctx context.Context, path string) ([]os.FileInfo, error)
	ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error)
	Stat(ctx context.Context, path string) (os.FileInfo, error)
	Lstat(ctx context.Context, path string) (os.FileInfo, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
}

// Writer provides write filesystem operations
type Writer interface {
	Create(ctx context.Context, path string) (io.WriteCloser, error)
	MkdirAll(ctx context.Context, path string, perm os.FileMode) error
	RemoveAll(ctx context.Context, path string) error
	Rename(ctx context.Context, oldPath, newPath string) error
	Chmod(ctx context.Context, path string, mode os.FileMode) error
	Preallocate(ctx context.Context, path string, size int64) error
}

// PathResolver provides path manipulation operations
type PathResolver interface {
	Separator() string
	Join(elem ...string) string
	Abs(path string) (string, error)
	Rel(basepath, targpath string) (string, error)
	Clean(path string) string
	Dir(path string) string
	Base(path string) string
	Ext(path string) string
}

// AnalysisResult represents the disk usage of a path
type AnalysisResult struct {
	Path        string
	Name        string
	Size        int64
	IsDirectory bool
	Children    []*AnalysisResult
	Parent      *AnalysisResult
	Percentage  float64 // Relative to parent
}

// FileSystem defines the interface for file system operations.
type FileSystem interface {
	Reader
	Writer
	PathResolver
	GetHomeDir() (string, error)
	IsLocal() bool
	IsReadOnly(ctx context.Context, path string) (bool, error)
	Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error
	Address() string
	User() string
	Close() error
}
