package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Level represents the severity of a log message
type Level = string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelFatal Level = "FATAL"
)

var (
	logMu         sync.Mutex
	customLogPath string
	currentLogger Logger = &fileLogger{}
)

// Logger interface for plugging in different logging implementations
type Logger interface {
	Log(level string, msg string)
}

type fileLogger struct{}

func (l *fileLogger) Log(level string, msg string) {
	path := GetLogPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Get caller info (3 levels up: this -> Log/Info/Error -> actual caller)
	_, file, line, ok := runtime.Caller(3)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	fmt.Fprintf(f, "[%s] [%s] [%s] %s\n", timestamp, level, caller, msg)
}

// SetLogger overrides the global logger (useful for testing)
func SetLogger(l Logger) {
	logMu.Lock()
	defer logMu.Unlock()
	currentLogger = l
}

// SetLogPath overrides the default log path (useful for testing).
func SetLogPath(path string) {
	logMu.Lock()
	defer logMu.Unlock()
	customLogPath = path
}

// GetLogPath returns the platform-specific path to the log file.
func GetLogPath() string {
	if customLogPath != "" {
		return customLogPath
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".fm.log")
	}
	return filepath.Join(configDir, "fm", "fm.log")
}

// Log writes a message using the current logger
func Log(level string, msg string) {
	logMu.Lock()
	defer logMu.Unlock()
	currentLogger.Log(level, msg)
}

// Logf writes a formatted message using the current logger
func Logf(level Level, format string, args ...any) {
	Log(level, fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func Debug(msg string) {
	Log(LevelDebug, msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...any) {
	Logf(LevelDebug, format, args...)
}

// Info logs an informational message.
func Info(msg string) {
	Log(LevelInfo, msg)
}

// Infof logs a formatted informational message.
func Infof(format string, args ...any) {
	Logf(LevelInfo, format, args...)
}

// Warn logs a warning message.
func Warn(msg string) {
	Log(LevelWarn, msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...any) {
	Logf(LevelWarn, format, args...)
}

// Error logs an error message.
func Error(msg string) {
	Log(LevelError, msg)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	Logf(LevelError, format, args...)
}

// Fatal logs a fatal message and exits.
func Fatal(msg string) {
	Log(LevelFatal, msg)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits.
func Fatalf(format string, args ...any) {
	Logf(LevelFatal, format, args...)
	os.Exit(1)
}
