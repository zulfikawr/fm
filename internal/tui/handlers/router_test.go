package handlers_test

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
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
		m.Navigation.Search.Results = []core.FileResult{{Path: "test", Matches: []core.Match{{Line: 1}}}}
		tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		time.Sleep(10 * time.Millisecond)
		if m.UI.InputActive {
			t.Error("expected input to be inactive after Enter")
		}
		if m.Navigation.Search.Results != nil {
			t.Error("expected search results to be cleared after Enter")
		}
	})

	t.Run("Input Tab Toggle", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.Mode = tuictx.InputCreate
		initialAlt := m.Inputs.AltMode
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})
		time.Sleep(10 * time.Millisecond)
		if m.Inputs.AltMode == initialAlt {
			t.Error("expected AltMode toggled on Tab")
		}
	})

	t.Run("Goto command shows initial prompt", func(t *testing.T) {
		m.UI.StopConfirming()
		m.StopInput(true)
		m.Operations.ActionType = constants.ActionNone

		// Press 'g'
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
		time.Sleep(10 * time.Millisecond)

		if !m.UI.Confirming {
			t.Error("expected confirming state after 'g'")
		}
		if m.Operations.ActionType != constants.ActionGoto {
			t.Errorf("expected ActionType 'goto', got %q", m.Operations.ActionType)
		}

		// Press 'l' for local
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
		time.Sleep(10 * time.Millisecond)

		if m.UI.Confirming {
			t.Error("expected confirmation to end after choosing 'l'")
		}
		if !m.UI.InputActive || m.Inputs.Mode != tuictx.InputGoto {
			t.Error("expected InputGoto to start after choosing 'l'")
		}
		if m.Inputs.AltMode {
			t.Error("expected AltMode to be false for Local")
		}

		// Cleanup and test 'r' for remote
		m.StopInput(true)
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
		time.Sleep(10 * time.Millisecond)
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
		time.Sleep(10 * time.Millisecond)

		if !m.UI.InputActive || m.Inputs.Mode != tuictx.InputGoto {
			t.Error("expected InputGoto to start after choosing 'r'")
		}
		if !m.Inputs.AltMode {
			t.Error("expected AltMode to be true for Remote")
		}
	})

	t.Run("Auth command shows initial prompt", func(t *testing.T) {
		m.UI.StopConfirming()
		m.StopInput(true)
		m.UI.RemoteAuth = true
		m.Operations.ActionType = constants.ActionAuth
		m.UI.StartConfirming()

		// Press 'p' for password
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
		time.Sleep(10 * time.Millisecond)

		if m.UI.Confirming {
			t.Error("expected confirmation to end after choosing 'p'")
		}
		if !m.UI.InputActive || m.Inputs.Mode != tuictx.InputAuth {
			t.Error("expected InputAuth to start after choosing 'p'")
		}
		if m.Inputs.AltMode {
			t.Error("expected AltMode to be false for Password")
		}
		if m.Inputs.ActiveInput.EchoMode != ui.EchoPassword {
			t.Error("expected EchoPassword mode for Password auth")
		}

		// Cleanup and test 'k' for key
		m.StopInput(true)
		m.UI.RemoteAuth = true
		m.Operations.ActionType = constants.ActionAuth
		m.UI.StartConfirming()

		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		time.Sleep(10 * time.Millisecond)

		if !m.UI.InputActive || m.Inputs.Mode != tuictx.InputAuth {
			t.Error("expected InputAuth to start after choosing 'k'")
		}
		if !m.Inputs.AltMode {
			t.Error("expected AltMode to be true for Key")
		}
		if m.Inputs.ActiveInput.EchoMode != ui.EchoNormal {
			t.Error("expected EchoNormal mode for Key Path auth")
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
		m.Navigation.Search.Results = []core.FileResult{{Path: "test"}}
		tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
		time.Sleep(10 * time.Millisecond)
		if m.UI.InputActive {
			t.Error("expected input to be inactive after Esc")
		}
		if m.Navigation.Search.Results != nil {
			t.Error("expected search results to be cleared after Esc")
		}
	})

	t.Run("Dot key does not open settings during input", func(t *testing.T) {
		m.StartInput(tuictx.InputSearch)
		m.Inputs.ActiveInput.Focus()
		m.UI.SettingsOpen = false

		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(".")})
		time.Sleep(10 * time.Millisecond)

		if m.UI.SettingsOpen {
			t.Error("expected settings to remain closed when input is active")
		}
		if m.Inputs.ActiveInput.Value() != "." {
			t.Errorf("expected input value to be '.', got %q", m.Inputs.ActiveInput.Value())
		}
	})

	t.Run("Question mark does not open help during input", func(t *testing.T) {
		m.StartInput(tuictx.InputSearch)
		m.Inputs.ActiveInput.Focus()
		m.UI.HelpOpen = false

		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
		time.Sleep(10 * time.Millisecond)

		if m.UI.HelpOpen {
			t.Error("expected help to remain closed when input is active")
		}
		if m.Inputs.ActiveInput.Value() != "?" {
			t.Errorf("expected input value to be '?', got %q", m.Inputs.ActiveInput.Value())
		}
	})

	t.Run("Dot and Question mark still work when input is NOT active", func(t *testing.T) {
		m.StopInput(true)
		m.UI.SettingsOpen = false
		m.UI.HelpOpen = false

		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(".")})
		time.Sleep(10 * time.Millisecond)
		if !m.UI.SettingsOpen {
			t.Error("expected settings to open")
		}
		m.UI.SettingsOpen = false

		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
		time.Sleep(10 * time.Millisecond)
		if !m.UI.HelpOpen {
			t.Error("expected help to open")
		}
		m.UI.HelpOpen = false
	})

	t.Run("Up and Down arrows navigate during filtering", func(t *testing.T) {
		m.StartInput(tuictx.InputSearch)
		m.Inputs.ActiveInput.Focus()
		m.Navigation.FilteredItems = []core.Item{
			{Name: "file1", Path: "/test/file1"},
			{Name: "file2", Path: "/test/file2"},
			{Name: "file3", Path: "/test/file3"},
		}
		m.Navigation.Cursor = 1 // Start at second item

		// Press Down
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})
		time.Sleep(10 * time.Millisecond)
		if m.Navigation.Cursor != 2 {
			t.Errorf("expected cursor to be 2 after Down arrow, got %d", m.Navigation.Cursor)
		}

		// Press Up
		tm.Send(tea.KeyMsg{Type: tea.KeyUp})
		time.Sleep(10 * time.Millisecond)
		if m.Navigation.Cursor != 1 {
			t.Errorf("expected cursor to be 1 after Up arrow, got %d", m.Navigation.Cursor)
		}

		if !m.UI.InputActive {
			t.Error("expected input to remain active during navigation")
		}
	})

	t.Run("Tab autocompletes during filtering", func(t *testing.T) {
		m.StartInput(tuictx.InputSearch)
		m.Inputs.ActiveInput.Focus()
		item := core.Item{Name: "gitignore", Path: "/test/.gitignore"}
		m.Navigation.Items = []core.Item{item}
		m.Navigation.FilteredItems = []core.Item{item}
		m.Navigation.Cursor = 0
		m.Inputs.ActiveInput.SetValue("git")
		utils.UpdateSearchSuggestion(m)

		if m.Inputs.ActiveInput.Suggestion != "gitignore" {
			t.Errorf("expected suggestion to be 'gitignore', got %q", m.Inputs.ActiveInput.Suggestion)
		}

		// Press Tab
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})
		time.Sleep(10 * time.Millisecond)

		if m.Inputs.ActiveInput.Value() != "gitignore" {
			t.Errorf("expected value to be 'gitignore' after Tab, got %q", m.Inputs.ActiveInput.Value())
		}
		if m.Inputs.ActiveInput.Suggestion != "" {
			t.Error("expected suggestion to be cleared after Tab")
		}
	})

	t.Run("CompletionsMsg updates suggestion", func(t *testing.T) {
		m.StartInput(tuictx.InputGoto)
		m.Inputs.ActiveInput.SetValue("/h")

		msg := messages.CompletionsMsg{Completions: []string{"/home/"}}
		handlers.HandleUpdate(m, msg)

		if m.Inputs.ActiveInput.Suggestion != "/home/" {
			t.Errorf("expected suggestion to be '/home/', got %q", m.Inputs.ActiveInput.Suggestion)
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
