package context

import "time"

// --- Message State ---

// Message represents a single status message
type Message struct {
	Text  string
	Time  time.Time
	IsErr bool
}

// MessageState holds status message state
type MessageState struct {
	Text  string    // Current status message (for backward compatibility)
	Time  time.Time // Time when message was set
	Error error     // Last error (if any)
	Stack []Message // Queue of messages to show
}

// Push adds a new message to the stack
func (ms *MessageState) Push(text string, isErr bool) {
	msg := Message{
		Text:  text,
		Time:  time.Now(),
		IsErr: isErr,
	}
	ms.Stack = append(ms.Stack, msg)
	ms.Text = text // Maintain compatibility
	ms.Time = msg.Time
}

// Pop removes the oldest message
func (ms *MessageState) Pop() {
	if len(ms.Stack) > 0 {
		ms.Stack = ms.Stack[1:]
		if len(ms.Stack) > 0 {
			ms.Text = ms.Stack[0].Text
			ms.Time = ms.Stack[0].Time
		} else {
			ms.Text = ""
		}
	}
}
