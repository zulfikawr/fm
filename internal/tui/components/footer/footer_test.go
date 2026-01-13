package footer

import (
	"strings"
	"testing"

	"fm/internal/files/sorting"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestRender(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	t.Run("Normal Mode", func(t *testing.T) {
		props := Props{
			Mode:     ModeNormal,
			Width:    80,
			SortMode: sorting.SortName,
			Styles:   styles,
		}
		res := Render(props)
		if res == "" {
			t.Error("Expected non-empty footer")
		}
	})

	t.Run("Message Mode", func(t *testing.T) {
		props := Props{
			Mode:    ModeMessage,
			Width:   80,
			Message: "Hello",
			Styles:  styles,
		}
		res := Render(props)
		if !strings.Contains(res, "Hello") {
			t.Error("Expected Hello in message mode")
		}
	})

	t.Run("Searching Mode", func(t *testing.T) {
		input := textinput.New()
		input.SetValue("findme")
		props := Props{
			Mode:        ModeSearching,
			Width:       80,
			ActiveInput: input,
			Styles:      styles,
		}
		res := Render(props)
		if !strings.Contains(res, "findme") {
			t.Error("Expected findme in search mode")
		}
	})

	t.Run("Progress Mode", func(t *testing.T) {
		props := Props{
			Mode:            ModeProgress,
			Width:           80,
			ProgressLabel:   "Working",
			ProgressPercent: 0.5,
			Styles:          styles,
		}
		res := Render(props)
		if !strings.Contains(res, "Working") {
			t.Error("Expected Working in progress mode")
		}
	})

	t.Run("Settings Mode", func(t *testing.T) {
		props := Props{
			Mode:   ModeSettings,
			Width:  80,
			Styles: styles,
		}
		res := Render(props)
		if !strings.Contains(res, "Navigate") {
			t.Error("Expected settings hints")
		}
	})

	t.Run("Goto Mode with Tab Hint", func(t *testing.T) {
		props := Props{
			Mode:    ModeGoto,
			Width:   80,
			AltMode: false,
			Styles:  styles,
		}
		res := Render(props)
		if !strings.Contains(res, "Tab") || !strings.Contains(res, "Remote") {
			t.Errorf("Expected tab hint for remote, got: %s", res)
		}

		props.AltMode = true
		res = Render(props)
		if !strings.Contains(res, "Local") {
			t.Errorf("Expected tab hint for local, got: %s", res)
		}
	})
}
