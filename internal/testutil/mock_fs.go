package testutil

import (
	"context"
	"io"
	"os"
	"sync"
)

// Call represents a recorded method call to the mock
type Call struct {
	Method string
	Args   []any
}

// MockFileSystem is an advanced, thread-safe mock of core.FileSystem
type MockFileSystem struct {
	mu sync.RWMutex

	// Function pointers for flexible behavior
	ReadDirFunc        func(ctx context.Context, path string) ([]os.FileInfo, error)
	ReadDirEntriesFunc func(ctx context.Context, path string) ([]os.DirEntry, error)
	StatFunc           func(ctx context.Context, path string) (os.FileInfo, error)
	LstatFunc          func(ctx context.Context, path string) (os.FileInfo, error)
	RemoveAllFunc      func(ctx context.Context, path string) error
	RenameFunc         func(ctx context.Context, oldPath, newPath string) error
	CreateFunc         func(ctx context.Context, path string) (io.WriteCloser, error)
	OpenFunc           func(ctx context.Context, path string) (io.ReadCloser, error)
	MkdirAllFunc       func(ctx context.Context, path string, perm os.FileMode) error
	ChmodFunc          func(ctx context.Context, path string, mode os.FileMode) error
	PreallocateFunc    func(ctx context.Context, path string, size int64) error
	GetHomeDirFunc     func() (string, error)
	SeparatorFunc      func() string
	IsLocalFunc        func() bool
	JoinFunc           func(elem ...string) string
	AbsFunc            func(path string) (string, error)
	RelFunc            func(basepath, targpath string) (string, error)
	CleanFunc          func(path string) string
	DirFunc            func(path string) string
	BaseFunc           func(path string) string
	ExtFunc            func(path string) string
	IsReadOnlyFunc     func(ctx context.Context, path string) (bool, error)
	WalkFunc           func(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error
	CloseFunc          func() error

	// Call tracking
	Calls []Call

	// Default return values (if Funcs are nil)
	DefaultError error
}

// NewMockFileSystem creates a new MockFileSystem with sensible defaults
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{}
}

func (m *MockFileSystem) recordCall(method string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, Call{Method: method, Args: args})
}

// AssertCalled verifies that a method was called
func (m *MockFileSystem) AssertCalled(t TB, method string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.Calls {
		if c.Method == method {
			return
		}
	}
	t.Errorf("expected method %s to be called, but it was not", method)
}

func (m *MockFileSystem) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	m.recordCall("ReadDir", path)
	if m.ReadDirFunc != nil {
		return m.ReadDirFunc(ctx, path)
	}
	return []os.FileInfo{}, m.DefaultError
}

func (m *MockFileSystem) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	m.recordCall("ReadDirEntries", path)
	if m.ReadDirEntriesFunc != nil {
		return m.ReadDirEntriesFunc(ctx, path)
	}
	return []os.DirEntry{}, m.DefaultError
}

func (m *MockFileSystem) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	m.recordCall("Stat", path)
	if m.StatFunc != nil {
		return m.StatFunc(ctx, path)
	}
	return &MockFileInfo{NameStr: "mock"}, m.DefaultError
}

func (m *MockFileSystem) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	m.recordCall("Lstat", path)
	if m.LstatFunc != nil {
		return m.LstatFunc(ctx, path)
	}
	return &MockFileInfo{NameStr: "mock"}, m.DefaultError
}

func (m *MockFileSystem) RemoveAll(ctx context.Context, path string) error {
	m.recordCall("RemoveAll", path)
	if m.RemoveAllFunc != nil {
		return m.RemoveAllFunc(ctx, path)
	}
	return m.DefaultError
}

func (m *MockFileSystem) Rename(ctx context.Context, oldPath, newPath string) error {
	m.recordCall("Rename", oldPath, newPath)
	if m.RenameFunc != nil {
		return m.RenameFunc(ctx, oldPath, newPath)
	}
	return m.DefaultError
}

