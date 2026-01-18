package handlers

import (
	tea "github.com/charmbracelet/bubbletea"

	"fm/internal/files/core"
	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"
	"fm/internal/tui/view"
)

type testModelWrapper struct {
	m *tuictx.Model
}

func newTestModelWrapper(m *tuictx.Model) *testModelWrapper {
	m.GS = testutil.NewMockGitService() // Inject mock

	// Mock external actions to avoid hanging on tea.ExecProcess during tests
	openFileAction = func(_ *tuictx.Model, _ core.Item) tea.Cmd {
		return nil
	}
	openFileAtLineAction = func(_ *tuictx.Model, _ string, _ int) tea.Cmd {
		return nil
	}

	return &testModelWrapper{m: m}
}

func (w *testModelWrapper) Init() tea.Cmd { return nil }
func (w *testModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := HandleUpdate(w.m, msg)
	return w, cmd
}
func (w *testModelWrapper) View() string {
	return view.Render(w.m)
}
