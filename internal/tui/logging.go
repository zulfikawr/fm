package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"fm/internal/logger"
)

// LogError logs an error to the log file and updates the model's message.
func (m *Model) LogError(err error, context string) tea.Cmd {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf("Error: %s: %v", context, err)
	logger.Error(msg)
	return m.setMsg(msg)
}

// LogInfo logs an informational message.
func (m *Model) LogInfo(msg string) tea.Cmd {
	logger.Info(msg)
	return m.setMsg(msg)
}

// ClearMsg clears the current status message after a delay.
func (m *Model) ClearMsg() {
	if time.Since(m.msgTime) > MessageDisplayDuration {
		m.msg = ""
	}
}
