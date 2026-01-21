package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestLogger(t *testing.T) {
	mock := &testutil.MockLogger{}
	SetLogger(mock)
	defer SetLogger(&fileLogger{}) // Reset after test

	Info("test info message")
	mock.AssertLogContains(t, LevelInfo, "test info message")

	Error("test error message")
	mock.AssertLogContains(t, LevelError, "test error message")
}

func TestFileLogger(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	logPath := filepath.Join(tmpDir, "test.log")

	SetLogPath(logPath)
	defer SetLogPath("") // Reset after test

	// Ensure we are using the real file logger
	realLogger := &fileLogger{}
	SetLogger(realLogger)
	defer SetLogger(&fileLogger{}) // Reset after test

	msg := "test log entry"
	Log(LevelInfo, msg)

	content, err := os.ReadFile(logPath)
	testutil.AssertNoError(t, err, "Log file should be readable")

	if !strings.Contains(string(content), msg) {
		t.Errorf("Expected log file to contain %q, but it did not", msg)
	}
}

func TestGetLogPath(t *testing.T) {
	path := GetLogPath()
	if path == "" {
		t.Error("GetLogPath returned empty string")
	}
}
