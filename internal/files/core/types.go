// Package core defines the fundamental interfaces and types for filesystem operations.
// It provides abstractions for local and remote filesystem access, enabling uniform
// handling of files across different storage backends.
package core

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// Reader provides read-only filesystem operations.
// All operations accept a context for cancellation and timeout control.
type Reader interface {
	// ReadDir returns file information for all entries in a directory.
	ReadDir(ctx context.Context, path string) ([]os.FileInfo, error)
	// ReadDirEntries returns directory entries for all items in a directory.
	ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error)
	// Stat returns file information following symlinks.
	Stat(ctx context.Context, path string) (os.FileInfo, error)
	// Lstat returns file information without following symlinks.
	Lstat(ctx context.Context, path string) (os.FileInfo, error)
	// Open opens a file for reading.
	Open(ctx context.Context, path string) (io.ReadCloser, error)
}

// Writer provides write filesystem operations.
// All operations accept a context for cancellation and timeout control.
type Writer interface {
	// Create creates or truncates a file for writing.
	Create(ctx context.Context, path string) (io.WriteCloser, error)
	// MkdirAll creates a directory and all necessary parent directories.
	MkdirAll(ctx context.Context, path string, perm os.FileMode) error
	// RemoveAll removes a path and all its contents recursively.
	RemoveAll(ctx context.Context, path string) error
	// Rename moves or renames a file or directory.
	Rename(ctx context.Context, oldPath, newPath string) error
	// Chmod changes the file mode/permissions.
	Chmod(ctx context.Context, path string, mode os.FileMode) error
	// Preallocate reserves disk space for a file to improve write performance.
	Preallocate(ctx context.Context, path string, size int64) error
}

// PathResolver provides path manipulation operations.
// Implementations handle platform-specific path separators and conventions.
type PathResolver interface {
	// Separator returns the path separator character for this filesystem.
	Separator() string
	// Join concatenates path elements into a single path.
	Join(elem ...string) string
	// Abs returns the absolute path.
	Abs(path string) (string, error)
	// Rel returns a relative path from basepath to targpath.
	Rel(basepath, targpath string) (string, error)
	// Clean returns the shortest equivalent path.
	Clean(path string) string
	// Dir returns the directory portion of a path.
	Dir(path string) string
	// Base returns the last element of a path.
	Base(path string) string
	// Ext returns the file extension.
	Ext(path string) string
}

// AnalysisResult represents the disk usage analysis of a path.
// It forms a tree structure where each node contains size information
// and references to its parent and children.
type AnalysisResult struct {
	Path        string             // Full path to the file or directory
	Name        string             // Display name
	Size        int64              // Total size in bytes (recursive for directories)
	IsDirectory bool               // Whether this is a directory
	Children    []*AnalysisResult  // Child nodes (for directories)
	Parent      *AnalysisResult    // Parent node (nil for root)
	Percentage  float64            // Size as percentage of parent (0-1)
}

// FileSystem defines the complete interface for file system operations.
// It combines Reader, Writer, and PathResolver interfaces with additional
// metadata and lifecycle methods. Implementations must be safe for concurrent use.
type FileSystem interface {
	Reader
	Writer
	PathResolver
	// GetHomeDir returns the home directory path for the current user.
	GetHomeDir() (string, error)
	// IsLocal returns true if this is a local filesystem (not remote).
	IsLocal() bool
	// IsReadOnly checks if a path is read-only.
	IsReadOnly(ctx context.Context, path string) (bool, error)
	// Walk traverses a directory tree, calling walkFn for each entry.
	Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error
	// Address returns the connection address (empty for local, host:port for remote).
	Address() string
	// User returns the username (empty for local, SSH user for remote).
	User() string
	// Close releases any resources held by the filesystem.
	Close() error
}
