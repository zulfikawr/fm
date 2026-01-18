package loading

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"
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
