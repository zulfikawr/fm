package state

import "time"

// MessageState holds status message state
type MessageState struct {
	Text  string    // Current status message
	Time  time.Time // Time when message was set
	Error error     // Last error (if any)
}
