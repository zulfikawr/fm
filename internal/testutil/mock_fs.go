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
	AddressFunc        func() string
	UserFunc           func() string
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

func (fs *MockFileSystem) recordCall(method string, args ...any) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, Call{Method: method, Args: args})
}

// AssertCalled verifies that a method was called
func (fs *MockFileSystem) AssertCalled(t TB, method string) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	for _, c := range fs.Calls {
		if c.Method == method {
			return
		}
	}
	t.Errorf("expected method %s to be called, but it was not", method)
}

func (fs *MockFileSystem) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	fs.recordCall("ReadDir", path)
	if fs.ReadDirFunc != nil {
		return fs.ReadDirFunc(ctx, path)
	}
	return []os.FileInfo{}, fs.DefaultError
}

func (fs *MockFileSystem) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	fs.recordCall("ReadDirEntries", path)
	if fs.ReadDirEntriesFunc != nil {
		return fs.ReadDirEntriesFunc(ctx, path)
	}
	return []os.DirEntry{}, fs.DefaultError
}

func (fs *MockFileSystem) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	fs.recordCall("Stat", path)
	if fs.StatFunc != nil {
		return fs.StatFunc(ctx, path)
	}
	return &MockFileInfo{NameStr: "mock"}, fs.DefaultError
}

func (fs *MockFileSystem) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	fs.recordCall("Lstat", path)
	if fs.LstatFunc != nil {
		return fs.LstatFunc(ctx, path)
	}
	return &MockFileInfo{NameStr: "mock"}, fs.DefaultError
}

func (fs *MockFileSystem) RemoveAll(ctx context.Context, path string) error {
	fs.recordCall("RemoveAll", path)
	if fs.RemoveAllFunc != nil {
		return fs.RemoveAllFunc(ctx, path)
	}
	return fs.DefaultError
}

func (fs *MockFileSystem) Rename(ctx context.Context, oldPath, newPath string) error {
	fs.recordCall("Rename", oldPath, newPath)
	if fs.RenameFunc != nil {
		return fs.RenameFunc(ctx, oldPath, newPath)
	}
	return fs.DefaultError
}

func (fs *MockFileSystem) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	fs.recordCall("Create", path)
	if fs.CreateFunc != nil {
		return fs.CreateFunc(ctx, path)
	}
	return nil, fs.DefaultError
}

func (fs *MockFileSystem) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	fs.recordCall("Open", path)
	if fs.OpenFunc != nil {
		return fs.OpenFunc(ctx, path)
	}
	return nil, fs.DefaultError
}

func (fs *MockFileSystem) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	fs.recordCall("MkdirAll", path, perm)
	if fs.MkdirAllFunc != nil {
		return fs.MkdirAllFunc(ctx, path, perm)
	}
	return fs.DefaultError
}

func (fs *MockFileSystem) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	fs.recordCall("Chmod", path, mode)
	if fs.ChmodFunc != nil {
		return fs.ChmodFunc(ctx, path, mode)
	}
	return fs.DefaultError
}

func (fs *MockFileSystem) Preallocate(ctx context.Context, path string, size int64) error {
	fs.recordCall("Preallocate", path, size)
	if fs.PreallocateFunc != nil {
		return fs.PreallocateFunc(ctx, path, size)
	}
	return fs.DefaultError
}

func (fs *MockFileSystem) GetHomeDir() (string, error) {
	fs.recordCall("GetHomeDir")
	if fs.GetHomeDirFunc != nil {
		return fs.GetHomeDirFunc()
	}
	return "/home/user", fs.DefaultError
}

func (fs *MockFileSystem) Separator() string {
	if fs.SeparatorFunc != nil {
		return fs.SeparatorFunc()
	}
	return "/"
}

func (fs *MockFileSystem) IsLocal() bool {
	if fs.IsLocalFunc != nil {
		return fs.IsLocalFunc()
	}
	return true
}

func (fs *MockFileSystem) Join(elem ...string) string {
	if fs.JoinFunc != nil {
		return fs.JoinFunc(elem...)
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

func (fs *MockFileSystem) Abs(path string) (string, error) {
	if fs.AbsFunc != nil {
		return fs.AbsFunc(path)
	}
	return path, fs.DefaultError
}

func (fs *MockFileSystem) Rel(basepath, targpath string) (string, error) {
	if fs.RelFunc != nil {
		return fs.RelFunc(basepath, targpath)
	}
	return targpath, fs.DefaultError
}

func (fs *MockFileSystem) Clean(path string) string {
	if fs.CleanFunc != nil {
		return fs.CleanFunc(path)
	}
	return path
}

func (fs *MockFileSystem) Dir(path string) string {
	if fs.DirFunc != nil {
		return fs.DirFunc(path)
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

func (fs *MockFileSystem) Base(path string) string {
	if fs.BaseFunc != nil {
		return fs.BaseFunc(path)
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

func (fs *MockFileSystem) Ext(path string) string {
	if fs.ExtFunc != nil {
		return fs.ExtFunc(path)
	}
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

func (fs *MockFileSystem) IsReadOnly(ctx context.Context, path string) (bool, error) {
	fs.recordCall("IsReadOnly", path)
	if fs.IsReadOnlyFunc != nil {
		return fs.IsReadOnlyFunc(ctx, path)
	}
	return false, fs.DefaultError
}

func (fs *MockFileSystem) Address() string {
	if fs.AddressFunc != nil {
		return fs.AddressFunc()
	}
	return ""
}

func (fs *MockFileSystem) User() string {
	if fs.UserFunc != nil {
		return fs.UserFunc()
	}
	return ""
}

func (fs *MockFileSystem) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	fs.recordCall("Walk", root)
	if fs.WalkFunc != nil {
		return fs.WalkFunc(ctx, root, walkFn)
	}
	return fs.DefaultError
}

func (fs *MockFileSystem) Close() error {
	fs.recordCall("Close")
	if fs.CloseFunc != nil {
		return fs.CloseFunc()
	}
	return fs.DefaultError
}
