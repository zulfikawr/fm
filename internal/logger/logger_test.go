package logger

import (
	"os"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestLogger(t *testing.T) {
	mock := testutil.NewMockLogger()
	SetLogger(mock)

	Info("test info message")
	mock.AssertLogContains(t, LevelInfo, "test info message")

	Error("test error message")
	mock.AssertLogContains(t, LevelError, "test error message")
}

func TestFileLogger(t *testing.T) {
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()

	logPath := tmp.Join("test.log")
	SetLogPath(logPath)
	defer SetLogPath("") // Reset after test

	// Restore real logger for this test
	realLogger := &fileLogger{}
	SetLogger(realLogger)

	msg := "test log entry"
	Log(LevelDebug, msg)

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
