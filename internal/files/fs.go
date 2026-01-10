package files

import (
	"io"
	"os"
	"path/filepath"
)

// FileSystem defines the interface for file system operations.
type FileSystem interface {
	ReadDir(path string) ([]os.FileInfo, error)
	Stat(path string) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error)
	RemoveAll(path string) error
	Rename(oldPath, newPath string) error
	Create(path string) (io.WriteCloser, error)
	Open(path string) (io.ReadCloser, error)
	MkdirAll(path string, perm os.FileMode) error
	Chmod(path string, mode os.FileMode) error
	GetHomeDir() (string, error)
	Separator() string
	IsLocal() bool
	Join(elem ...string) string
	Abs(path string) (string, error)
	Dir(path string) string
	Base(path string) string
	GetGitStatus(path string) (map[string]string, string)
	Close() error
}

// LocalFS implements FileSystem for the local disk.
type LocalFS struct{}

func (l *LocalFS) Close() error {
	return nil
}

func (l *LocalFS) ReadDir(path string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (l *LocalFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (l *LocalFS) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (l *LocalFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (l *LocalFS) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (l *LocalFS) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (l *LocalFS) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (l *LocalFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (l *LocalFS) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

func (l *LocalFS) GetHomeDir() (string, error) {
	return os.UserHomeDir()
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

func (l *LocalFS) GetGitStatus(path string) (map[string]string, string) {
	return GetGitStatus(path)
}
