package ssh

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	sshx "golang.org/x/crypto/ssh"
)

func TestResolveRemote(t *testing.T) {
	t.Run("user@host", func(t *testing.T) {
		details := ResolveRemote("user@example.com")
		if details.User != "user" || details.Host != "example.com" || details.StartPath != "" {
			t.Errorf("Unexpected details: %+v", details)
		}
	})

	t.Run("user@host:path", func(t *testing.T) {
		details := ResolveRemote("user@example.com:/home/user")
		if details.User != "user" || details.Host != "example.com" || details.StartPath != "/home/user" {
			t.Errorf("Unexpected details: %+v", details)
		}
	})

	t.Run("host only", func(t *testing.T) {
		details := ResolveRemote("example.com")
		if details.User != "" || details.Host != "example.com" || details.StartPath != "" {
			t.Errorf("Unexpected details: %+v", details)
		}
	})

	t.Run("SSH config alias", func(t *testing.T) {
		// Mock .ssh/config
		tmpHome := testutil.TempDir(t)
		oldHome := os.Getenv("HOME")
		testutil.AssertNoError(t, os.Setenv("HOME", tmpHome), "set home")
		defer func() {
			if err := os.Setenv("HOME", oldHome); err != nil {
				t.Errorf("failed to restore HOME: %v", err)
			}
		}()

		sshDir := filepath.Join(tmpHome, ".ssh")
		configPath := filepath.Join(sshDir, "config")

		err := os.MkdirAll(sshDir, 0700)
		testutil.AssertFatalError(t, err, "create .ssh dir")

		content := `Host myalias
  HostName realhost.com
  User dev
  IdentityFile ~/.ssh/id_rsa
`
		err = os.WriteFile(configPath, []byte(content), 0600)
		testutil.AssertFatalError(t, err, "write mock config")

		details := ResolveRemote("myalias")
		if details.User != "dev" || details.Host != "realhost.com" || details.KeyPath == "" {
			t.Errorf("Unexpected details for alias: %+v", details)
		}
	})
}

func TestAddToKnownHosts(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	oldHome := os.Getenv("HOME")
	testutil.AssertNoError(t, os.Setenv("HOME", tmpDir), "set home")
	defer func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Errorf("failed to restore HOME: %v", err)
		}
	}()

	// Ensure .ssh dir exists
	err := os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)
	testutil.AssertNoError(t, err, "create .ssh dir")

	keyData := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOm6y8v0W6Wz7mHn6/W1uF1v7Q+V2b2u5u5u5u5u5u5u"
	parts := strings.Split(keyData, " ")
	pk, out, options, rest, err := sshx.ParseAuthorizedKey([]byte(parts[0] + " " + parts[1]))
	if err != nil {
		t.Fatalf("parse key failed (out: %v, options: %v, rest: %v): %v", out, options, rest, err)
	}

	addr, err := net.ResolveTCPAddr("tcp", "example.com:22")
	testutil.AssertNoError(t, err, "resolve addr")

	err = AddToKnownHosts("example.com", addr, pk)
	testutil.AssertNoError(t, err, "AddToKnownHosts should succeed")

	// Check if file was created
	knownHostsPath := filepath.Join(tmpDir, ".ssh", "known_hosts")
	if info, err := os.Stat(knownHostsPath); err != nil {
		t.Errorf("known_hosts file not created (info: %+v): %v", info, err)
	}
}

func TestGetHostKeyCallback(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	oldHome := os.Getenv("HOME")
	testutil.AssertNoError(t, os.Setenv("HOME", tmpDir), "set home")
	defer func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Errorf("failed to restore HOME: %v", err)
		}
	}()

	// Ensure .ssh dir exists
	err := os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)
	testutil.AssertNoError(t, err, "create .ssh dir")

	askChan := make(chan *HostConfirmRequest, 1)
	cb, err := GetHostKeyCallback(askChan)
	testutil.AssertNoError(t, err, "GetHostKeyCallback should succeed")

	keyData := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOm6y8v0W6Wz7mHn6/W1uF1v7Q+V2b2u5u5u5u5u5u5u"
	parts := strings.Split(keyData, " ")
	pk, out, options, rest, err := sshx.ParseAuthorizedKey([]byte(parts[0] + " " + parts[1]))
	if err != nil {
		t.Fatalf("parse key failed (out: %v, options: %v, rest: %v): %v", out, options, rest, err)
	}

	addr, err := net.ResolveTCPAddr("tcp", "example.com:22")
	testutil.AssertNoError(t, err, "resolve addr")

	// Test new host
	go func() {
		req := <-askChan
		req.Resolve <- true
	}()

	err = cb("example.com:22", addr, pk)
	testutil.AssertNoError(t, err, "Callback should succeed when resolved")
}
