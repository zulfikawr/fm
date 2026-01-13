package actions

import (
	"errors"
	"testing"

	tuitestutil "fm/internal/tui/testutil"
)

func TestErrorHandler(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Test User Error through LogError
	userErr := errors.New("user mistake")
	cmd := LogError(m, userErr, "context")
	if cmd == nil {
		t.Error("Expected command for user error")
	}

	// Test System Error through LogError
	sysErr := errors.New("system failure")
	cmd = LogError(m, sysErr, "critical context")
	if cmd == nil {
		t.Error("Expected command for system error")
	}
}
