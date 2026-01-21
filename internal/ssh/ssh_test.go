package ssh

import (
	"os"
	"path/filepath"
	"testing"
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
