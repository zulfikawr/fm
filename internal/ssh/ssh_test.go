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
		home, _ := os.UserHomeDir()
		sshDir := filepath.Join(home, ".ssh")
		configPath := filepath.Join(sshDir, "config")

		err := os.MkdirAll(sshDir, 0700)
		if err != nil {
			t.Skip("Failed to create .ssh dir for test")
		}

		// Backup existing config if any
		var backup []byte
		if _, err := os.Stat(configPath); err == nil {
			backup, _ = os.ReadFile(configPath)
		}

		content := `Host myalias
  HostName realhost.com
  User dev
  IdentityFile ~/.ssh/id_rsa
`
		err = os.WriteFile(configPath, []byte(content), 0600)
		if err != nil {
			t.Skip("Failed to write mock config")
		}
		defer func() {
			if backup != nil {
				_ = os.WriteFile(configPath, backup, 0600)
			} else {
				_ = os.Remove(configPath)
			}
		}()

		details := ResolveRemote("myalias")
		if details.User != "dev" || details.Host != "realhost.com" || details.KeyPath == "" {
			t.Errorf("Unexpected details for alias: %+v", details)
		}
	})
}

func TestAddToKnownHosts(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	// Ensure .ssh dir exists
	_ = os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)

	keyData := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOm6y8v0W6Wz7mHn6/W1uF1v7Q+V2b2u5u5u5u5u5u5u"
	parts := strings.Split(keyData, " ")
	pk, _, _, _, _ := sshx.ParseAuthorizedKey([]byte(parts[0] + " " + parts[1]))
	addr, _ := net.ResolveTCPAddr("tcp", "example.com:22")

	err := AddToKnownHosts("example.com", addr, pk)
	testutil.AssertNoError(t, err, "AddToKnownHosts should succeed")

	// Check if file was created
	knownHostsPath := filepath.Join(tmpDir, ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); err != nil {
		t.Errorf("known_hosts file not created: %v", err)
	}
}

func TestGetHostKeyCallback(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	_ = os.Setenv("HOME", tmpDir)

	// Ensure .ssh dir exists
	_ = os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)

	askChan := make(chan *HostConfirmRequest, 1)
	cb, err := GetHostKeyCallback(askChan)
	testutil.AssertNoError(t, err, "GetHostKeyCallback should succeed")

	keyData := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOm6y8v0W6Wz7mHn6/W1uF1v7Q+V2b2u5u5u5u5u5u5u"
	parts := strings.Split(keyData, " ")
	pk, _, _, _, _ := sshx.ParseAuthorizedKey([]byte(parts[0] + " " + parts[1]))
	addr, _ := net.ResolveTCPAddr("tcp", "example.com:22")

	// Test new host
	go func() {
		req := <-askChan
		req.Resolve <- true
	}()

	err = cb("example.com:22", addr, pk)
	testutil.AssertNoError(t, err, "Callback should succeed when resolved")
}
