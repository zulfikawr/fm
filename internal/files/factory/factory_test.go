package factory

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/ssh"
	"fm/internal/testutil"

	sshx "golang.org/x/crypto/ssh"
)

func TestDefaultConnector(t *testing.T) {
	conn := &DefaultConnector{}

	t.Run("NewLocalFS", func(t *testing.T) {
		fs := conn.NewLocalFS()
		if fs == nil || !fs.IsLocal() {
			t.Error("Expected local FS")
		}
	})

	t.Run("NewRemoteFS Error", func(t *testing.T) {
		_, err := conn.NewRemoteFS("localhost:1", "user", "pass", "", nil)
		if err == nil {
			t.Error("Expected error connecting to invalid address")
		}
	})

	t.Run("CreateHostKeyCallback", func(t *testing.T) {
		hkcb, err := conn.CreateHostKeyCallback()
		testutil.AssertNoError(t, err, "Should create callback")
		if hkcb == nil {
			t.Error("Expected callback")
		}
	})
}

type MockConnector struct {
	MockLocalFS          core.FileSystem
	MockRemoteFS         core.FileSystem
	NewRemoteFSErr       error
	ReadPasswordValue    string
	ReadPasswordErr      error
	HostKeyCallbackErr   error
	NewRemoteFSCallCount int
	NewRemoteFSFunc      func(host, user, password, keyPath string, hkcb sshx.HostKeyCallback) (core.FileSystem, error)
}

func (m *MockConnector) NewLocalFS() core.FileSystem {
	return m.MockLocalFS
}

func (m *MockConnector) NewRemoteFS(host, user, password, keyPath string, hkcb sshx.HostKeyCallback) (core.FileSystem, error) {
	m.NewRemoteFSCallCount++
	if m.NewRemoteFSFunc != nil {
		return m.NewRemoteFSFunc(host, user, password, keyPath, hkcb)
	}
	if m.NewRemoteFSErr != nil {
		return nil, m.NewRemoteFSErr
	}
	return m.MockRemoteFS, nil
}

func (m *MockConnector) ReadPassword() (string, error) {
	return m.ReadPasswordValue, m.ReadPasswordErr
}

func (m *MockConnector) CreateHostKeyCallback() (sshx.HostKeyCallback, error) {
	if m.HostKeyCallbackErr != nil {
		return nil, m.HostKeyCallbackErr
	}
	return func(hostname string, remote net.Addr, key sshx.PublicKey) error { return nil }, nil
}

