package app_test

import (
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSettings_ToggleLogic(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	config.SetConfigPath(filepath.Join(tmpDir, "config.json"))
	defer config.SetConfigPath("")

	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.SettingsOpen = true
	m.Config.ShowHidden = false

	app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if !m.Config.ShowHidden {
		t.Error("Toggle failed in direct call")
	}
}
