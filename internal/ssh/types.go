// Package ssh provides SSH/SFTP connection utilities including authentication,
// known_hosts management, and SSH config file parsing.
package ssh

import (
	"fmt"
	"net"

	sshx "golang.org/x/crypto/ssh"
)

// SSHConfig encapsulates SSH connection parameters.
type SSHConfig struct {
	Address         string
	User            string
	Password        string
	KeyPath         string
	HostKeyCallback sshx.HostKeyCallback

	// Internal use for SSH config file parsing if needed
	HostName     string
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