func (m *MockFileSystem) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	m.recordCall("Create", path)
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, path)
	}
	return nil, m.DefaultError
}

func (m *MockFileSystem) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	m.recordCall("Open", path)
	if m.OpenFunc != nil {
		return m.OpenFunc(ctx, path)
	}
	return nil, m.DefaultError
}

func (m *MockFileSystem) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	m.recordCall("MkdirAll", path, perm)
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(ctx, path, perm)
	}
	return m.DefaultError
}

func (m *MockFileSystem) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	m.recordCall("Chmod", path, mode)
	if m.ChmodFunc != nil {
		return m.ChmodFunc(ctx, path, mode)
	}
	return m.DefaultError
}

func (m *MockFileSystem) Preallocate(ctx context.Context, path string, size int64) error {
	m.recordCall("Preallocate", path, size)
	if m.PreallocateFunc != nil {
		return m.PreallocateFunc(ctx, path, size)
	}
	return m.DefaultError
}

func (m *MockFileSystem) GetHomeDir() (string, error) {
	m.recordCall("GetHomeDir")
	if m.GetHomeDirFunc != nil {
		return m.GetHomeDirFunc()
	}
	return "/home/user", m.DefaultError
}

func (m *MockFileSystem) Separator() string {
	if m.SeparatorFunc != nil {
		return m.SeparatorFunc()
	}
	return "/"
}

func (m *MockFileSystem) IsLocal() bool {
	if m.IsLocalFunc != nil {
		return m.IsLocalFunc()
	}
	return true
}

func (m *MockFileSystem) Join(elem ...string) string {
	if m.JoinFunc != nil {
		return m.JoinFunc(elem...)
	}
	res := ""
	for i, e := range elem {
		if i > 0 && res != "" && res[len(res)-1] != '/' {
			res += "/"
		}
		res += e
	}
	return res
}

func (m *MockFileSystem) Abs(path string) (string, error) {
	if m.AbsFunc != nil {
		return m.AbsFunc(path)
	}
	return path, m.DefaultError
}

func (m *MockFileSystem) Rel(basepath, targpath string) (string, error) {
	if m.RelFunc != nil {
		return m.RelFunc(basepath, targpath)
	}
	return targpath, m.DefaultError
}

func (m *MockFileSystem) Clean(path string) string {
	if m.CleanFunc != nil {
		return m.CleanFunc(path)
	}
	return path
}

func (m *MockFileSystem) Dir(path string) string {
	if m.DirFunc != nil {
		return m.DirFunc(path)
	}
	if path == "" || path == "." {
		return "."
	}
	i := len(path) - 1
	for i >= 0 && path[i] == '/' {
		i--
	}
	for i >= 0 && path[i] != '/' {
		i--
	}
	if i < 0 {
		return "."
	}
	res := path[:i]
	if res == "" {
		return "/"
	}
	return res
}

func (m *MockFileSystem) Base(path string) string {
	if m.BaseFunc != nil {
		return m.BaseFunc(path)
	}
	if path == "" {
		return "."
	}
	i := len(path) - 1
	for i >= 0 && path[i] == '/' {
		i--
	}
	if i < 0 {
		return "/"
	}
	end := i + 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	return path[i+1 : end]
}

func (m *MockFileSystem) Ext(path string) string {
	if m.ExtFunc != nil {
		return m.ExtFunc(path)
	}
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

func (m *MockFileSystem) IsReadOnly(ctx context.Context, path string) (bool, error) {
	m.recordCall("IsReadOnly", path)
	if m.IsReadOnlyFunc != nil {
		return m.IsReadOnlyFunc(ctx, path)
	}
	return false, m.DefaultError
}

func (m *MockFileSystem) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	m.recordCall("Walk", root)
	if m.WalkFunc != nil {
		return m.WalkFunc(ctx, root, walkFn)
	}
	return m.DefaultError
}

func (m *MockFileSystem) Close() error {
	m.recordCall("Close")
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return m.DefaultError
}
