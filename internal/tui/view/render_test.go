package view_test

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/view"
)

func TestRender_Modals(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/test")
	m.Display.Width = 80
	m.Display.Height = 24

	t.Run("Settings", func(t *testing.T) {
		m.UI.SettingsOpen = true
		v := view.Render(m)
		if !strings.Contains(testutil.StripANSI(v), "Settings") {
			t.Error("expected Settings in view")
		}
		m.UI.SettingsOpen = false
	})

	t.Run("Logs", func(t *testing.T) {
		m.UI.LogOpen = true
		v := view.Render(m)
		// Empty logs show "No operations logged yet."
		if !strings.Contains(testutil.StripANSI(v), "logged yet") {
			t.Error("expected empty logs message in view")
		}
		m.UI.LogOpen = false
	})

	t.Run("Loading", func(t *testing.T) {
		m.UI.Loading = true
		v := view.Render(m)
		if !strings.Contains(testutil.StripANSI(v), "Loading...") {
			t.Error("expected Loading... in view")
		}
		m.UI.Loading = false
	})
}
