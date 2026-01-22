package remote

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/sftp"
	sshx "golang.org/x/crypto/ssh"
)

func (fs *RemoteFS) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-ticker.C:
			fs.mu.RLock()
			conn := fs.conn
			fs.mu.RUnlock()

			if conn != nil {
				// Send a global request as a keep-alive heartbeat
				_, _, _ = conn.SendRequest("keepalive@openssh.com", true, nil)
			}
		}
	}
}

func (fs *RemoteFS) reconnect() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Close old connection if still open
	if fs.client != nil {
		fs.client.Close()
	}
	if fs.conn != nil {
		fs.conn.Close()
	}

	// Dial again
	conn, err := sshx.Dial("tcp", fs.opts.Address, fs.config)
	if err != nil {
		return fmt.Errorf("reconnect dial failed: %w", err)
	}

	client, err := sftp.NewClient(conn, sftp.UseConcurrentWrites(true))
	if err != nil {
		conn.Close()
		return fmt.Errorf("reconnect sftp failed: %w", err)
	}

	fs.conn = conn
	fs.client = client
	return nil
}

func (fs *RemoteFS) isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return err == io.EOF ||
		err == io.ErrUnexpectedEOF ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection lost") ||
		strings.Contains(errStr, "connection closed") ||
		strings.Contains(errStr, "use of closed network connection")
}

func (fs *RemoteFS) runWithRetry(fn func() error) error {
	// Try 1: Standard attempt
	fs.mu.RLock()
	client := fs.client
	fs.mu.RUnlock()

	if client == nil {
		if err := fs.reconnect(); err != nil {
			return err
		}
	}

	err := fn()
	if fs.isConnectionError(err) {
		// Try 2: Reconnect and retry once
		if recErr := fs.reconnect(); recErr == nil {
			return fn()
		}
	}
	return err
}
