package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMain(m *testing.M) {
	// Isolate config for all tests in this package
	tempDir, err := os.MkdirTemp("", "fm-test-*")
	if err != nil {
		panic(err)
	}

	config.SetConfigPath(filepath.Join(tempDir, "config.json"))

	code := m.Run()

	// Clean up
	_ = os.RemoveAll(tempDir)
	os.Exit(code)
}

type testModelWrapper struct {
	m *tuictx.Model
}

func NewTestModelWrapper(m *tuictx.Model) *testModelWrapper {
	// Mock external actions to avoid hanging on tea.ExecProcess during tests
	OpenFileAction = func(_ *tuictx.Model, _ core.Item) tea.Cmd {
		return nil
	}
	OpenFileAtLineAction = func(_ *tuictx.Model, _ string, _ int) tea.Cmd {
		return nil
	}

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
