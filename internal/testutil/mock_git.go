package testutil

import (
	"context"
	"sync"
)

// MockGitService is an advanced, thread-safe mock of git.GitService
type MockGitService struct {
	mu sync.RWMutex

	IsEnabledFunc  func() bool
	SetEnabledFunc func(enabled bool)
	GetStatusFunc  func(ctx context.Context, path string) (map[string]string, string)
	GetRootFunc    func(ctx context.Context, path string) string

	Calls []Call
}

// NewMockGitService creates a new MockGitService
func NewMockGitService() *MockGitService {
	return &MockGitService{}
}

func (m *MockGitService) recordCall(method string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, Call{Method: method, Args: args})
}

// AssertCalled verifies that a method was called
func (m *MockGitService) AssertCalled(t TB, method string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.Calls {
		if c.Method == method {
			return
		}
	}
	t.Errorf("expected git method %s to be called, but it was not", method)
}

func (m *MockGitService) IsEnabled() bool {
	m.recordCall("IsEnabled")
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc()
	}
	return true
}

func (m *MockGitService) SetEnabled(enabled bool) {
	m.recordCall("SetEnabled", enabled)
	if m.SetEnabledFunc != nil {
		m.SetEnabledFunc(enabled)
	}
}

func (m *MockGitService) GetStatus(ctx context.Context, path string) (map[string]string, string) {
	m.recordCall("GetStatus", path)
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx, path)
	}
	return make(map[string]string), "main"
}

func (m *MockGitService) GetRoot(ctx context.Context, path string) string {
	m.recordCall("GetRoot", path)
	if m.GetRootFunc != nil {
		return m.GetRootFunc(ctx, path)
	}
	return ""
}
