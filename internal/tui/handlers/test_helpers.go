package handlers

import (
	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"
	"fm/internal/tui/view"
	tea "github.com/charmbracelet/bubbletea"
)

type testModelWrapper struct {
	m *tuictx.Model
}

func newTestModelWrapper(m *tuictx.Model) *testModelWrapper {
	m.GS = testutil.NewMockGitService() // Inject mock
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
