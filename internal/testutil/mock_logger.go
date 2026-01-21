package testutil

import (
	"strings"
	"sync"
	"testing"
)

// MockLogger is a generic mock for a logger.
type MockLogger struct {
	mu   sync.RWMutex
	Logs []string
}

func (m *MockLogger) Log(level string, msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Logs = append(m.Logs, level+": "+msg)
}

func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Logs = nil
}

func (m *MockLogger) Contains(target string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, l := range m.Logs {
		if strings.Contains(l, target) {
			return true
		}
	}
	return false
}

func (m *MockLogger) AssertLogContains(t *testing.T, level string, target string) {
	t.Helper()
	if !m.Contains(level + ": " + target) {
		t.Errorf("expected logs to contain %q with level %q, but they did not", target, level)
	}
}
