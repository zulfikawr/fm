package footer

import (
	"strings"
	"testing"

	"fm/internal/testutil"
	"fm/internal/tui/components/ui"
	"fm/internal/tui/theme"
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

func TestFooter_Responsive(t *testing.T) {
	styles := theme.GetStylesheet(0)

	t.Run("Wide width shows shortcuts", func(t *testing.T) {
		props := Props{
			Mode:          ModeNormal,
			Width:         100,
			SelectedCount: 1,
			Styles:        styles,
		}
		v := Render(props)
		stripped := testutil.StripANSI(v)
		if !strings.Contains(stripped, "1 selected") {
			t.Error("expected '1 selected' to be present")
		}
		if !strings.Contains(stripped, "Copy") {
			t.Error("expected shortcuts to be present")
		}
	})

	t.Run("Narrow width hides shortcuts", func(t *testing.T) {
		props := Props{
			Mode:          ModeNormal,
			Width:         20,
			SelectedCount: 1,
			Styles:        styles,
		}
		v := Render(props)
		stripped := testutil.StripANSI(v)
		if !strings.Contains(stripped, "1 selected") {
			t.Error("expected '1 selected' to be present")
		}
		if strings.Contains(stripped, "Copy") {
			t.Error("expected shortcuts to be hidden")
		}
	})
}
