package messages

import (
	"fm/internal/constants"
	"fm/internal/testutil"
	"fm/internal/tui/theme"
	"strings"
	"testing"
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
