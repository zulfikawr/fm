package app

import (
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSettings_ToggleLogic(t *testing.T) {
	tmp := testutil.NewTempFolder(t)
	defer tmp.Cleanup()
	config.SetConfigPath(tmp.Join("config.json"))
	defer config.SetConfigPath("")

	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.SettingsOpen = true
	m.Config.ShowHidden = false

	HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if !m.Config.ShowHidden {
		t.Error("Toggle failed in direct call")
	}
}
