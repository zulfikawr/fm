package testutil

import (
	"io"
	"os"
	"time"
)

// MockFileInfo implements os.FileInfo for testing.
type MockFileInfo struct {
	FName    string
	FSize    int64
	FMode    os.FileMode
	FModTime time.Time
	FIsDir   bool
}

func (m *MockFileInfo) Name() string       { return m.FName }
func (m *MockFileInfo) Size() int64        { return m.FSize }
func (m *MockFileInfo) Mode() os.FileMode  { return m.FMode }
func (m *MockFileInfo) ModTime() time.Time { return m.FModTime }
func (m *MockFileInfo) IsDir() bool        { return m.FIsDir }
func (m *MockFileInfo) Sys() interface{}   { return nil }

// MockReadWriteCloser is a generic mock for io.ReadWriteCloser.
type MockReadWriteCloser struct {
	ReadFunc  func(p []byte) (n int, err error)
	WriteFunc func(p []byte) (n int, err error)
	CloseFunc func() error
}

func (m *MockReadWriteCloser) Read(p []byte) (int, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}
	return 0, io.EOF
}

func (m *MockReadWriteCloser) Write(p []byte) (int, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	return len(p), nil
}

func (m *MockReadWriteCloser) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockDirEntry implements os.DirEntry for testing.
type MockDirEntry struct {
	NameStr   string
	IsDirBool bool
	ModeBits  os.FileMode
	InfoErr   error
}

func (m *MockDirEntry) Name() string      { return m.NameStr }
func (m *MockDirEntry) IsDir() bool       { return m.IsDirBool }
func (m *MockDirEntry) Type() os.FileMode { return m.ModeBits }
func (m *MockDirEntry) Info() (os.FileInfo, error) {
	if m.InfoErr != nil {
		return nil, m.InfoErr
	}
	return &MockFileInfo{FName: m.NameStr, FIsDir: m.IsDirBool, FMode: m.ModeBits}, nil
}

// NewMockFile returns a MockReadWriteCloser initialized with data.
func NewMockFile(name string, data []byte) *MockReadWriteCloser {
	off := 0
	return &MockReadWriteCloser{
		ReadFunc: func(p []byte) (int, error) {
			if off >= len(data) {
				return 0, io.EOF
			}
			n := copy(p, data[off:])
			off += n
			return n, nil
		},
		WriteFunc: func(p []byte) (int, error) {
			return len(p), nil
		},
	}
}
