package testutil

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// NewTestModel creates a new teatest TestModel
func NewTestModel(t *testing.T, m tea.Model, opts ...teatest.TestOption) *teatest.TestModel {
	return teatest.NewTestModel(t, m, opts...)
}

// WaitAndAssertView waits for the target string to appear in the model's view
func WaitAndAssertView(t *testing.T, tm *teatest.TestModel, target string, timeout time.Duration) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return strings.Contains(StripANSI(string(out)), target)
	}, teatest.WithDuration(timeout))
}
