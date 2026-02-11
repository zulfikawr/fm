package ssh

import (
	"strings"

	"github.com/zulfikawr/fm/internal/logger"
)

// ResolveRemote parses a remote string and returns connection details.
// It checks SSH config first, then falls back to user@host:path format.
func ResolveRemote(remoteStr string) *RemoteConnectionDetails {
	details := &RemoteConnectionDetails{
		Host: remoteStr,
	}

	// Check SSH config first
	sshConfigs, err := ParseSSHConfig()
	if err != nil {
		logger.LogIfError(err, "Failed to parse SSH config")
		sshConfigs = make(map[string]*SSHConfig)
	}
	if cfg, ok := sshConfigs[remoteStr]; ok {
		if cfg.HostName != "" {
			details.Host = cfg.HostName
		}
		details.User = cfg.User
		details.KeyPath = cfg.IdentityFile
	} else if strings.Contains(remoteStr, "@") {
		parts := strings.SplitN(remoteStr, "@", 2)
		details.User = parts[0]
		rest := parts[1]

		if strings.Contains(rest, ":") {
			parts2 := strings.SplitN(rest, ":", 2)
			details.Host = parts2[0]
			details.StartPath = parts2[1]
		} else {
			details.Host = rest
		}
	}

	return details
}
