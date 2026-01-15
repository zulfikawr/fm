package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logMu         sync.Mutex
	customLogPath string
	currentLogger Logger = &fileLogger{}
)

// Logger interface for plugging in different logging implementations
type Logger interface {
	Log(level, msg string)
}

type fileLogger struct{}

func (l *fileLogger) Log(level, msg string) {
	path := GetLogPath()
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0o755)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] %s: %s\n", timestamp, level, msg)
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
func Log(level, msg string) {
	logMu.Lock()
	defer logMu.Unlock()
	currentLogger.Log(level, msg)
}

// Info logs an informational message.
func Info(msg string) {
	Log("INFO", msg)
}

// Error logs an error message.
func Error(msg string) {
	Log("ERROR", msg)
}
