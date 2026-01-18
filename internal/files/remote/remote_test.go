package remote

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"fm/internal/files/core"
	"fm/internal/testutil"

	"github.com/pkg/sftp"
)

func TestRemoteFS_Helpers(t *testing.T) {
	fs := &RemoteFS{
		address: "example.com:22",
		user:    "user",
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
	go func() { _ = server.Serve() }()
	defer server.Close()

	// Create a client on the client end of the pipe
	client, err := sftp.NewClientPipe(clientRWC, clientRWC)
	if err != nil {
		t.Fatal(err)
	}

	ictx, cancel := context.WithCancel(context.Background())
	fs := &RemoteFS{
		client: client,
		cache:  core.NewMetadataCache(time.Second),
		ctx:    ictx,
		cancel: cancel,
	}
	ctx := context.Background()

	t.Run("ReadDir", func(t *testing.T) {
		// Since we use a default server, it will use the local filesystem
		// We use a temp dir for testing
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		tmp.WriteFile("test.txt", "hello")

		entries, err := fs.ReadDir(ctx, tmp.Path)
		testutil.AssertNoError(t, err, "ReadDir should succeed")
		testutil.AssertEqual(t, 1, len(entries), "Should find one file")
	})

	t.Run("Stat and Lstat", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		path := tmp.WriteFile("stat.txt", "")

		info, err := fs.Stat(ctx, path)
		testutil.AssertNoError(t, err, "Stat should succeed")
		testutil.AssertEqual(t, "stat.txt", info.Name(), "Name should match")

		_, err = fs.Lstat(ctx, path)
		testutil.AssertNoError(t, err, "Lstat should succeed")
	})

	t.Run("Create and Open", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		path := tmp.Join("create.txt")

		w, err := fs.Create(ctx, path)
		testutil.AssertNoError(t, err, "Create should succeed")
		_, _ = w.Write([]byte("data"))
		w.Close()

		r, err := fs.Open(ctx, path)
		testutil.AssertNoError(t, err, "Open should succeed")
		data, _ := io.ReadAll(r)
		r.Close()
		testutil.AssertEqual(t, "data", string(data), "Content should match")
	})

	t.Run("MkdirAll and Rename", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		dir := tmp.Join("a/b")
		err := fs.MkdirAll(ctx, dir, 0755)
		testutil.AssertNoError(t, err, "MkdirAll should succeed")

		newDir := tmp.Join("c")
		err = fs.Rename(ctx, dir, newDir)
		testutil.AssertNoError(t, err, "Rename should succeed")
	})

	t.Run("RemoveAll", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		path := tmp.WriteFile("to_delete.txt", "")

		err := fs.RemoveAll(ctx, path)
		testutil.AssertNoError(t, err, "RemoveAll file should succeed")

		dir := tmp.Join("dir_to_delete")
		_ = fs.MkdirAll(ctx, dir, 0755)
		tmp.WriteFile("dir_to_delete/file.txt", "")
		err = fs.RemoveAll(ctx, dir)
		testutil.AssertNoError(t, err, "RemoveAll dir should succeed")
	})

	t.Run("Chmod", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		path := tmp.WriteFile("chmod.txt", "")
		err := fs.Chmod(ctx, path, 0644)
		testutil.AssertNoError(t, err, "Chmod should succeed")
	})

	t.Run("GetHomeDir", func(t *testing.T) {
		_, err := fs.GetHomeDir()
		testutil.AssertNoError(t, err, "GetHomeDir should succeed")
	})

	t.Run("Walk", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		tmp.WriteFile("file1.txt", "1")
		tmp.WriteFile("dir1/file2.txt", "2")

		// Re-initialize FS to avoid any weirdness
		count := 0
		err := fs.Walk(ctx, tmp.Path, func(path string, info os.FileInfo, err error) error {
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
			// Try a simpler walk if parallel walk failed to find items
			t.Errorf("expected 2 files, got %d", count)
		}
	})

	t.Run("IsReadOnly", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		ro, err := fs.IsReadOnly(ctx, tmp.Path)
		testutil.AssertNoError(t, err, "IsReadOnly should succeed")
		testutil.AssertEqual(t, false, ro, "tempDir should not be read-only")
	})

	t.Run("ReadDirEntries", func(t *testing.T) {
		tmp := testutil.NewTempFolder(t)
		defer tmp.Cleanup()
		tmp.WriteFile("entry.txt", "")
		entries, err := fs.ReadDirEntries(ctx, tmp.Path)
		testutil.AssertNoError(t, err, "ReadDirEntries should succeed")
		if len(entries) < 1 {
			t.Errorf("expected at least 1 entry")
		}
	})

	t.Run("Abs", func(t *testing.T) {
		// Abs with absolute path
		abs, err := fs.Abs("/a/b")
		testutil.AssertNoError(t, err, "Abs with absolute path should succeed")
		testutil.AssertEqual(t, "/a/b", abs, "Abs path should match")

		// Abs with relative path
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
		info, _ := os.Stat(os.Args[0]) // Get some FileInfo
		entry := infoToDirEntry(info)
		testutil.AssertEqual(t, info.Name(), entry.Name(), "Name should match")
		testutil.AssertEqual(t, info.IsDir(), entry.IsDir(), "IsDir should match")
		testutil.AssertEqual(t, info.Mode().Type(), entry.Type(), "Type should match")
		inf, err := entry.Info()
		testutil.AssertNoError(t, err, "Info should succeed")
		testutil.AssertEqual(t, info, inf, "FileInfo should match")
	})

	t.Run("Close", func(t *testing.T) {
		err := fs.Close()
		testutil.AssertNoError(t, err, "Close should succeed")
	})
}
