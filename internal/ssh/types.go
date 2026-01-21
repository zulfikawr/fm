package ssh

import (
	"fmt"
	"net"

	sshx "golang.org/x/crypto/ssh"
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

// HostNotFoundError is returned when a host is not in known_hosts.
type HostNotFoundError struct {
	Hostname string
	Remote   net.Addr
	Key      sshx.PublicKey
}

func (e *HostNotFoundError) Error() string {
	return fmt.Sprintf("host not found in known_hosts: %s", e.Hostname)
}

// HostConfirmRequest represents a request to the user to confirm a host key.
type HostConfirmRequest struct {
	Hostname string
	Remote   net.Addr
	Key      sshx.PublicKey
	Resolve  chan bool
}
