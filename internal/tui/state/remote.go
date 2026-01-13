package state

import "fm/internal/sshutil"

// RemoteState holds remote connection state
type RemoteState struct {
	Host            string                           // For interactive remote connection
	User            string                           // For interactive remote connection
	HostConfirmChan chan *sshutil.HostConfirmRequest // Channel for host confirmation requests
	HostConfirmReq  *sshutil.HostConfirmRequest      // Current host confirmation request
}
