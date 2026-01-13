package actions

import (
	"errors"
	tuitestutil "fm/internal/tui/testutil"
	"testing"
)

func TestLogInfo(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	cmd := LogInfo(m, "test message")
	if cmd == nil {
		t.Error("Expected SetMsg command")
	}
}

func TestLogError(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	err := errors.New("test error")
	cmd := LogError(m, err, "context")
	if cmd == nil {
		t.Error("Expected SetMsg command")
	}
	if m.Message.Error == nil {
		t.Error("Expected error to be set in model")
	}
}
