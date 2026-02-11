package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorType categorizes errors by their severity and handling strategy
type ErrorType int

const (
	// ErrorTypeUser - User-facing errors that should be displayed in the UI
	// Examples: file not found, permission denied, invalid input
	ErrorTypeUser ErrorType = iota

	// ErrorTypeSystem - System errors that should be logged and gracefully handled
	// Examples: temporary network issues, cache failures, resource constraints
	ErrorTypeSystem

	// ErrorTypeFatal - Fatal errors requiring cleanup and potential exit
	// Examples: filesystem unavailable, critical resource exhaustion
	ErrorTypeFatal

	// ErrorTypeTransient - Temporary errors that should be retried
	// Examples: network timeout, temporary file lock, rate limiting
	ErrorTypeTransient
)

// String returns the string representation of ErrorType
func (t ErrorType) String() string {
	switch t {
	case ErrorTypeUser:
		return "User"
	case ErrorTypeSystem:
		return "System"
	case ErrorTypeFatal:
		return "Fatal"
	case ErrorTypeTransient:
		return "Transient"
	default:
		return "Unknown"
	}
}

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// Error represents a TUI error with context and stack trace
type Error struct {
	// Type categorizes the error
	Type ErrorType

	// Code is an optional error code for programmatic handling
	Code string

	// Message is the human-readable error message
	Message string

	// Operation is the operation that was being performed
	Operation string

	// Cause is the underlying error
	Cause error

	// Stack is the stack trace where the error occurred
	Stack []StackFrame

	// Context contains additional context information
	Context map[string]interface{}

	// Timestamp when the error occurred
	Timestamp time.Time

	// Retryable indicates if the operation can be retried
	Retryable bool

	// RetryCount tracks how many times retry has been attempted
	RetryCount int

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
}

// New creates a new TUI error
func New(errType ErrorType, operation, message string) *Error {
	return &Error{
		Type:      errType,
		Operation: operation,
		Message:   message,
		Stack:     captureStack(2), // Skip New and its caller
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// Wrap wraps an existing error with TUI error context
func Wrap(err error, errType ErrorType, operation string) *Error {
	if err == nil {
		return nil
	}

	// If already a TUI error, update it
	if tuiErr, ok := err.(*Error); ok {
		if tuiErr.Operation == "" {
			tuiErr.Operation = operation
		}
		if tuiErr.Type != errType {
			tuiErr.Type = errType
		}
		return tuiErr
	}

	return &Error{
		Type:      errType,
		Operation: operation,
		Message:   err.Error(),
		Cause:     err,
		Stack:     captureStack(2),
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// WithCode adds an error code
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// WithContext adds context information
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRetry marks the error as retryable with max attempts
func (e *Error) WithRetry(maxRetries int) *Error {
	e.Retryable = true
	e.MaxRetries = maxRetries
	return e
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("%s: %s", e.Operation, e.Message)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *Error) Unwrap() error {
	return e.Cause
}

// UserMessage returns a user-friendly error message for display in the UI
func (e *Error) UserMessage() string {
	switch e.Type {
	case ErrorTypeUser:
		return e.Message
	case ErrorTypeSystem:
		if e.Message != "" {
			return fmt.Sprintf("System Error: %s", e.Message)
		}
		return "A system error occurred. Please try again."
	case ErrorTypeFatal:
		return "A critical error occurred. The application may need to restart."
	case ErrorTypeTransient:
		if e.Retryable && e.RetryCount < e.MaxRetries {
			return fmt.Sprintf("%s (Retrying... %d/%d)", e.Message, e.RetryCount+1, e.MaxRetries)
		}
		return e.Message
	default:
		return "An unexpected error occurred."
	}
}

// LogMessage returns a detailed message for logging
func (e *Error) LogMessage() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s", e.Type, e.Error()))

	if e.Code != "" {
		sb.WriteString(fmt.Sprintf(" [Code: %s]", e.Code))
	}

	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(" | Cause: %v", e.Cause))
	}

	if len(e.Context) > 0 {
		sb.WriteString(" | Context: {")
		first := true
		for k, v := range e.Context {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %v", k, v))
			first = false
		}
		sb.WriteString("}")
	}

	return sb.String()
}

// StackTrace returns a formatted stack trace
func (e *Error) StackTrace() string {
	if len(e.Stack) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Stack trace:\n")
	for i, frame := range e.Stack {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, frame.Function))
		sb.WriteString(fmt.Sprintf("     %s:%d\n", frame.File, frame.Line))
	}
	return sb.String()
}

// ShouldRetry returns true if the error should be retried
func (e *Error) ShouldRetry() bool {
	return e.Retryable && e.RetryCount < e.MaxRetries
}

