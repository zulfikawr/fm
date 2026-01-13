package testutil

import (
	"context"
	"fm/internal/files/core"
	"io"
	"os"
	"time"
)

// MockGitService implements a minimal GitService for testing
type MockGitService struct{}

func (m *MockGitService) IsEnabled() bool         { return true }
func (m *MockGitService) SetEnabled(enabled bool) {}
func (m *MockGitService) GetStatus(ctx context.Context, path string) (map[string]string, string) {
	return make(map[string]string), "main"
}
func (m *MockGitService) GetRoot(ctx context.Context, path string) string {
	return "/repo"
}

// NewMockGitService creates a new NewMockGitService with default behaviors
func NewMockGitService() *MockGitService {
	return &MockGitService{}
}

// MockFileSystem implements core.FileSystem for testing
type MockFileSystem struct {
	core.FileSystem
	ReadDirFunc    func(ctx context.Context, path string) ([]os.FileInfo, error)
	StatFunc       func(ctx context.Context, path string) (os.FileInfo, error)
	LstatFunc      func(ctx context.Context, path string) (os.FileInfo, error)
	RemoveAllFunc  func(ctx context.Context, path string) error
	RenameFunc     func(ctx context.Context, oldPath, newPath string) error
	CreateFunc     func(ctx context.Context, path string) (io.WriteCloser, error)
	OpenFunc       func(ctx context.Context, path string) (io.ReadCloser, error)
	MkdirAllFunc   func(ctx context.Context, path string, perm os.FileMode) error
	ChmodFunc      func(ctx context.Context, path string, mode os.FileMode) error
	GetHomeDirFunc func() (string, error)
	SeparatorFunc  func() string
	IsLocalFunc    func() bool
	JoinFunc       func(elem ...string) string
	AbsFunc        func(path string) (string, error)
	DirFunc        func(path string) string
	BaseFunc       func(path string) string
	IsReadOnlyFunc func(ctx context.Context, path string) (bool, error)
	CloseFunc      func() error
}

func (m *MockFileSystem) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	if m.ReadDirFunc != nil {
		return m.ReadDirFunc(ctx, path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.ReadDir(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(ctx, path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Stat(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	if m.LstatFunc != nil {
		return m.LstatFunc(ctx, path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Lstat(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) RemoveAll(ctx context.Context, path string) error {
	if m.RemoveAllFunc != nil {
		return m.RemoveAllFunc(ctx, path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.RemoveAll(ctx, path)
	}
	return nil
}

func (m *MockFileSystem) Rename(ctx context.Context, oldPath, newPath string) error {
	if m.RenameFunc != nil {
		return m.RenameFunc(ctx, oldPath, newPath)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Rename(ctx, oldPath, newPath)
	}
	return nil
}

func (m *MockFileSystem) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Create(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(ctx, path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Open(ctx, path)
	}
	return nil, nil
}

func (m *MockFileSystem) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(ctx, path, perm)
	}
	if m.FileSystem != nil {
		return m.FileSystem.MkdirAll(ctx, path, perm)
	}
	return nil
}

func (m *MockFileSystem) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	if m.ChmodFunc != nil {
		return m.ChmodFunc(ctx, path, mode)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Chmod(ctx, path, mode)
	}
	return nil
}

func (m *MockFileSystem) GetHomeDir() (string, error) {
	if m.GetHomeDirFunc != nil {
		return m.GetHomeDirFunc()
	}
	if m.FileSystem != nil {
		return m.FileSystem.GetHomeDir()
	}
	return "", nil
}

func (m *MockFileSystem) Separator() string {
	if m.SeparatorFunc != nil {
		return m.SeparatorFunc()
	}
	if m.FileSystem != nil {
		return m.FileSystem.Separator()
	}
	return "/"
}

func (m *MockFileSystem) IsLocal() bool {
	if m.IsLocalFunc != nil {
		return m.IsLocalFunc()
	}
	if m.FileSystem != nil {
		return m.FileSystem.IsLocal()
	}
	return true
}

func (m *MockFileSystem) Join(elem ...string) string {
	if m.JoinFunc != nil {
		return m.JoinFunc(elem...)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Join(elem...)
	}
	return ""
}

func (m *MockFileSystem) Abs(path string) (string, error) {
	if m.AbsFunc != nil {
		return m.AbsFunc(path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Abs(path)
	}
	return path, nil
}

func (m *MockFileSystem) Dir(path string) string {
	if m.DirFunc != nil {
		return m.DirFunc(path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Dir(path)
	}
	return ""
}

func (m *MockFileSystem) Base(path string) string {
	if m.BaseFunc != nil {
		return m.BaseFunc(path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.Base(path)
	}
	return ""
}

func (m *MockFileSystem) IsReadOnly(ctx context.Context, path string) (bool, error) {
	if m.IsReadOnlyFunc != nil {
		return m.IsReadOnlyFunc(ctx, path)
	}
	if m.FileSystem != nil {
		return m.FileSystem.IsReadOnly(ctx, path)
	}
	return false, nil
}

func (m *MockFileSystem) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if m.FileSystem != nil {
		return m.FileSystem.Close()
	}
	return nil
}

// NewMockFileSystem creates a new MockFileSystem with default behaviors
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{}
}

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	os.FileInfo
	NameStr    string
	SizeInt    int64
	ModeBits   os.FileMode
	ModTimeVal interface{}
	IsDirBool  bool
}

func (m *MockFileInfo) Name() string       { return m.NameStr }
func (m *MockFileInfo) Size() int64        { return m.SizeInt }
func (m *MockFileInfo) Mode() os.FileMode  { return m.ModeBits }
func (m *MockFileInfo) IsDir() bool        { return m.IsDirBool }
func (m *MockFileInfo) Sys() interface{}   { return nil }
func (m *MockFileInfo) ModTime() time.Time { return time.Time{} }
