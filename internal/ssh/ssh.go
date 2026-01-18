package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"fm/internal/logger"

	sshx "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConfig represents settings for an SSH connection.
type SSHConfig struct {
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

// RemoteConnectionDetails holds information needed to establish an SSH connection.
type RemoteConnectionDetails struct {
	Host      string
	User      string
	KeyPath   string
	StartPath string
}

// ResolveRemote parses a remote string and returns connection details.
// It checks SSH config first, then falls back to user@host:path format.
func ResolveRemote(remoteStr string) *RemoteConnectionDetails {
	details := &RemoteConnectionDetails{
		Host: remoteStr,
	}

	// Check SSH config first
	sshConfigs, _ := ParseSSHConfig()
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

// HostNotFoundError is returned when a host is not in known_hosts.
type HostNotFoundError struct {
	Hostname string
	Remote   net.Addr
	Key      sshx.PublicKey
}

func (e *HostNotFoundError) Error() string {
	return fmt.Sprintf("host not found in known_hosts: %s", e.Hostname)
}

// ParseSSHConfig parses the ~/.ssh/config file.
func ParseSSHConfig() (map[string]*SSHConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ssh", "config")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*SSHConfig), nil
		}
		return nil, err
	}
	defer file.Close()

	configs := make(map[string]*SSHConfig)
	var currentHost []string
	var currentCfg *SSHConfig

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "host":
			// Save previous config if any
			if currentCfg != nil {
				for _, h := range currentHost {
					// We don't handle wildcards well here, but it's a start
					if h != "*" {
						configs[h] = currentCfg
					}
				}
			}
			currentHost = parts[1:]
			currentCfg = &SSHConfig{}
		case "hostname":
			if currentCfg != nil {
				currentCfg.HostName = value
			}
		case "user":
			if currentCfg != nil {
				currentCfg.User = value
			}
		case "port":
			if currentCfg != nil {
				currentCfg.Port = value
			}
		case "identityfile":
			if currentCfg != nil {
				// Expand ~ if present
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(home, value[1:])
				}
				currentCfg.IdentityFile = value
			}
		}
	}

	// Save last config
	if currentCfg != nil {
		for _, h := range currentHost {
			if h != "*" {
				configs[h] = currentCfg
			}
		}
	}

	return configs, nil
}

// GetHostKeyCallback returns an sshx.HostKeyCallback that handles known_hosts.
func GetHostKeyCallback(askChan chan<- *HostConfirmRequest) (sshx.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Ensure directory exists
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		logger.Warnf("Failed to create SSH directory: %v", err)
	}
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		if err := os.WriteFile(knownHostsPath, []byte{}, 0o600); err != nil {
			logger.Warnf("Failed to create known_hosts file: %v", err)
		}
	}

	cb, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key sshx.PublicKey) error {
		err := cb(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			if askChan != nil {
				resolve := make(chan bool)
				askChan <- &HostConfirmRequest{
					Hostname: hostname,
					Remote:   remote,
					Key:      key,
					Resolve:  resolve,
				}
				if <-resolve {
					return AddToKnownHosts(hostname, remote, key)
				}
				return &HostNotFoundError{Hostname: hostname, Remote: remote, Key: key}
			}
			return &HostNotFoundError{Hostname: hostname, Remote: remote, Key: key}
		}
		return err
	}, nil
}

// HostConfirmRequest represents a request to the user to confirm a host key.
type HostConfirmRequest struct {
	Hostname string
	Remote   net.Addr
	Key      sshx.PublicKey
	Resolve  chan bool
}

// AddToKnownHosts adds a host key to the known_hosts file.
func AddToKnownHosts(hostname string, remote net.Addr, key sshx.PublicKey) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")

	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Use both hostname and remote address for known_hosts
	entry := knownhosts.Line([]string{hostname, remote.String()}, key)
	_, err = f.WriteString(entry + "\n")
	return err
}

// CreateCLIHostKeyCallback returns an sshx.HostKeyCallback that prompts the user on the CLI.
func CreateCLIHostKeyCallback() (sshx.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Ensure directory exists
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		logger.Warnf("Failed to create SSH directory: %v", err)
	}
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		if err := os.WriteFile(knownHostsPath, []byte{}, 0o600); err != nil {
			logger.Warnf("Failed to create known_hosts file: %v", err)
		}
	}

	cb, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key sshx.PublicKey) error {
		err := cb(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			// Host not found in known_hosts
			fmt.Printf("The authenticity of host '%s' can't be established.\n", hostname)
			fmt.Printf("%s key fingerprint is %s.\n", key.Type(), sshx.FingerprintSHA256(key))
			fmt.Print("Are you sure you want to continue connecting [y] Yes [n] No? ")

			var response string
			_, _ = fmt.Scanln(&response)
			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				// Add to known_hosts
				return AddToKnownHosts(hostname, remote, key)
			}
		}
		return err
	}, nil
}