// IncrementRetry increments the retry counter
func (e *Error) IncrementRetry() {
	e.RetryCount++
}

// IsType checks if the error is of a specific type
func (e *Error) IsType(errType ErrorType) bool {
	return e.Type == errType
}

// captureStack captures the current stack trace
func captureStack(skip int) []StackFrame {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip+1, pcs[:])

	frames := make([]StackFrame, 0, n)
	pcSlice := pcs[:n]
	for i := range pcSlice {
		pc := pcSlice[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)

		// Skip runtime frames
		if strings.HasPrefix(file, "runtime/") {
			continue
		}

		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// Handler provides utilities for handling errors
type Handler struct {
	// OnUser is called for user-facing errors
	OnUser func(*Error) error

	// OnSystem is called for system errors
	OnSystem func(*Error) error

	// OnFatal is called for fatal errors
	OnFatal func(*Error) error

	// OnTransient is called for transient errors
	OnTransient func(*Error) error

	// Logger is called to log all errors
	Logger func(*Error)
}

// Handle processes an error according to its type
func (h *Handler) Handle(err error) error {
	if err == nil {
		return nil
	}

	tuiErr, ok := err.(*Error)
	if !ok {
		// Wrap unknown errors as system errors
		tuiErr = Wrap(err, ErrorTypeSystem, "unknown operation")
	}

	// Log the error
	if h.Logger != nil {
		h.Logger(tuiErr)
	}

	// Handle based on type
	switch tuiErr.Type {
	case ErrorTypeUser:
		if h.OnUser != nil {
			return h.OnUser(tuiErr)
		}
	case ErrorTypeSystem:
		if h.OnSystem != nil {
			return h.OnSystem(tuiErr)
		}
	case ErrorTypeFatal:
		if h.OnFatal != nil {
			return h.OnFatal(tuiErr)
		}
	case ErrorTypeTransient:
		if h.OnTransient != nil {
			return h.OnTransient(tuiErr)
		}
	}

	return tuiErr
}

// Common error constructors for convenience

// UserError creates a user-facing error
func UserError(operation, message string) *Error {
	return New(ErrorTypeUser, operation, message)
}

// SystemError creates a system error
func SystemError(operation string, cause error) *Error {
	return Wrap(cause, ErrorTypeSystem, operation)
}

// FatalError creates a fatal error
func FatalError(operation string, cause error) *Error {
	return Wrap(cause, ErrorTypeFatal, operation)
}

// TransientError creates a transient error with retry
func TransientError(operation, message string, maxRetries int) *Error {
	return New(ErrorTypeTransient, operation, message).WithRetry(maxRetries)
}

// IsUserError checks if an error is user-facing
func IsUserError(err error) bool {
	if tuiErr, ok := err.(*Error); ok {
		return tuiErr.Type == ErrorTypeUser
	}
	return false
}

// IsSystemError checks if an error is a system error
func IsSystemError(err error) bool {
	if tuiErr, ok := err.(*Error); ok {
		return tuiErr.Type == ErrorTypeSystem
	}
	return false
}

// IsFatalError checks if an error is fatal
func IsFatalError(err error) bool {
	if tuiErr, ok := err.(*Error); ok {
		return tuiErr.Type == ErrorTypeFatal
	}
	return false
}

// IsTransientError checks if an error is transient
func IsTransientError(err error) bool {
	if tuiErr, ok := err.(*Error); ok {
		return tuiErr.Type == ErrorTypeTransient
	}
	return false
}

// Recovery provides error recovery mechanisms
type Recovery struct {
	MaxRetries    int
	RetryDelay    time.Duration
	BackoffFactor float64
}

// DefaultRecovery returns a recovery with sensible defaults
func DefaultRecovery() *Recovery {
	return &Recovery{
		MaxRetries:    3,
		RetryDelay:    time.Second,
		BackoffFactor: 2.0,
	}
}

// Retry executes a function with retry logic for transient errors
func (r *Recovery) Retry(operation string, fn func() error) error {
	var lastErr error
	delay := r.RetryDelay

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * r.BackoffFactor)
		}

		err := fn()
		if err == nil {
			return nil
		}

		// Check if it's a transient error
		if tuiErr, ok := err.(*Error); ok {
			if tuiErr.Type == ErrorTypeTransient && tuiErr.ShouldRetry() {
				tuiErr.IncrementRetry()
				lastErr = tuiErr
				continue
			}
			// Non-transient error, return immediately
			return err
		}

		// Unknown error, wrap and return
		return SystemError(operation, err)
	}

	// Max retries exceeded
	if lastErr != nil {
		return lastErr
	}
	return SystemError(operation, fmt.Errorf("max retries exceeded"))
}
