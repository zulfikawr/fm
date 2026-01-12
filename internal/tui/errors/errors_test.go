package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	err := New(ErrorTypeUser, "test operation", "test message")

	if err.Type != ErrorTypeUser {
		t.Errorf("Expected type User, got %s", err.Type)
	}
	if err.Operation != "test operation" {
		t.Errorf("Expected operation 'test operation', got '%s'", err.Operation)
	}
	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}
	if len(err.Stack) == 0 {
		t.Error("Expected stack trace to be captured")
	}
	if err.Context == nil {
		t.Error("Expected context map to be initialized")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(originalErr, ErrorTypeSystem, "wrapping operation")

	if err.Type != ErrorTypeSystem {
		t.Errorf("Expected type System, got %s", err.Type)
	}
	if err.Operation != "wrapping operation" {
		t.Errorf("Expected operation 'wrapping operation', got '%s'", err.Operation)
	}
	if err.Cause != originalErr {
		t.Error("Expected cause to be original error")
	}
	if err.Unwrap() != originalErr {
		t.Error("Expected Unwrap() to return original error")
	}
}

func TestWrapNil(t *testing.T) {
	err := Wrap(nil, ErrorTypeUser, "operation")
	if err != nil {
		t.Error("Expected nil when wrapping nil error")
	}
}

func TestWrapTUIError(t *testing.T) {
	original := New(ErrorTypeUser, "original", "message")
	wrapped := Wrap(original, ErrorTypeSystem, "new operation")

	// Should return the same error with updated operation and type
	if wrapped.Message != "message" {
		t.Error("Expected original message to be preserved")
	}
	if wrapped.Operation != "original" {
		t.Error("Expected original operation to be preserved (non-empty)")
	}
	if wrapped.Type != ErrorTypeSystem {
		t.Error("Expected type to be updated")
	}
}

func TestWithCode(t *testing.T) {
	err := New(ErrorTypeUser, "op", "msg").WithCode("ERR_001")

	if err.Code != "ERR_001" {
		t.Errorf("Expected code 'ERR_001', got '%s'", err.Code)
	}
}

func TestWithContext(t *testing.T) {
	err := New(ErrorTypeUser, "op", "msg").
		WithContext("path", "/test/path").
		WithContext("user", "testuser")

	if err.Context["path"] != "/test/path" {
		t.Error("Expected context 'path' to be set")
	}
	if err.Context["user"] != "testuser" {
		t.Error("Expected context 'user' to be set")
	}
}

func TestWithRetry(t *testing.T) {
	err := New(ErrorTypeTransient, "op", "msg").WithRetry(5)

	if !err.Retryable {
		t.Error("Expected error to be retryable")
	}
	if err.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", err.MaxRetries)
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "with operation",
			err:      New(ErrorTypeUser, "read file", "file not found"),
			expected: "read file: file not found",
		},
		{
			name:     "without operation",
			err:      &Error{Message: "generic error"},
			expected: "generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		contains string
	}{
		{
			name:     "user error",
			err:      UserError("op", "Permission denied"),
			contains: "Permission denied",
		},
		{
			name:     "system error",
			err:      SystemError("op", errors.New("internal")),
			contains: "system error",
		},
		{
			name:     "fatal error",
			err:      FatalError("op", errors.New("critical")),
			contains: "critical error",
		},
		{
			name:     "transient error without retry",
			err:      New(ErrorTypeTransient, "op", "timeout"),
			contains: "timeout",
		},
		{
			name:     "transient error with retry",
			err:      TransientError("op", "timeout", 3),
			contains: "Retrying",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.UserMessage()
			if !strings.Contains(strings.ToLower(msg), strings.ToLower(tt.contains)) {
				t.Errorf("Expected message to contain '%s', got '%s'", tt.contains, msg)
			}
		})
	}
}

func TestLogMessage(t *testing.T) {
	err := New(ErrorTypeSystem, "test op", "test error").
		WithCode("ERR_001").
		WithContext("path", "/test")

	logMsg := err.LogMessage()

	if !strings.Contains(logMsg, "System") {
		t.Error("Expected log message to contain error type")
	}
	if !strings.Contains(logMsg, "test op") {
		t.Error("Expected log message to contain operation")
	}
	if !strings.Contains(logMsg, "ERR_001") {
		t.Error("Expected log message to contain error code")
	}
	if !strings.Contains(logMsg, "path") {
		t.Error("Expected log message to contain context")
	}
}

func TestLogMessageWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := Wrap(cause, ErrorTypeSystem, "operation")

	logMsg := err.LogMessage()

	if !strings.Contains(logMsg, "Cause:") {
		t.Error("Expected log message to contain cause")
	}
	if !strings.Contains(logMsg, "underlying error") {
		t.Error("Expected log message to contain cause message")
	}
}

