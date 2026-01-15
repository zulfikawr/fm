package loading

import (
	"strings"
	"testing"

	"fm/internal/testutil"
	"fm/internal/tui/components/ui"
	"fm/internal/tui/theme"
)

func TestRender(t *testing.T) {
	styles := theme.GetStylesheet(0)
	s := ui.NewSpinner(styles)

	props := Props{
		Width:   80,
		Height:  10,
		Message: "Syncing",
		Spinner: s,
		Style:   styles,
	}

	v := Render(props)
	stripped := testutil.StripANSI(v)

	if !strings.Contains(stripped, "Syncing") {
		t.Error("Loading view should contain message")
	}
}
