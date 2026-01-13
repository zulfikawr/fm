package commands

import (
	"testing"

	"fm/internal/files/core"
)

func TestWatcher(t *testing.T) {
	t.Run("ListenToProgress Command", func(t *testing.T) {
		progChan := make(chan core.Progress, 1)
		progChan <- core.Progress{Percent: 0.5, Label: "Testing"}
		cmd := ListenToProgress(progChan)
		msg := cmd()
		pMsg, ok := msg.(ProgressMsg)
		if !ok {
			t.Errorf("Expected ProgressMsg, got %T", msg)
		}
		if pMsg.Percent != 0.5 {
			t.Errorf("Expected 0.5, got %f", pMsg.Percent)
		}

		close(progChan)
		msg = cmd()
		if msg != nil {
			t.Error("Expected nil msg when channel closed")
		}
	})
}
