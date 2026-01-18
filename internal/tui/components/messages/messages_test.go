package messages

import (
	"strings"
	"testing"

	"fm/internal/constants"
	"fm/internal/testutil"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestMessages_Render(t *testing.T) {
	styles := theme.GetStylesheet(0)

	t.Run("Alert", func(t *testing.T) {
		props := Props{
			Mode:    ModeAlert,
			Width:   80,
			Message: "test message",
			Style:   styles,
		}
		v := Render(props)
		stripped := testutil.StripANSI(v)
		if !strings.Contains(stripped, "test message") {
			t.Errorf("expected alert message, got %q", stripped)
		}
	})

	t.Run("Confirm Delete", func(t *testing.T) {
		props := Props{
			Mode:       ModeConfirming,
			ActionType: constants.ActionDelete,
			Width:      80,
			Style:      styles,
		}
		v := Render(props)
		stripped := testutil.StripANSI(v)
		if !strings.Contains(stripped, "Delete selected items") {
			t.Errorf("expected delete prompt, got %q", stripped)
		}
	})

	t.Run("Host Confirm", func(t *testing.T) {
		props := Props{
			Mode:  ModeHostConfirm,
			Width: 80,
			Style: styles,
		}
		v := Render(props)
		stripped := testutil.StripANSI(v)
		if !strings.Contains(stripped, "Add host") {
			t.Errorf("expected host prompt, got %q", stripped)
		}
	})

	t.Run("Input Prompts", func(t *testing.T) {
		props := Props{
			Mode:  ModeSearching,
			Width: 80,
			Style: styles,
		}
		v := Render(props)
		if v == "" {
			t.Error("expected non-empty search prompt")
		}

		props.Mode = ModeGoto
		props.AltMode = true // Remote
		v = Render(props)
		if !strings.Contains(testutil.StripANSI(v), "Go to") {
			t.Error("expected Goto prompt")
		}
	})
}

func TestMessages_ColorizeKeys(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	styles := theme.GetStylesheet(0)
	props := Props{
		Style: styles,
	}

	tests := []struct {
		input    string
		contains string
	}{
		{"Press [Ctrl+C] again", "Press [Ctrl+C] again"},
		{"[Space] Toggle", "[Space] Toggle"},
	}

	for _, tt := range tests {
		v := ColorizeKeys(props, tt.input)
		stripped := testutil.StripANSI(v)
		if stripped != tt.input {
			t.Errorf("ColorizeKeys(%q) stripped = %q, want %q", tt.input, stripped, tt.input)
		}
		if !strings.Contains(v, "\x1b[") {
			t.Errorf("ColorizeKeys(%q) expected ANSI codes, got %q", tt.input, v)
		}
	}
}
