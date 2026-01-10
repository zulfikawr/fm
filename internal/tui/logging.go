package tui

import (
	"fmt"
	"time"

	"filemanager/internal/logger"
)

// LogError logs an error to the log file and updates the model's message.
func (m *Model) LogError(err error, context string) {
	if err == nil {
		return
	}
	msg := fmt.Sprintf("Error: %s: %v", context, err)
	logger.Error(msg)
	m.setMsg(msg)
}

// LogInfo logs an informational message.
func (m *Model) LogInfo(msg string) {
	logger.Info(msg)
	m.setMsg(msg)
}

// ClearMsg clears the current status message after a delay.
func (m *Model) ClearMsg() {
	if time.Since(m.msgTime) > 5*time.Second {
		m.msg = ""
	}
}
