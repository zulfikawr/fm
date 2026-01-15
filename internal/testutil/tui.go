package testutil

import (
	"io"
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

// AssertViewContains verifies that the current view contains the target string
func AssertViewContains(t *testing.T, tm *teatest.TestModel, target string) {
	t.Helper()
	out, err := io.ReadAll(tm.Output())
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	view := StripANSI(string(out))
	if !strings.Contains(view, target) {
		t.Errorf("expected view to contain %q, but it did not.\nView content:\n%s", target, view)
	}
}
