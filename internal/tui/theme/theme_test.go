package theme

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestGetStylesheet(t *testing.T) {
	t.Run("Default theme", func(t *testing.T) {
		s := GetStylesheet(0)
		if s.Header.GetForeground() == nil {
			t.Error("Expected header foreground color to be set")
		}
	})

	t.Run("Out of bounds index", func(t *testing.T) {
		s := GetStylesheet(999)
		// Should fallback to default (0)
		expected := GetStylesheet(0)
		testutil.AssertEqual(t, expected.Header.GetForeground(), s.Header.GetForeground(), "Should fallback to default theme")
	})
}

func TestNewStylesheet(t *testing.T) {
	theme := Themes[0]
	s := NewStylesheet(theme)

	testutil.AssertEqual(t, theme.Dir, s.DirCol.GetForeground(), "Dir color should match")
}
