package testutil

import (
	"strings"
	"sync"
)

// LogEntry represents a single log message captured by the mock
type LogEntry struct {
	Level   string
	Message string
}

// MockLogger captures log messages for verification
type MockLogger struct {
	mu      sync.RWMutex
	Entries []LogEntry
}

// NewMockLogger creates a new MockLogger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Log(level, msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, LogEntry{Level: level, Message: msg})
}

// AssertLogContains verifies that a log message containing the target string was recorded
func (m *MockLogger) AssertLogContains(t TB, level, target string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, entry := range m.Entries {
		if (level == "" || entry.Level == level) && strings.Contains(entry.Message, target) {
			return
		}
	}
	t.Errorf("expected log at level %q containing %q, but it was not found", level, target)
}

// Reset clears all captured log entries
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = nil
}
