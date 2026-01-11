package files

import (
	"context"
	"io"
	"os"
)

type MockFS struct {
	FileSystem
	RenameFunc    func(ctx context.Context, oldPath, newPath string) error
	StatFunc      func(ctx context.Context, path string) (os.FileInfo, error)
	LstatFunc     func(ctx context.Context, path string) (os.FileInfo, error)
	OpenFunc      func(ctx context.Context, path string) (io.ReadCloser, error)
	CreateFunc    func(ctx context.Context, path string) (io.WriteCloser, error)
	RemoveAllFunc func(ctx context.Context, path string) error
	MkdirAllFunc  func(ctx context.Context, path string, perm os.FileMode) error
	ReadDirFunc   func(ctx context.Context, path string) ([]os.FileInfo, error)
}

func (m *MockFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if m.ReadDirFunc != nil {
		return m.ReadDirFunc(ctx, path)
	}
	return m.FileSystem.ReadDir(ctx, path)
}

func (m *MockFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(ctx, path, perm)
	}
	return m.FileSystem.MkdirAll(ctx, path, perm)
}

func (m *MockFS) RemoveAll(ctx context.Context, path string) error {
	if m.RemoveAllFunc != nil {
		return m.RemoveAllFunc(ctx, path)
	}
	return m.FileSystem.RemoveAll(ctx, path)
}

func (m *MockFS) Rename(ctx context.Context, oldPath, newPath string) error {
	if m.RenameFunc != nil {
		return m.RenameFunc(ctx, oldPath, newPath)
	}
	return m.FileSystem.Rename(ctx, oldPath, newPath)
}

func (m *MockFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(ctx, path)
	}
	return m.FileSystem.Stat(ctx, path)
}

func (m *MockFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if m.LstatFunc != nil {
		return m.LstatFunc(ctx, path)
	}
	return m.FileSystem.Lstat(ctx, path)
}

func (m *MockFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(ctx, path)
	}
	return m.FileSystem.Open(ctx, path)
}

func (m *MockFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, path)
	}
	return m.FileSystem.Create(ctx, path)
}

func (m *MockFS) Join(elem ...string) string {
	return m.FileSystem.Join(elem...)
}

func (m *MockFS) Base(path string) string {
	return m.FileSystem.Base(path)
}

func (m *MockFS) GetDirSize(ctx context.Context, path string) int64 {
	return m.FileSystem.GetDirSize(ctx, path)
}
