package errors

import (
	"errors"
	"strings"
	"testing"

	"fm/internal/testutil"
)

func TestError(t *testing.T) {
	t.Run("New Error", func(t *testing.T) {
		err := New(ErrorTypeUser, "TestOp", "test message")
		testutil.AssertEqual(t, "TestOp: test message", err.Error(), "Error message should match")
		testutil.AssertEqual(t, ErrorTypeUser, err.Type, "Error type should match")
	})

	t.Run("Wrap Error", func(t *testing.T) {
		base := errors.New("base error")
		wrapped := Wrap(base, ErrorTypeSystem, "WrapOp")
		testutil.AssertEqual(t, "WrapOp: base error", wrapped.Error(), "Wrapped error message should match")
		testutil.AssertEqual(t, base, wrapped.Unwrap(), "Unwrap should return base error")
	})

	t.Run("UserMessage", func(t *testing.T) {
		err := UserError("Op", "specific msg")
		testutil.AssertEqual(t, "specific msg", err.UserMessage(), "User message should be the specific message")

		sysErr := SystemError("Op", errors.New("internal"))
		if !strings.Contains(sysErr.UserMessage(), "System Error: internal") {
			t.Errorf("System error should show internal message, got: %s", sysErr.UserMessage())
		}

		fatalErr := FatalError("Op", errors.New("internal"))
		if !strings.Contains(fatalErr.UserMessage(), "critical") {
			t.Error("Fatal error message should mention critical")
		}
	})

	t.Run("LogMessage and StackTrace", func(t *testing.T) {
		err := UserError("Op", "msg")
		logMsg := err.LogMessage()
		if !strings.Contains(logMsg, "[User]") || !strings.Contains(logMsg, "Op: msg") {
			t.Errorf("LogMessage failed: %s", logMsg)
		}

		stack := err.StackTrace()
		if !strings.Contains(stack, "Stack trace") {
			t.Error("StackTrace failed to capture stack")
		}
	})

	t.Run("Helper Checkers", func(t *testing.T) {
		err := UserError("Op", "msg")
		if !IsUserError(err) {
			t.Error("IsUserError failed")
		}
		if IsSystemError(err) {
			t.Error("IsSystemError should be false")
		}

		sys := SystemError("Op", errors.New("err"))
		if !IsSystemError(sys) {
			t.Error("IsSystemError failed")
		}

		fat := FatalError("Op", errors.New("err"))
		if !IsFatalError(fat) {
			t.Error("IsFatalError failed")
		}

		trn := TransientError("Op", "msg", 3)
		if !IsTransientError(trn) {
			t.Error("IsTransientError failed")
		}
	})

	t.Run("Retry Logic", func(t *testing.T) {
		r := DefaultRecovery()
		r.RetryDelay = 1 // fast tests

		attempts := 0
		err := r.Retry("Test", func() error {
			attempts++
			if attempts < 2 {
				return TransientError("Test", "retryable", 3)
			}
			return nil
		})

		if err != nil {
			t.Errorf("Retry failed: %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})
}

func TestHandler(t *testing.T) {
	called := false
	h := &Handler{
		OnUser: func(e *Error) error {
			called = true
			return e
		},
	}

	err := UserError("Op", "msg")
	h.Handle(err)
	testutil.AssertEqual(t, true, called, "OnUser handler should be called")

	// Test default wrap
	h2 := &Handler{
		OnSystem: func(e *Error) error {
			called = true
			return e
		},
	}
	called = false
	h2.Handle(errors.New("raw"))
	if !called {
		t.Error("Handler should wrap raw errors as System")
	}
}
