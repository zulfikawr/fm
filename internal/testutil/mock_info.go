package testutil

import (
	"os"
	"time"
)

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	NameStr    string
	SizeInt    int64
	ModeBits   os.FileMode
	ModTimeVal time.Time
	IsDirBool  bool
}

func (m *MockFileInfo) Name() string       { return m.NameStr }
func (m *MockFileInfo) Size() int64        { return m.SizeInt }
func (m *MockFileInfo) Mode() os.FileMode  { return m.ModeBits }
func (m *MockFileInfo) IsDir() bool        { return m.IsDirBool }
func (m *MockFileInfo) Sys() interface{}   { return nil }
func (m *MockFileInfo) ModTime() time.Time { return m.ModTimeVal }

// SetPermissions is a helper to easily set common permissions
func (m *MockFileInfo) SetPermissions(readable, writable bool) {
	var mode os.FileMode
	if readable {
		mode |= 0444
	}
	if writable {
		mode |= 0222
	}
	if m.IsDirBool {
		mode |= os.ModeDir
	}
	m.ModeBits = mode
}

// MockDirEntry implements os.DirEntry for testing
type MockDirEntry struct {
	NameStr   string
	IsDirBool bool
	TypeBits  os.FileMode
	InfoVal   os.FileInfo
}

func (m *MockDirEntry) Name() string      { return m.NameStr }
func (m *MockDirEntry) IsDir() bool       { return m.IsDirBool }
func (m *MockDirEntry) Type() os.FileMode { return m.TypeBits }

// WithInfo sets the FileInfo to be returned by Info()
func (m *MockDirEntry) WithInfo(info os.FileInfo) *MockDirEntry {
	m.InfoVal = info
	return m
}

func (m *MockDirEntry) Info() (os.FileInfo, error) {

	if m.InfoVal != nil {
		return m.InfoVal, nil
	}
	return &MockFileInfo{
		NameStr:    m.NameStr,
		IsDirBool:  m.IsDirBool,
		ModeBits:   m.TypeBits,
		ModTimeVal: time.Now(),
	}, nil
}
