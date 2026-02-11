package handlers_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers"
)

func TestResize_Handling(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Terminal resize updates model", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		handlers.HandleUpdate(m, msg)

		if m.Display.Width != 120 {
			t.Errorf("expected width 120, got %d", m.Display.Width)
		}
		if m.Display.Height != 40 {
			t.Errorf("expected height 40, got %d", m.Display.Height)
		}
	})

	t.Run("Viewport height recalculated on resize", func(t *testing.T) {
		m.Config.UI.ShowHeader = true
		msg := tea.WindowSizeMsg{Width: 80, Height: 24}
		handlers.HandleUpdate(m, msg)

		if m.Display.ViewportHeight == 0 {
			t.Error("viewport height should be non-zero after resize")
		}
	})
}

func TestGlobal_Quit(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("First ctrl+c shows warning", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		handlers.HandleUpdate(m, msg)

		if m.Message.Text != "press [ctrl+c] again to close" {
			t.Errorf("expected warning message, got %q", m.Message.Text)
		}
	})

	t.Run("Second ctrl+c returns Quit command", func(t *testing.T) {
		m.Message.Push("press [ctrl+c] again to close", false)
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		cmd := handlers.HandleUpdate(m, msg)

		if cmd == nil {
			t.Fatal("expected quit command, got nil")
		}
		res := cmd()
		if quitMsg, ok := res.(tea.QuitMsg); !ok {
			t.Errorf("expected tea.Quit command, got %T (%+v)", res, quitMsg)
		}
	})
}

func TestGlobal_InputSafety(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Dot key ignored during input", func(t *testing.T) {
		m.StartInput(tuictx.InputSearch)
		m.UI.ActiveView = tuictx.ViewMain

		handlers.HandleUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(".")})

		if m.UI.ActiveView != tuictx.ViewMain {
			t.Error("Settings should NOT open when input is active")
		}
	})

	t.Run("Question mark ignored during input", func(t *testing.T) {
		m.StartInput(tuictx.InputSearch)
		m.UI.ActiveView = tuictx.ViewMain

		handlers.HandleUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

		if m.UI.ActiveView != tuictx.ViewMain {
			t.Error("Help should NOT open when input is active")
		}
	})
}
