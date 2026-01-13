package footer

import (
	"os"
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRenderPermissionInfo(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	t.Run("Normal File", func(t *testing.T) {
		items := []core.Item{{Mode: os.FileMode(0644)}}
		res := renderPermissionInfo(items, 0, styles)
		if !strings.Contains(res, "rw-r--r--") {
			t.Errorf("Expected rw-r--r--, got %q", res)
		}
	})

	t.Run("Up Dir", func(t *testing.T) {
		items := []core.Item{{IsUp: true}}
		res := renderPermissionInfo(items, 0, styles)
		if res != "" {
			t.Errorf("Expected empty string for up dir, got %q", res)
		}
	})

	t.Run("Invalid Cursor", func(t *testing.T) {
		res := renderPermissionInfo(nil, 0, styles)
		if res != "" {
			t.Errorf("Expected empty string for invalid cursor, got %q", res)
		}
	})
}
