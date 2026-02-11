package remote

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/testutil"

	"github.com/pkg/sftp"
)

func TestRemoteFS_Helpers(t *testing.T) {
	fs := &RemoteFS{
		opts: ssh.SSHConfig{
			Address: "example.com:22",
			User:    "user",
		},
	}

	t.Run("Basic Helpers", func(t *testing.T) {
		testutil.AssertEqual(t, false, fs.IsLocal(), "IsLocal should be false")
		testutil.AssertEqual(t, "example.com:22", fs.Address(), "Address should match")
		testutil.AssertEqual(t, "user", fs.User(), "User should match")
		testutil.AssertEqual(t, "/", fs.Separator(), "Separator should be /")
	})

	t.Run("Path Helpers", func(t *testing.T) {
		testutil.AssertEqual(t, "file.txt", fs.Base("/a/b/file.txt"), "Base should work")
		testutil.AssertEqual(t, "/a/b", fs.Dir("/a/b/file.txt"), "Dir should work")
		testutil.AssertEqual(t, "/a/b/c", fs.Join("/a/b", "c"), "Join should work")
		testutil.AssertEqual(t, ".txt", fs.Ext("file.txt"), "Ext should work")
		testutil.AssertEqual(t, "/", fs.Clean("//"), "Clean should work")
	})

	t.Run("Rel", func(t *testing.T) {
		tests := []struct {
			base     string
			targ     string
			expected string
		}{
			{"/a/b", "/a/b/c/d", "c/d"},
			{"/a/b/c", "/a/b", ".."},
			{"/a/b", "/a/b", "."},
			{"/", "/a", "a"},
			{"/a", "/", ".."},
			{"/a/b/c", "/a/d/e", "../../d/e"},
			{"/a/b", "/a/b/c", "c"},
		}

		for _, tt := range tests {
			rel, err := fs.Rel(tt.base, tt.targ)
			testutil.AssertNoError(t, err, "Rel should succeed")
			testutil.AssertEqual(t, tt.expected, rel, "Rel should match")
		}
	})

	t.Run("isConnectionError", func(t *testing.T) {
		testutil.AssertEqual(t, true, fs.isConnectionError(errors.New("use of closed network connection")), "use of closed network connection should be connection error")
		testutil.AssertEqual(t, true, fs.isConnectionError(io.EOF), "EOF should be connection error")
		testutil.AssertEqual(t, false, fs.isConnectionError(os.ErrNotExist), "not exist should not be connection error")
	})
}

type rwCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

