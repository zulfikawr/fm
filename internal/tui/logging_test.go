package tui

import (
	"errors"
	"os"
	"testing"
	"time"

	"fm/internal/files"
	"fm/internal/logger"
)

func TestLogging(t *testing.T) {
	// Mock log file
	tmpLog, err := os.CreateTemp("", "fm-log-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpLog.Name())
	logger.SetLogPath(tmpLog.Name())

	m := NewModel(&files.LocalFS{}, ".")

	t.Run("LogError", func(t *testing.T) {
		err := errors.New("test error")
		cmd := m.LogError(err, "testing")
		if cmd == nil {
			t.Error("Expected non-nil command")
		}
		if m.msg == "" {
			t.Error("Expected model message to be set")
		}

		// nil error case
		cmd = m.LogError(nil, "testing")
		if cmd != nil {
			t.Error("Expected nil command for nil error")
		}
	})

	t.Run("LogInfo", func(t *testing.T) {
		cmd := m.LogInfo("test info")
		if cmd == nil {
			t.Error("Expected non-nil command")
		}
	})

	t.Run("ClearMsg", func(t *testing.T) {
		m.msg = "temporary message"
		m.msgTime = time.Now().Add(-MessageDisplayDuration - time.Second)
		m.ClearMsg()
		if m.msg != "" {
			t.Error("Expected message to be cleared")
		}

		m.msg = "recent message"
		m.msgTime = time.Now()
		m.ClearMsg()
		if m.msg == "" {
			t.Error("Expected message to NOT be cleared")
		}
	})
}
