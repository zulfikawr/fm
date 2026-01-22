package testutil

import (
	"context"
)

// MockGitService implements a git service interface for testing.
type MockGitService struct {
	GetStatusFunc       func(ctx context.Context, path string) (map[string]string, string)
	GetRootFunc         func(ctx context.Context, path string) string
	GetIgnoredFilesFunc func(ctx context.Context, repoRoot string) ([]string, error)
	IsEnabledFunc       func() bool
	SetEnabledFunc      func(enabled bool)
}

func NewMockGitService() *MockGitService {
	return &MockGitService{
		IsEnabledFunc: func() bool { return true },
	}
}

func (m *MockGitService) GetStatus(ctx context.Context, path string) (map[string]string, string) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx, path)
	}
	return make(map[string]string), "main"
}

func (m *MockGitService) GetRoot(ctx context.Context, path string) string {
	if m.GetRootFunc != nil {
		return m.GetRootFunc(ctx, path)
	}
	return ""
}

func (m *MockGitService) GetIgnoredFiles(ctx context.Context, repoRoot string) ([]string, error) {
	if m.GetIgnoredFilesFunc != nil {
		return m.GetIgnoredFilesFunc(ctx, repoRoot)
	}
	return nil, nil
}

func (m *MockGitService) IsEnabled() bool {
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc()
	}
	return true
}

func (m *MockGitService) SetEnabled(enabled bool) {
	if m.SetEnabledFunc != nil {
		m.SetEnabledFunc(enabled)
	}
}
