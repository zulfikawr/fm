package handlers_test

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestRouter_GlobalMessages(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("WindowSizeMsg", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		handlers.HandleUpdate(m, msg)
		if m.Display.Width != 100 || m.Display.Height != 50 {
			t.Errorf("expected 100x50, got %dx%d", m.Display.Width, m.Display.Height)
		}
	})

	t.Run("WatchEventMsg", func(t *testing.T) {
		m.Watcher.IsListening = true
		handlers.HandleUpdate(m, messages.WatchEventMsg{})
		if !m.Watcher.IsListening {
			t.Error("expected IsListening to still be true (debouncing)")
		}
		if m.Watcher.DebounceTimer == nil {
			t.Error("expected debounce timer to be set")
		}

		// Simulate debounce timer expiration
		handlers.HandleUpdate(m, messages.DebounceWatchMsg{})
		if m.Watcher.IsListening {
			t.Error("expected IsListening to be false after debounce")
		}
	})

	t.Run("ClearMsg", func(t *testing.T) {
		m.Message.Push("test", false)
		m.Operations.Progress.Show("label")
		handlers.HandleUpdate(m, messages.ClearMsg{})
		if m.Message.Text != "" {
			t.Error("expected message to be cleared")
		}
	})

	t.Run("ErrorMsg", func(t *testing.T) {
		m.UI.Loading = true
		msg := messages.ErrorMsg{Err: errors.New("fail"), LogID: "123"}
		cmd := handlers.HandleUpdate(m, msg)
		if cmd == nil {
			t.Fatal("expected reload command after error")
		}
	})
}

func TestRouter_GlobalKeys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	wrapper := NewTestModelWrapper(m)
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
	wrapper := NewTestModelWrapper(m)
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

func TestRouter_BatchOperations(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("PerformPasteMsg", func(t *testing.T) {
		msg := messages.PerformPasteMsg{OpName: "Copy", Paths: []string{"/src"}, DestDir: "/dest"}
		cmd := handlers.HandleUpdate(m, msg)
		if cmd == nil {
			t.Error("expected non-nil command for Paste")
		}
	})

	t.Run("PerformZipMsg", func(t *testing.T) {
		msg := messages.PerformZipMsg{Targets: []string{"/src"}, Dst: "/out.zip"}
		cmd := handlers.HandleUpdate(m, msg)
		if cmd == nil {
			t.Error("expected non-nil command for Zip")
		}
	})

	t.Run("PerformRenameMsg", func(t *testing.T) {
		msg := messages.PerformRenameMsg{Selected: core.Item{Name: "old"}, OldPath: "/old", NewPath: "/new", NewName: "new"}
		handlers.HandleUpdate(m, msg)
	})

	t.Run("PerformUnzipMsg", func(t *testing.T) {
		msg := messages.PerformUnzipMsg{ZipPath: "/test.zip", Dst: "/dest"}
		cmd := handlers.HandleUpdate(m, msg)
		if cmd == nil {
			t.Error("expected non-nil command for Unzip")
		}
	})
}

func TestRouter_Events(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("RemotePollMsg", func(t *testing.T) {
		handlers.HandleUpdate(m, messages.RemotePollMsg{})
	})

	t.Run("Watcher messages", func(t *testing.T) {
		handlers.HandleUpdate(m, messages.WatcherErrorMsg{Err: errors.New("err")})
		handlers.HandleUpdate(m, messages.WatcherClosedMsg{})
	})

	t.Run("OperationFinishedEventMsg", func(t *testing.T) {
		m.Logs.AddEntry(tuictx.LogEntry{ID: "log1", Message: "Pasting file"})
		msg := messages.OperationFinishedEventMsg{LogID: "log1"}
		handlers.HandleUpdate(m, msg)
		if len(m.Logs.Entries) > 0 && m.Logs.Entries[0].Status != tuictx.StatusSuccess {
			t.Errorf("expected StatusSuccess, got %v", m.Logs.Entries[0].Status)
		}
	})
}
