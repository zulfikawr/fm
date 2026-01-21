package testutil

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// Define minimal interfaces in testutil to avoid import cycles.
// These are compatible with internal/files/core.FileSystem.

type MockFileSystem struct {
	ReadDirFunc        func(ctx context.Context, path string) ([]os.FileInfo, error)
	ReadDirEntriesFunc func(ctx context.Context, path string) ([]os.DirEntry, error)
	StatFunc           func(ctx context.Context, path string) (os.FileInfo, error)
	LstatFunc          func(ctx context.Context, path string) (os.FileInfo, error)
	OpenFunc           func(ctx context.Context, path string) (io.ReadCloser, error)
	CreateFunc         func(ctx context.Context, path string) (io.WriteCloser, error)
	MkdirAllFunc       func(ctx context.Context, path string, perm os.FileMode) error
	RemoveAllFunc      func(ctx context.Context, path string) error
	RenameFunc         func(ctx context.Context, oldPath, newPath string) error
	ChmodFunc          func(ctx context.Context, path string, mode os.FileMode) error
	PreallocateFunc    func(ctx context.Context, path string, size int64) error
	GetHomeDirFunc     func() (string, error)
	IsLocalFunc        func() bool
	IsReadOnlyFunc     func(ctx context.Context, path string) (bool, error)
	WalkFunc           func(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error
	AddressFunc        func() string
	UserFunc           func() string
	CloseFunc          func() error
	SeparatorFunc      func() string
	JoinFunc           func(elem ...string) string
	AbsFunc            func(path string) (string, error)
	RelFunc            func(basepath, targpath string) (string, error)
	CleanFunc          func(path string) string
	DirFunc            func(path string) string
	BaseFunc           func(path string) string
	ExtFunc            func(path string) string
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		IsLocalFunc:   func() bool { return true },
		SeparatorFunc: func() string { return "/" },
	}
}

func (m *MockFileSystem) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if m.ReadDirFunc != nil {
		return m.ReadDirFunc(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	if m.ReadDirEntriesFunc != nil {
		return m.ReadDirEntriesFunc(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(ctx, path)
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if m.LstatFunc != nil {
		return m.LstatFunc(ctx, path)
	}
	return m.Stat(ctx, path)
}

func (m *MockFileSystem) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, path)
	}
	return &MockReadWriteCloser{}, nil
}

func (m *MockFileSystem) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(ctx, path, perm)
	}
	return nil
}

func (m *MockFileSystem) RemoveAll(ctx context.Context, path string) error {
	if m.RemoveAllFunc != nil {
		return m.RemoveAllFunc(ctx, path)
	}
	return nil
}

func (m *MockFileSystem) Rename(ctx context.Context, oldPath, newPath string) error {
	if m.RenameFunc != nil {
		return m.RenameFunc(ctx, oldPath, newPath)
	}
	return nil
}

func (m *MockFileSystem) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	if m.ChmodFunc != nil {
		return m.ChmodFunc(ctx, path, mode)
	}
	return nil
}

func (m *MockFileSystem) Preallocate(ctx context.Context, path string, size int64) error {
	if m.PreallocateFunc != nil {
		return m.PreallocateFunc(ctx, path, size)
	}
	return nil
}

func (m *MockFileSystem) GetHomeDir() (string, error) {
	if m.GetHomeDirFunc != nil {
		return m.GetHomeDirFunc()
	}
	return "/home/user", nil
}

func (m *MockFileSystem) IsLocal() bool {
	if m.IsLocalFunc != nil {
		return m.IsLocalFunc()
	}
	return true
}

func (m *MockFileSystem) IsReadOnly(ctx context.Context, path string) (bool, error) {
	if m.IsReadOnlyFunc != nil {
		return m.IsReadOnlyFunc(ctx, path)
	}
	return false, nil
}

func (m *MockFileSystem) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	if m.WalkFunc != nil {
		return m.WalkFunc(ctx, root, walkFn)
	}
	return nil
}

func (m *MockFileSystem) Address() string {
	if m.AddressFunc != nil {
		return m.AddressFunc()
	}
	return ""
}

func (m *MockFileSystem) User() string {
	if m.UserFunc != nil {
		return m.UserFunc()
	}
	return ""
}

func (m *MockFileSystem) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockFileSystem) Separator() string {
	if m.SeparatorFunc != nil {
		return m.SeparatorFunc()
	}
	return "/"
}

func (m *MockFileSystem) Join(elem ...string) string {
	if m.JoinFunc != nil {
		return m.JoinFunc(elem...)
	}
	return filepath.Join(elem...)
}

func (m *MockFileSystem) Abs(pathStr string) (string, error) {
	if m.AbsFunc != nil {
		return m.AbsFunc(pathStr)
	}
	return pathStr, nil
}

func (m *MockFileSystem) Rel(basepath, targpath string) (string, error) {
	if m.RelFunc != nil {
		return m.RelFunc(basepath, targpath)
	}
	return filepath.Rel(basepath, targpath)
}

func (m *MockFileSystem) Clean(pathStr string) string {
	if m.CleanFunc != nil {
		return m.CleanFunc(pathStr)
	}
	return filepath.Clean(pathStr)
}

func (m *MockFileSystem) Dir(pathStr string) string {
	if m.DirFunc != nil {
		return m.DirFunc(pathStr)
	}
	return filepath.Dir(pathStr)
}

func (m *MockFileSystem) Base(pathStr string) string {
	if m.BaseFunc != nil {
		return m.BaseFunc(pathStr)
	}
	return filepath.Base(pathStr)
}

func (m *MockFileSystem) Ext(pathStr string) string {
	if m.ExtFunc != nil {
		return m.ExtFunc(pathStr)
	}
	return filepath.Ext(pathStr)
}