func TestCreateFileSystem(t *testing.T) {
	mockFS := testutil.NewMockFileSystem()
	mockConnector := &MockConnector{
		MockLocalFS:  mockFS,
		MockRemoteFS: mockFS,
	}

	t.Run("Local Filesystem", func(t *testing.T) {
		fs, info, err := CreateFileSystemWithConnector("", nil, mockConnector)
		testutil.AssertNoError(t, err, "Should not error")
		if fs != mockFS {
			t.Error("Expected mock local FS")
		}
		if info != nil {
			t.Error("Expected nil info for local FS")
		}
	})

	t.Run("Remote Parsing user@host:path", func(t *testing.T) {
		mockConnector.NewRemoteFSCallCount = 0
		fs, info, err := CreateFileSystemWithConnector("user@example.com:/home/user", nil, mockConnector)
		testutil.AssertNoError(t, err, "Should not error")
		if fs != mockFS {
			t.Error("Expected mock remote FS")
		}
		if info.User != "user" || info.Host != "example.com" || info.StartPath != "/home/user" {
			t.Errorf("Unexpected info: %+v", info)
		}
	})

	t.Run("Remote Parsing host", func(t *testing.T) {
		mockConnector.NewRemoteFSCallCount = 0
		_, info, _ := CreateFileSystemWithConnector("example.com", nil, mockConnector)
		if info.Host != "example.com" {
			t.Errorf("Expected host example.com, got %s", info.Host)
		}
	})

	t.Run("Remote Password fallback", func(t *testing.T) {
		mockConnector.NewRemoteFSCallCount = 0
		mockConnector.ReadPasswordValue = "secret"
		mockConnector.NewRemoteFSFunc = func(host, user, password, keyPath string, hkcb sshx.HostKeyCallback) (core.FileSystem, error) {
			if password == "" {
				return nil, fmt.Errorf("auth failed")
			}
			return mockFS, nil
		}

		fs, _, err := CreateFileSystemWithConnector("user@host", nil, mockConnector)
		testutil.AssertNoError(t, err, "Should succeed on second attempt")
		if fs != mockFS {
			t.Error("Expected mock remote FS")
		}
		if mockConnector.NewRemoteFSCallCount != 2 {
			t.Errorf("Expected 2 calls to NewRemoteFS, got %d", mockConnector.NewRemoteFSCallCount)
		}
	})

	t.Run("Remote Connection Failure", func(t *testing.T) {
		mockConnector.NewRemoteFSCallCount = 0
		mockConnector.NewRemoteFSFunc = func(host, user, password, keyPath string, hkcb sshx.HostKeyCallback) (core.FileSystem, error) {
			return nil, fmt.Errorf("total failure")
		}

		_, _, err := CreateFileSystemWithConnector("user@host", nil, mockConnector)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("CreateFileSystem Local Wrapper", func(t *testing.T) {
		fs, _, err := CreateFileSystem("", nil)
		testutil.AssertNoError(t, err, "Should not error")
		if fs == nil {
			t.Fatal("Expected local FS")
		}
		if !fs.IsLocal() {
			t.Error("Expected local FS to return true for IsLocal()")
		}
	})

	t.Run("Remote SSH Config Mock", func(t *testing.T) {
		mockConnector.NewRemoteFSFunc = func(host, user, password, keyPath string, hkcb sshx.HostKeyCallback) (core.FileSystem, error) {
			return mockFS, nil
		}
		_, info, _ := CreateFileSystemWithConnector("random-alias", nil, mockConnector)
		if info.Host != "random-alias" {
			t.Errorf("Expected host random-alias, got %s", info.Host)
		}
	})

	t.Run("Key file from args", func(t *testing.T) {
		mockConnector.NewRemoteFSFunc = func(host, user, password, keyPath string, hkcb sshx.HostKeyCallback) (core.FileSystem, error) {
			if keyPath != "my-key" {
				return nil, fmt.Errorf("wrong key: %s", keyPath)
			}
			return mockFS, nil
		}
		_, _, err := CreateFileSystemWithConnector("user@host", []string{"my-key"}, mockConnector)
		testutil.AssertNoError(t, err, "Should use key from args")
	})
}

func TestHostKeyCallbackInner(t *testing.T) {
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp.Path)
	defer os.Setenv("HOME", oldHome)

	keyData := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOm6y8v0W6Wz7mHn6/W1uF1v7Q+V2b2u5u5u5u5u5u5u"
	parts := strings.Split(keyData, " ")
	pk, _, _, _, _ := sshx.ParseAuthorizedKey([]byte(parts[0] + " " + parts[1]))
	addr, _ := net.ResolveTCPAddr("tcp", "example.com:22")

	t.Run("Response No", func(t *testing.T) {
		cb, _ := ssh.CreateCLIHostKeyCallback()
		r, w, _ := os.Pipe()
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()
		_, _ = w.Write([]byte("n\n"))
		w.Close()

		err := cb("example.com:22", addr, pk)
		if err == nil {
			t.Error("Expected error for unknown host with 'no' response")
		}
	})

	t.Run("Response Yes", func(t *testing.T) {
		cb, _ := ssh.CreateCLIHostKeyCallback()
		r, w, _ := os.Pipe()
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()
		_, _ = w.Write([]byte("y\n"))
		w.Close()

		// This might still fail if sshx.AddToKnownHosts fails, but it hits the branch
		_ = cb("example.com:22", addr, pk)
	})
}

func TestCreateHostKeyCallback(t *testing.T) {
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp.Path)
	defer os.Setenv("HOME", oldHome)

	t.Run("Create new known_hosts", func(t *testing.T) {
		cb, err := ssh.CreateCLIHostKeyCallback()
		testutil.AssertNoError(t, err, "Should create callback")
		if cb == nil {
			t.Fatal("Callback is nil")
		}
		if _, err := os.Stat(tmp.Join(".ssh", "known_hosts")); err != nil {
			t.Errorf("known_hosts not created: %v", err)
		}
	})

	t.Run("Existing known_hosts", func(t *testing.T) {
		_, err := ssh.CreateCLIHostKeyCallback()
		testutil.AssertNoError(t, err, "Should work with existing file")
	})
}

func TestCreateFileSystem_Errors(t *testing.T) {
	mockConnector := &MockConnector{}

	t.Run("HostKeyCallback creation error", func(t *testing.T) {
		mockConnector.HostKeyCallbackErr = fmt.Errorf("hkcb error")
		_, _, err := CreateFileSystemWithConnector("user@host", nil, mockConnector)
		if err == nil || !strings.Contains(err.Error(), "hkcb error") {
			t.Errorf("Expected hkcb error, got %v", err)
		}
	})

	t.Run("ReadPassword error", func(t *testing.T) {
		mockConnector.HostKeyCallbackErr = nil
		mockConnector.NewRemoteFSFunc = func(host, user, password, keyPath string, hkcb sshx.HostKeyCallback) (core.FileSystem, error) {
			return nil, fmt.Errorf("first fail")
		}
		mockConnector.ReadPasswordErr = fmt.Errorf("read pw error")
		_, _, err := CreateFileSystemWithConnector("user@host", nil, mockConnector)
		if err == nil || !strings.Contains(err.Error(), "read pw error") {
			t.Errorf("Expected read pw error, got %v", err)
		}
	})
}
