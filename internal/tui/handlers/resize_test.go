package handlers

import (
	"testing"

	"fm/internal/testutil"
	"fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleUpdate_Resize(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/test")

	t.Run("Standard Resize", func(t *testing.T) {
		m.Config.ShowHeader = false
		msg := tea.WindowSizeMsg{Width: 100, Height: 20}
		HandleUpdate(m, msg)

		testutil.AssertEqual(t, 100, m.Display.Width, "Width should be 100")
		testutil.AssertEqual(t, 20, m.Display.Height, "Height should be 20")
		// Height(20) - Header(1) - Footer(1) = 18
		testutil.AssertEqual(t, 18, m.Display.ViewportHeight, "ViewportHeight should be 18")
	})

	t.Run("Resize with List Header", func(t *testing.T) {
		m.Config.ShowHeader = true
		msg := tea.WindowSizeMsg{Width: 100, Height: 20}
		HandleUpdate(m, msg)

		// Height(20) - AppHeader(1) - AppFooter(1) - ListHeader(3) = 15
		testutil.AssertEqual(t, 15, m.Display.ViewportHeight, "ViewportHeight should be 15 with list header")
	})
}