func TestStackTrace(t *testing.T) {
	err := New(ErrorTypeUser, "op", "msg")

	trace := err.StackTrace()

	if trace == "" {
		t.Error("Expected non-empty stack trace")
	}
	if !strings.Contains(trace, "Stack trace:") {
		t.Error("Expected stack trace header")
	}
	if !strings.Contains(trace, "errors_test.go") {
		t.Error("Expected test file in stack trace")
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected bool
	}{
		{
			name:     "retryable with attempts left",
			err:      TransientError("op", "msg", 3),
			expected: true,
		},
		{
			name: "retryable with no attempts left",
			err: func() *Error {
				e := TransientError("op", "msg", 3)
				e.RetryCount = 3
				return e
			}(),
			expected: false,
		},
		{
			name:     "not retryable",
			err:      UserError("op", "msg"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.ShouldRetry() != tt.expected {
				t.Errorf("Expected ShouldRetry() to be %v, got %v", tt.expected, tt.err.ShouldRetry())
			}
		})
	}
}

func TestIncrementRetry(t *testing.T) {
	err := TransientError("op", "msg", 5)

	if err.RetryCount != 0 {
		t.Error("Expected initial retry count to be 0")
	}

	err.IncrementRetry()
	if err.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", err.RetryCount)
	}

	err.IncrementRetry()
	if err.RetryCount != 2 {
		t.Errorf("Expected retry count 2, got %d", err.RetryCount)
	}
}

func TestIsType(t *testing.T) {
	err := UserError("op", "msg")

	if !err.IsType(ErrorTypeUser) {
		t.Error("Expected IsType(ErrorTypeUser) to be true")
	}
	if err.IsType(ErrorTypeSystem) {
		t.Error("Expected IsType(ErrorTypeSystem) to be false")
	}
}

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errType  ErrorType
		expected string
	}{
		{ErrorTypeUser, "User"},
		{ErrorTypeSystem, "System"},
		{ErrorTypeFatal, "Fatal"},
		{ErrorTypeTransient, "Transient"},
		{ErrorType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.errType.String() != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.errType.String())
			}
		})
	}
}

func TestConvenienceConstructors(t *testing.T) {
	t.Run("UserError", func(t *testing.T) {
		err := UserError("operation", "message")
		if err.Type != ErrorTypeUser {
			t.Error("Expected User error type")
		}
	})

	t.Run("SystemError", func(t *testing.T) {
		cause := errors.New("cause")
		err := SystemError("operation", cause)
		if err.Type != ErrorTypeSystem {
			t.Error("Expected System error type")
		}
		if err.Cause != cause {
			t.Error("Expected cause to be set")
		}
	})

	t.Run("FatalError", func(t *testing.T) {
		cause := errors.New("cause")
		err := FatalError("operation", cause)
		if err.Type != ErrorTypeFatal {
			t.Error("Expected Fatal error type")
		}
	})

	t.Run("TransientError", func(t *testing.T) {
		err := TransientError("operation", "message", 5)
		if err.Type != ErrorTypeTransient {
			t.Error("Expected Transient error type")
		}
		if !err.Retryable {
			t.Error("Expected error to be retryable")
		}
	})
}

func TestIsErrorType(t *testing.T) {
	userErr := UserError("op", "msg")
	sysErr := SystemError("op", errors.New("err"))
	fatalErr := FatalError("op", errors.New("err"))
	transErr := TransientError("op", "msg", 3)
	stdErr := errors.New("standard error")

	if !IsUserError(userErr) {
		t.Error("Expected IsUserError to be true for user error")
	}
	if IsUserError(sysErr) {
		t.Error("Expected IsUserError to be false for system error")
	}
	if IsUserError(stdErr) {
		t.Error("Expected IsUserError to be false for standard error")
	}

	if !IsSystemError(sysErr) {
		t.Error("Expected IsSystemError to be true for system error")
	}
	if IsSystemError(userErr) {
		t.Error("Expected IsSystemError to be false for user error")
	}

	if !IsFatalError(fatalErr) {
		t.Error("Expected IsFatalError to be true for fatal error")
	}
	if IsFatalError(userErr) {
		t.Error("Expected IsFatalError to be false for user error")
	}

	if !IsTransientError(transErr) {
		t.Error("Expected IsTransientError to be true for transient error")
	}
	if IsTransientError(userErr) {
		t.Error("Expected IsTransientError to be false for user error")
	}
}

