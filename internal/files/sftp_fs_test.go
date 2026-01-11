package files

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/pkg/sftp"
)

type pipeConn struct {
	io.Reader
	io.Writer
}

func (p *pipeConn) Close() error { return nil }

func TestSftpFS_Basic(t *testing.T) {
	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()
	serverConn := &pipeConn{serverReader, serverWriter}

	server, err := sftp.NewServer(serverConn)
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve()

	client, err := sftp.NewClientPipe(clientReader, clientWriter)
	if err != nil {
		t.Fatal(err)
	}

	fs := &SftpFS{
		client: client,
		conn:   nil,
	}

	t.Run("Separator", func(t *testing.T) {
		if fs.Separator() != "/" {
			t.Errorf("Expected /, got %s", fs.Separator())
		}
	})

	t.Run("IsLocal", func(t *testing.T) {
		if fs.IsLocal() {
			t.Error("SftpFS should not be local")
		}
	})

	t.Run("File Operations", func(t *testing.T) {
		err := fs.MkdirAll(context.Background(), "testdir/sub", 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		w, err := fs.Create(context.Background(), "testdir/sub/hello.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		w.Write([]byte("hello world"))
		w.Close()

		info, err := fs.Stat(context.Background(), "testdir/sub/hello.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if info.Size() != 11 {
			t.Errorf("Expected size 11, got %d", info.Size())
		}

		// Lstat
		_, err = fs.Lstat(context.Background(), "testdir/sub/hello.txt")
		if err != nil {
			t.Errorf("Lstat failed: %v", err)
		}

		// Chmod
		err = fs.Chmod(context.Background(), "testdir/sub/hello.txt", 0600)
		if err != nil {
			t.Errorf("Chmod failed: %v", err)
		}

		// Abs
		abs, _ := fs.Abs("rel/path")
		if !strings.HasSuffix(abs, "rel/path") {
			t.Errorf("Abs failed: %s", abs)
		}
		abs2, _ := fs.Abs("/abs/path")
		if abs2 != "/abs/path" {
			t.Errorf("Abs failed: %s", abs2)
		}

		entries, err := fs.ReadDir(context.Background(), "testdir/sub")
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}
		if len(entries) != 1 || entries[0].Name() != "hello.txt" {
			t.Errorf("ReadDir mismatch: %+v", entries)
		}

		r, err := fs.Open(context.Background(), "testdir/sub/hello.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		content, _ := io.ReadAll(r)
		r.Close()
		if string(content) != "hello world" {
			t.Errorf("Expected hello world, got %s", string(content))
		}

		err = fs.Rename(context.Background(), "testdir/sub/hello.txt", "testdir/sub/hi.txt")
		if err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		err = fs.RemoveAll(context.Background(), "testdir")
		if err != nil {
			t.Fatalf("RemoveAll failed: %v", err)
		}
		_, err = fs.Stat(context.Background(), "testdir")
		if err == nil {
			t.Error("testdir should be gone")
		}
	})
}

func TestSftpFS_GetGitStatus_Nil(t *testing.T) {
	fs := &SftpFS{conn: nil}
	statuses, branch := fs.GetGitStatus(context.Background(), ".")
	if statuses != nil || branch != "" {
		t.Error("Expected nil/empty for nil connection")
	}
}

func TestSftpFS_Helpers(t *testing.T) {
	fs := &SftpFS{}

	t.Run("Join", func(t *testing.T) {
		got := fs.Join("a", "b", "c")
		if got != "a/b/c" {
			t.Errorf("Expected a/b/c, got %s", got)
		}
	})

	t.Run("Dir and Base", func(t *testing.T) {
		if fs.Dir("/a/b/c") != "/a/b" {
			t.Errorf("Dir failed: %s", fs.Dir("/a/b/c"))
		}
		if fs.Base("/a/b/c") != "c" {
			t.Errorf("Base failed: %s", fs.Base("/a/b/c"))
		}
	})
}

func TestRemoveAll_File(t *testing.T) {
	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()
	serverConn := &pipeConn{serverReader, serverWriter}
	server, _ := sftp.NewServer(serverConn)
	go server.Serve()
	client, _ := sftp.NewClientPipe(clientReader, clientWriter)
	fs := &SftpFS{client: client}

	w, _ := fs.Create(context.Background(), "single_file.txt")
	w.Write([]byte("test"))
	w.Close()

	if err := fs.RemoveAll(context.Background(), "single_file.txt"); err != nil {
		t.Errorf("RemoveAll on file failed: %v", err)
	}
}

func TestRemoveAll_NonExistent(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			// Expected panic or handle if s.client is nil
		}
	}()
	fs := &SftpFS{}
	fs.RemoveAll(context.Background(), "nonexistent")
}
