package header

import (
	"fm/internal/tui/theme"
	"strings"
	"testing"
)

func TestRenderRemote(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	t.Run("Connected", func(t *testing.T) {
		connected := true
		user := "user"
		host := "example.com"

		result := RenderRemote(connected, user, host, styles)

		if !strings.Contains(result, "user@example.com") {
			t.Error("Should contain user@host")
		}
		if strings.Contains(result, "(Disconnected)") {
			t.Error("Should not contain Disconnected")
		}
	})

	t.Run("Disconnected", func(t *testing.T) {
		connected := false
		user := "user"
		host := "example.com"

		result := RenderRemote(connected, user, host, styles)

		if !strings.Contains(result, "user@example.com") {
			t.Error("Should contain user@host")
		}
		if !strings.Contains(result, "(Disconnected)") {
			t.Error("Should contain Disconnected")
		}
	})

	t.Run("No Host", func(t *testing.T) {
		connected := false
		user := ""
		host := ""

		result := RenderRemote(connected, user, host, styles)

		if result != "" {
			t.Error("Should be empty if no host")
		}
	})
}
