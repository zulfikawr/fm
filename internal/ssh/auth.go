package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulfikawr/fm/internal/logger"

	sshx "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

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
	defer logger.CloseAndLog(f, "known_hosts file during append")

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
			_, err := fmt.Scanln(&response)
			logger.LogIfError(err, "ssh: failed to read user response for host authenticity")
			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				// Add to known_hosts
				return AddToKnownHosts(hostname, remote, key)
			}
		}
		return err
	}, nil
}
