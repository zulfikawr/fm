package testutil

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

// MockFile is a thread-safe mock implementation of a file that satisfies
// io.ReadWriteCloser and io.ReadSeeker.
type MockFile struct {
	mu     sync.Mutex
	buf    *bytes.Buffer
	reader *bytes.Reader
	closed bool

	NameStr  string
	ModeBits os.FileMode
}

func NewMockFile(name string, data []byte) *MockFile {
	return &MockFile{
		buf:      bytes.NewBuffer(data),
		reader:   bytes.NewReader(data),
		NameStr:  name,
		ModeBits: 0644,
	}
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.reader.Read(p)
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.buf.Write(p)
}

func (m *MockFile) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.reader.Seek(offset, whence)
}

func (m *MockFile) Bytes() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buf.Bytes()
}

func (m *MockFile) Stat() (os.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &MockFileInfo{
		NameStr:    m.NameStr,
		SizeInt:    int64(m.buf.Len()),
		ModeBits:   m.ModeBits,
		ModTimeVal: time.Now(),
		IsDirBool:  false,
	}, nil
}
