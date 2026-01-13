package view

import (
	"testing"

	"fm/internal/sshutil"
	"fm/internal/tui/components/footer"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
)

func TestDetermineMode(t *testing.T) {
	s := &ViewState{
		UI:       &state.UIState{},
		Progress: &state.ProgressState{},
	}

	t.Run("Progress priority", func(t *testing.T) {
		s.Progress.Visible = true
		if DetermineMode(s) != footer.ModeProgress {
			t.Error("Expected ModeProgress")
		}
	})

	t.Run("Search mode", func(t *testing.T) {
		s.Progress.Visible = false
		s.UI.InputActive = true
		s.InputMode = state.InputSearch
		if DetermineMode(s) != footer.ModeSearching {
			t.Error("Expected ModeSearching")
		}
	})

	t.Run("Rename mode", func(t *testing.T) {
		s.UI.InputActive = true
		s.InputMode = state.InputRename
		if DetermineMode(s) != footer.ModeRenaming {
			t.Error("Expected ModeRenaming")
		}
	})

	t.Run("Goto mode", func(t *testing.T) {
		s.UI.InputActive = true
		s.InputMode = state.InputGoto
		if DetermineMode(s) != footer.ModeGoto {
			t.Error("Expected ModeGoto")
		}
	})

	t.Run("Auth mode", func(t *testing.T) {
		s.UI.InputActive = true
		s.InputMode = state.InputAuth
		if DetermineMode(s) != footer.ModeAuth {
			t.Error("Expected ModeAuth")
		}
	})

	t.Run("Host confirm mode", func(t *testing.T) {
		s.UI.InputActive = false
		s.UI.HostConfirm = true
		if DetermineMode(s) != footer.ModeHostConfirm {
			t.Error("Expected ModeHostConfirm")
		}
	})

	t.Run("Confirm mode", func(t *testing.T) {
		s.UI.HostConfirm = false
		s.UI.Confirming = true
		if DetermineMode(s) != footer.ModeConfirming {
			t.Error("Expected ModeConfirming")
		}
	})

	t.Run("Message mode", func(t *testing.T) {
		s.UI.Confirming = false
		s.Msg = "msg"
		if DetermineMode(s) != footer.ModeMessage {
			t.Error("Expected ModeMessage")
		}
	})

	t.Run("Settings mode", func(t *testing.T) {
		s.Msg = ""
		s.UI.SettingsOpen = true
		if DetermineMode(s) != footer.ModeSettings {
			t.Error("Expected ModeSettings")
		}
	})
}

func TestMemoizePrompts(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	s := &ViewState{
		UI: &state.UIState{
			Confirming:  true,
			PromptCache: make(map[string]string),
		},
		Clipboard: &state.ClipboardState{},
	}

	memoizePrompts(s, styles)
	if len(s.UI.PromptCache) == 0 {
		t.Error("Expected prompts to be memoized")
	}

	// Test HostConfirm branch
	s.UI.Confirming = false
	s.UI.HostConfirm = true
	s.HostConfirmReq = &sshutil.HostConfirmRequest{Hostname: "host"}
	memoizePrompts(s, styles)
	if len(s.UI.PromptCache) == 0 {
		t.Error("Expected host confirm prompt to be memoized")
	}
}
