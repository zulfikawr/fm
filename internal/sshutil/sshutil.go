package sshutil

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConfig represents settings for an SSH connection.
type SSHConfig struct {
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

// HostNotFoundError is returned when a host is not in known_hosts.
type HostNotFoundError struct {
	Hostname string
	Remote   net.Addr
	Key      ssh.PublicKey
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

// GetHostKeyCallback returns an ssh.HostKeyCallback that handles known_hosts.
func GetHostKeyCallback(askChan chan<- *HostConfirmRequest) (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Ensure directory exists
	_ = os.MkdirAll(sshDir, 0700)
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		_ = os.WriteFile(knownHostsPath, []byte{}, 0600)
	}

	cb, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
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
	Key      ssh.PublicKey
	Resolve  chan bool
}

// AddToKnownHosts adds a host key to the known_hosts file.
func AddToKnownHosts(hostname string, remote net.Addr, key ssh.PublicKey) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")

	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Use both hostname and remote address for known_hosts
	entry := knownhosts.Line([]string{hostname, remote.String()}, key)
	_, err = f.WriteString(entry + "\n")
	return err
}