func TestRemoteFS_ClientMethods(t *testing.T) {
	// Pair 1: client to server
	c2sr, c2sw := io.Pipe()
	// Pair 2: server to client
	s2cr, s2cw := io.Pipe()

	serverRWC := &rwCloser{c2sr, s2cw, c2sr}
	clientRWC := &rwCloser{s2cr, c2sw, s2cr}

	// Start a server on the server end of the pipe
	server, err := sftp.NewServer(serverRWC)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := server.Serve(); err != nil && err != io.EOF {
			// server.Serve() always returns an error when the connection is closed
		}
	}()
	defer func() {
		if err := server.Close(); err != nil {
			t.Errorf("failed to close server: %v", err)
		}
	}()

	// Create a client on the client end of the pipe
	client, err := sftp.NewClientPipe(clientRWC, clientRWC)
	if err != nil {
		t.Fatal(err)
	}

	ictx, cancel := context.WithCancel(context.Background())
	fs := &RemoteFS{
		client: client,
		ctx:    ictx,
		cancel: cancel,
	}
	ctx := context.Background()

	tmpDir := testutil.TempDir(t)

	t.Run("ReadDir", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}

		entries, err := fs.ReadDir(ctx, tmpDir)
		testutil.AssertNoError(t, err, "ReadDir should succeed")
		testutil.AssertEqual(t, 1, len(entries), "Should find one file")
	})

	t.Run("Stat and Lstat", func(t *testing.T) {
		path := filepath.Join(tmpDir, "stat.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		info, err := fs.Stat(ctx, path)
		testutil.AssertNoError(t, err, "Stat should succeed")
		testutil.AssertEqual(t, "stat.txt", info.Name(), "Name should match")

		_, err = fs.Lstat(ctx, path)
		testutil.AssertNoError(t, err, "Lstat should succeed")
	})

	t.Run("Create and Open", func(t *testing.T) {
		path := filepath.Join(tmpDir, "create.txt")

		w, err := fs.Create(ctx, path)
		testutil.AssertNoError(t, err, "Create should succeed")
		if _, err := w.Write([]byte("data")); err != nil {
			t.Errorf("failed to write: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Errorf("failed to close writer: %v", err)
		}

		r, err := fs.Open(ctx, path)
		testutil.AssertNoError(t, err, "Open should succeed")
		data, err := io.ReadAll(r)
		testutil.AssertNoError(t, err, "ReadAll should succeed")
		if err := r.Close(); err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
		testutil.AssertEqual(t, "data", string(data), "Content should match")
	})

	t.Run("MkdirAll and Rename", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "remote_a/b")
		err := fs.MkdirAll(ctx, dir, 0755)
		testutil.AssertNoError(t, err, "MkdirAll should succeed")

		newDir := filepath.Join(tmpDir, "remote_c")
		err = fs.Rename(ctx, filepath.Join(tmpDir, "remote_a"), newDir)
		testutil.AssertNoError(t, err, "Rename should succeed")
	})

	t.Run("RemoveAll", func(t *testing.T) {
		path := filepath.Join(tmpDir, "to_delete.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		err := fs.RemoveAll(ctx, path)
		testutil.AssertNoError(t, err, "RemoveAll file should succeed")

		dir := filepath.Join(tmpDir, "dir_to_delete")
		err = fs.MkdirAll(ctx, dir, 0755)
		testutil.AssertNoError(t, err, "MkdirAll should succeed")
		if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		err = fs.RemoveAll(ctx, dir)
		testutil.AssertNoError(t, err, "RemoveAll dir should succeed")
	})

	t.Run("Chmod", func(t *testing.T) {
		path := filepath.Join(tmpDir, "chmod.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		err := fs.Chmod(ctx, path, 0644)
		testutil.AssertNoError(t, err, "Chmod should succeed")
	})

	t.Run("GetHomeDir", func(t *testing.T) {
		_, err := fs.GetHomeDir()
		testutil.AssertNoError(t, err, "GetHomeDir should succeed")
	})

	t.Run("Walk", func(t *testing.T) {
		walkDir := filepath.Join(tmpDir, "walk_test")
		if err := os.MkdirAll(filepath.Join(walkDir, "dir1"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(walkDir, "file1.txt"), []byte("1"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(walkDir, "dir1/file2.txt"), []byte("2"), 0644); err != nil {
			t.Fatal(err)
		}

		count := 0
		err := fs.Walk(ctx, walkDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info != nil && !info.IsDir() {
				count++
			}
			return nil
		})
		testutil.AssertNoError(t, err, "Walk should succeed")
		if count != 2 {
			t.Errorf("expected 2 files, got %d", count)
		}
	})

	t.Run("IsReadOnly", func(t *testing.T) {
		ro, err := fs.IsReadOnly(ctx, tmpDir)
		testutil.AssertNoError(t, err, "IsReadOnly should succeed")
		testutil.AssertEqual(t, false, ro, "tempDir should not be read-only")
	})

	t.Run("ReadDirEntries", func(t *testing.T) {
		path := filepath.Join(tmpDir, "entry.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		entries, err := fs.ReadDirEntries(ctx, tmpDir)
		testutil.AssertNoError(t, err, "ReadDirEntries should succeed")
		if len(entries) < 1 {
			t.Errorf("expected at least 1 entry")
		}
	})

	t.Run("Abs", func(t *testing.T) {
		abs, err := fs.Abs("/a/b")
		testutil.AssertNoError(t, err, "Abs with absolute path should succeed")
		testutil.AssertEqual(t, "/a/b", abs, "Abs path should match")

		abs, err = fs.Abs("relative")
		testutil.AssertNoError(t, err, "Abs with relative path should succeed")
		if !path.IsAbs(abs) {
			t.Errorf("expected absolute path, got %s", abs)
		}
	})

	t.Run("Preallocate", func(t *testing.T) {
		err := fs.Preallocate(ctx, "file.txt", 1024)
		testutil.AssertNoError(t, err, "Preallocate should be a no-op success")
	})

	t.Run("infoToDirEntry", func(t *testing.T) {
		info, err := os.Stat(os.Args[0])
		testutil.AssertNoError(t, err, "Stat current executable")
		entry := infoToDirEntry(info)
		testutil.AssertEqual(t, info.Name(), entry.Name(), "Name should match")
		testutil.AssertEqual(t, info.IsDir(), entry.IsDir(), "IsDir should match")
		testutil.AssertEqual(t, info.Mode().Type(), entry.Type(), "Type should match")
		inf, err := entry.Info()
		testutil.AssertNoError(t, err, "Info should succeed")
		testutil.AssertEqual(t, info.Name(), inf.Name(), "FileInfo name should match")
	})

	t.Run("Close", func(t *testing.T) {
		err := fs.Close()
		testutil.AssertNoError(t, err, "Close should succeed")
	})
}
