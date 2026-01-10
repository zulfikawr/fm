package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLog(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-logger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")
	SetLogPath(logPath)
	defer SetLogPath("") // Reset

	testMsg := "Test log message unique"
	Info(testMsg)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Could not read log file: %v", err)
	}

	if !strings.Contains(string(data), testMsg) {
		t.Errorf("Log file does not contain expected message. Content: %s", string(data))
	}

	if !strings.Contains(string(data), "INFO") {
		t.Errorf("Log file missing level INFO. Content: %s", string(data))
	}
}
