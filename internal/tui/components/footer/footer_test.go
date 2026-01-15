package footer

import (
	"fm/internal/testutil"
	"fm/internal/tui/components/ui"
	"fm/internal/tui/theme"
	"strings"
	"testing"
)

func TestFooter_RenderModes(t *testing.T) {
	styles := theme.GetStylesheet(0)

	tests := []struct {
		name     string
		mode     Mode
		prompt   string
		contains string
	}{
		{"Normal", ModeNormal, "", "Default"},
		{"Progress", ModeProgress, "", "0%"},
		{"Searching", ModeSearching, "", "Filter"},
		{"FuzzySearch", ModeFuzzySearch, "", "Search"},
		{"Renaming", ModeRenaming, "", "Rename"},
		{"Goto", ModeGoto, "", "Go to"},
		{"Message", ModeMessage, "", "Alert"},
		{"Settings", ModeSettings, "", "Back"},
		{"Log", ModeLog, "", "Back"},
		{"Clipboard", ModeClipboard, "", "Back"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := ui.NewInput(styles)
			if tt.prompt != "" {
				ti.Prompt = tt.prompt
			}
			ti.SetValue("test-input")

			props := Props{
				Mode:        tt.mode,
				Width:       80,
				Message:     "Alert",
				Styles:      styles,
				ActiveInput: ti,
			}
			v := Render(props)
			stripped := testutil.StripANSI(v)
			if !strings.Contains(stripped, tt.contains) {
				t.Errorf("mode %v: expected output to contain %q, got %q", tt.mode, tt.contains, stripped)
			}
		})
	}
}

func TestFooter_Stats(t *testing.T) {
	styles := theme.GetStylesheet(0)
	props := Props{
		Mode:       ModeNormal,
		Width:      80,
		Cursor:     5,
		TotalItems: 10,
		Styles:     styles,
	}

	v := Render(props)
	stripped := testutil.StripANSI(v)
	if !strings.Contains(stripped, "6/10") {
		t.Errorf("expected stats 6/10, got %q", stripped)
	}
}