func TestHandler(t *testing.T) {
	var handledType ErrorType
	var loggedErr *Error

	handler := &Handler{
		OnUser: func(e *Error) error {
			handledType = ErrorTypeUser
			return nil
		},
		OnSystem: func(e *Error) error {
			handledType = ErrorTypeSystem
			return nil
		},
		OnFatal: func(e *Error) error {
			handledType = ErrorTypeFatal
			return nil
		},
		OnTransient: func(e *Error) error {
			handledType = ErrorTypeTransient
			return nil
		},
		Logger: func(e *Error) {
			loggedErr = e
		},
	}

	t.Run("handle user error", func(t *testing.T) {
		handledType = -1
		loggedErr = nil
		err := UserError("op", "msg")
		handler.Handle(err)

		if handledType != ErrorTypeUser {
			t.Error("Expected user error to be handled")
		}
		if loggedErr == nil {
			t.Error("Expected error to be logged")
		}
	})

	t.Run("handle system error", func(t *testing.T) {
		handledType = -1
		err := SystemError("op", errors.New("err"))
		handler.Handle(err)

		if handledType != ErrorTypeSystem {
			t.Error("Expected system error to be handled")
		}
	})

	t.Run("handle standard error as system", func(t *testing.T) {
		handledType = -1
		handler.Handle(errors.New("standard error"))

		if handledType != ErrorTypeSystem {
			t.Error("Expected standard error to be handled as system error")
		}
	})

	t.Run("handle nil error", func(t *testing.T) {
		err := handler.Handle(nil)
		if err != nil {
			t.Error("Expected nil to be handled gracefully")
		}
	})
}

func TestRecoveryRetry(t *testing.T) {
	t.Run("success on first try", func(t *testing.T) {
		recovery := DefaultRecovery()
		attempts := 0

		err := recovery.Retry("operation", func() error {
			attempts++
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("success after retries", func(t *testing.T) {
		recovery := &Recovery{
			MaxRetries:    3,
			RetryDelay:    time.Millisecond,
			BackoffFactor: 1.0,
		}
		attempts := 0

		err := recovery.Retry("operation", func() error {
			attempts++
			if attempts < 3 {
				return TransientError("op", "temporary failure", 5)
			}
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error after retries, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("non-transient error stops immediately", func(t *testing.T) {
		recovery := &Recovery{
			MaxRetries:    3,
			RetryDelay:    time.Millisecond,
			BackoffFactor: 1.0,
		}
		attempts := 0

		err := recovery.Retry("operation", func() error {
			attempts++
			return UserError("op", "permanent failure")
		})

		if err == nil {
			t.Error("Expected error to be returned")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt (no retries), got %d", attempts)
		}
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		recovery := &Recovery{
			MaxRetries:    2,
			RetryDelay:    time.Millisecond,
			BackoffFactor: 1.0,
		}
		attempts := 0

		err := recovery.Retry("operation", func() error {
			attempts++
			return TransientError("op", "persistent failure", 10)
		})

		if err == nil {
			t.Error("Expected error after max retries")
		}
		if attempts != 3 { // Initial + 2 retries
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("standard error wrapped as system error", func(t *testing.T) {
		recovery := DefaultRecovery()

		err := recovery.Retry("operation", func() error {
			return errors.New("standard error")
		})

		if err == nil {
			t.Error("Expected error")
		}
		if !IsSystemError(err) {
			t.Error("Expected standard error to be wrapped as system error")
		}
	})
}

func TestDefaultRecovery(t *testing.T) {
	r := DefaultRecovery()

	if r.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", r.MaxRetries)
	}
	if r.RetryDelay != time.Second {
		t.Errorf("Expected RetryDelay 1s, got %v", r.RetryDelay)
	}
	if r.BackoffFactor != 2.0 {
		t.Errorf("Expected BackoffFactor 2.0, got %v", r.BackoffFactor)
	}
}

func TestStackFrameCapture(t *testing.T) {
	err := New(ErrorTypeUser, "op", "msg")

	if len(err.Stack) == 0 {
		t.Fatal("Expected stack frames to be captured")
	}

	// Check that this test function appears in the stack
	found := false
	for _, frame := range err.Stack {
		if strings.Contains(frame.Function, "TestStackFrameCapture") {
			found = true
			if !strings.Contains(frame.File, "errors_test.go") {
				t.Error("Expected file path in stack frame")
			}
			if frame.Line == 0 {
				t.Error("Expected non-zero line number")
			}
			break
		}
	}

	if !found {
		t.Error("Expected test function in stack trace")
	}
}

func TestConcurrentErrorCreation(t *testing.T) {
	// Test thread-safety of error creation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			err := New(ErrorTypeSystem, fmt.Sprintf("op-%d", id), fmt.Sprintf("msg-%d", id))
			err.WithContext("id", id).WithCode(fmt.Sprintf("ERR_%d", id))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
