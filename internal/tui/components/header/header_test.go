package header

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

func TestHeader_Render(t *testing.T) {
	styles := theme.GetStylesheet(0)

	props := Props{
		Width:     80,
		Path:      "/test/path",
		Separator: "/",
		Git: GitStatusInfo{
			Branch: "main",
		},
		Style: styles,
	}

	v := Render(props)
	stripped := testutil.StripANSI(v)

	if !strings.Contains(stripped, "test") {
		t.Errorf("expected test in header, got:\n%s", stripped)
	}
	if !strings.Contains(stripped, "main") {
		t.Errorf("expected branch main in header, got:\n%s", stripped)
	}
}
