package handlers

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
)

func TestRouter_GlobalMessages(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("WindowSizeMsg", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		HandleUpdate(m, msg)
		if m.Display.Width != 100 || m.Display.Height != 50 {
			t.Errorf("expected 100x50, got %dx%d", m.Display.Width, m.Display.Height)
		}
	})

	t.Run("WatchEventMsg", func(t *testing.T) {
		m.Watcher.IsListening = true
		HandleUpdate(m, WatchEventMsg{})
		if !m.Watcher.IsListening {
			t.Error("expected IsListening to still be true (debouncing)")
		}
		if m.Watcher.DebounceTimer == nil {
			t.Error("expected debounce timer to be set")
		}

		// Simulate debounce timer expiration
		HandleUpdate(m, DebounceWatchMsg{})
		if m.Watcher.IsListening {
			t.Error("expected IsListening to be false after debounce")
		}
	})

	t.Run("ClearMsg", func(t *testing.T) {
		m.Message.Push("test", false)
		m.Operations.Progress.Show("label")
		HandleUpdate(m, ClearMsg{})
		if m.Message.Text != "" {
			t.Error("expected message to be cleared")
		}
	})

	t.Run("ErrorMsg", func(t *testing.T) {
		m.UI.Loading = true
		msg := ErrorMsg{Err: errors.New("fail"), LogID: "123"}
		cmd := HandleUpdate(m, msg)
		if cmd == nil {
			t.Fatal("expected reload command after error")
		}
		// m.UI.Loading is false in the reducer before returning Reload()
		// but Reload() immediately sets it to true if called or we can just check the reducer state
	})
}

func TestRouter_GlobalKeys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	t.Run("Toggle Logs (alt+l)", func(t *testing.T) {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l"), Alt: true})
		time.Sleep(10 * time.Millisecond)
		if !m.UI.LogOpen {
			t.Error("expected logs to be open")
		}
	})

	t.Run("Global Esc", func(t *testing.T) {
		m.UI.LogOpen = true
		tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
		time.Sleep(10 * time.Millisecond)
		if m.UI.LogOpen {
			t.Error("expected logs closed on Esc")
		}
	})

	_ = tm.Quit()
}

func TestRouter_FinalizeInput(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	t.Run("Finalize Search", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.Mode = tuictx.InputSearch
		m.Inputs.ActiveInput.SetValue("query")
		tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(10 * time.Millisecond)
		if m.UI.InputActive {
			t.Error("expected input to be inactive after Enter")
		}
	})

	t.Run("Finalize InputFuzzySearch", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.Mode = tuictx.InputFuzzySearch
		m.Search.Results = []core.FileResult{{Path: "test", Matches: []core.Match{{Line: 1}}}}
		tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(10 * time.Millisecond)
		if m.UI.InputActive {
			t.Error("expected input to be inactive after Enter")
		}
		if m.Search.Results != nil {
			t.Error("expected search results to be cleared after Enter")
		}
	})

	t.Run("Input Tab Toggle", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.Mode = tuictx.InputGoto
		initialAlt := m.Inputs.AltMode
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})
		time.Sleep(10 * time.Millisecond)
		if m.Inputs.AltMode == initialAlt {
			t.Error("expected AltMode toggled on Tab")
		}
	})

	t.Run("Esc Input", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.Mode = tuictx.InputSearch
		tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
		time.Sleep(10 * time.Millisecond)
		if m.UI.InputActive {
			t.Error("expected input to be inactive after Esc")
		}
	})

	t.Run("Esc InputFuzzySearch", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.Mode = tuictx.InputFuzzySearch
		m.Search.Results = []core.FileResult{{Path: "test"}}
		tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
		time.Sleep(10 * time.Millisecond)
		if m.UI.InputActive {
			t.Error("expected input to be inactive after Esc")
		}
		if m.Search.Results != nil {
			t.Error("expected search results to be cleared after Esc")
		}
	})

	_ = tm.Quit()
}
