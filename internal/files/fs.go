package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"golang.org/x/sys/unix"
)

// FileSystem defines the interface for file system operations.
type FileSystem interface {
	ReadDir(ctx context.Context, path string) ([]os.FileInfo, error)
	Stat(ctx context.Context, path string) (os.FileInfo, error)
	Lstat(ctx context.Context, path string) (os.FileInfo, error)
	RemoveAll(ctx context.Context, path string) error
	Rename(ctx context.Context, oldPath, newPath string) error
	Create(ctx context.Context, path string) (io.WriteCloser, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	MkdirAll(ctx context.Context, path string, perm os.FileMode) error
	Chmod(ctx context.Context, path string, mode os.FileMode) error
	GetHomeDir() (string, error)
	Separator() string
	IsLocal() bool
	Join(elem ...string) string
	Abs(path string) (string, error)
	Dir(path string) string
	Base(path string) string
	GetGitStatus(ctx context.Context, path string) (map[string]string, string)
	IsReadOnly(ctx context.Context, path string) (bool, error)
	GetDirSize(ctx context.Context, path string) int64
	Close() error
}

const MaxDirectoryDepth = 50

// LocalFS implements FileSystem for the local disk.
type LocalFS struct{}

func (l *LocalFS) Close() error {
	return nil
}

// sizeSem limits the total number of concurrent directory walking goroutines
// across all GetDirSize calls to avoid overwhelming the system.
var sizeSem = make(chan struct{}, runtime.NumCPU()*4)

// GetDirSize calculates the total size of a directory for LocalFS using a fast, concurrent approach.
func (l *LocalFS) GetDirSize(ctx context.Context, path string) int64 {
	var totalSize int64
	var wg sync.WaitGroup

	// Track visited inodes to prevent loops (symlinks, etc.)
	visited := sync.Map{}

	// Optional: Get the device ID of the starting path to implement "one file system" (du -x)
	// This prevents counting /proc, /sys, etc. when measuring root.
	var startDev uint64
	if info, err := os.Stat(path); err == nil {
		if sys := info.Sys(); sys != nil {
			if stat, ok := sys.(*syscall.Stat_t); ok {
				startDev = uint64(stat.Dev)
			}
		}
	}

	var walk func(string, int)
	walk = func(p string, depth int) {
		defer wg.Done()

		select {
		case <-ctx.Done():
			return
		default:
		}

		if depth > MaxDirectoryDepth {
			return
		}

		entries, err := os.ReadDir(p)
		if err != nil {
			return
		}

		for _, entry := range entries {
			select {
			case <-ctx.Done():
				return
			default:
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			var size int64 = info.Size()
			// Use actual disk usage (blocks) if available on Unix
			if sys := info.Sys(); sys != nil {
				if stat, ok := sys.(*syscall.Stat_t); ok {
					// st_blocks is always in 512-byte units in POSIX/Linux
					size = stat.Blocks * 512

					// If we are on a different device (e.g. /proc inside /), skip it
					// unless we explicitly started inside that device.
					if startDev != 0 && uint64(stat.Dev) != startDev {
						continue
					}

					// Loop detection using Inode
					key := fmt.Sprintf("%d:%d", stat.Dev, stat.Ino)
					if _, loaded := visited.LoadOrStore(key, true); loaded {
						continue
					}
				}
			}

			if info.IsDir() {
				// Count the size of the directory entry itself
				atomic.AddInt64(&totalSize, size)

				wg.Add(1)
				select {
				case sizeSem <- struct{}{}:
					go func(dirPath string, d int) {
						walk(dirPath, d)
						<-sizeSem
					}(filepath.Join(p, info.Name()), depth+1)
				default:
					// If semaphore is full, just walk synchronously
					walk(filepath.Join(p, info.Name()), depth+1)
				}
			} else if info.Mode().IsRegular() || (info.Mode()&os.ModeSymlink != 0) {
				// Count regular files and symlinks
				atomic.AddInt64(&totalSize, size)
			}
		}
	}
	wg.Add(1)
	walk(path, 0)
	wg.Wait()

	return totalSize
}

func (l *LocalFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, WrapError(err, "ReadDir")
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
	return info, WrapError(err, "Stat")
}

func (l *LocalFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	info, err := os.Lstat(path)
	return info, WrapError(err, "Lstat")
}

func (l *LocalFS) RemoveAll(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return WrapError(os.RemoveAll(path), "RemoveAll")
}

func (l *LocalFS) Rename(ctx context.Context, oldPath, newPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return WrapError(os.Rename(oldPath, newPath), "Rename")
}

func (l *LocalFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := os.Create(path)
	return f, WrapError(err, "Create")
}

func (l *LocalFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	f, err := os.Open(path)
	return f, WrapError(err, "Open")
}

func (l *LocalFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return WrapError(os.MkdirAll(path, perm), "MkdirAll")
}

func (l *LocalFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return WrapError(os.Chmod(path, mode), "Chmod")
}

func (l *LocalFS) GetHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	return dir, WrapError(err, "GetHomeDir")
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

func (l *LocalFS) GetGitStatus(ctx context.Context, path string) (map[string]string, string) {
	return GetGitStatus(ctx, path)
}

func (l *LocalFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		var stat unix.Statfs_t
		if err := unix.Statfs(path, &stat); err != nil {
			return false, err
		}
		// Check for MS_RDONLY (linux) or MNT_RDONLY (darwin)
		// On many systems MS_RDONLY is 1
		return (stat.Flags & unix.MS_RDONLY) != 0, nil
	}
	// For other OSs, we can try to create a temporary file or check attributes
	// For now, default to false.
	return false, nil
}
