package help

import (
	"fm/internal/tui/theme"
	"testing"
)

func TestHelp(t *testing.T) {
	s := theme.NewStylesheet(theme.Themes[0])

	t.Run("Help Screen", func(t *testing.T) {
		Print(s, "Gruvbox")
	})
}
