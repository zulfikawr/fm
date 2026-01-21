package handlers

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
)

func TestResize_Handling(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Terminal resize updates model", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		HandleUpdate(m, msg)

		if m.Display.Width != 120 {
			t.Errorf("expected width 120, got %d", m.Display.Width)
		}
		if m.Display.Height != 40 {
			t.Errorf("expected height 40, got %d", m.Display.Height)
		}
	})

	t.Run("Viewport height recalculated on resize", func(t *testing.T) {
		m.Config.ShowHeader = true
		msg := tea.WindowSizeMsg{Width: 80, Height: 24}
		HandleUpdate(m, msg)

		if m.Display.ViewportHeight == 0 {
			t.Error("viewport height should be non-zero after resize")
		}
	})
}
