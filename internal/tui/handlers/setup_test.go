package handlers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers"
	"github.com/zulfikawr/fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMain(m *testing.M) {
	// Isolate config for all tests in this package
	tempDir, err := os.MkdirTemp("", "fm-handler-test-*")
	if err != nil {
		panic(err)
	}

	config.SetConfigPath(filepath.Join(tempDir, "config.json"))

	code := m.Run()

	// Clean up
	if err := os.RemoveAll(tempDir); err != nil {
		panic(err)
	}
	os.Exit(code)
}

type testModelWrapper struct {
	m *tuictx.Model
}

func NewTestModelWrapper(m *tuictx.Model) *testModelWrapper {
	// Mock external actions to avoid hanging on tea.ExecProcess during tests
	handlers.OpenFileAction = func(model *tuictx.Model, item core.Item) tea.Cmd {
		return nil
	}
	handlers.OpenFileAtLineAction = func(model *tuictx.Model, path string, line int) tea.Cmd {
		return nil
	}

	m.GS = testutil.NewMockGitService() // Inject mock
	return &testModelWrapper{m: m}
}

func (w *testModelWrapper) Init() tea.Cmd { return nil }
func (w *testModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := handlers.HandleUpdate(w.m, msg)
	return w, cmd
}
func (w *testModelWrapper) View() string {
	return view.Render(w.m)
}
