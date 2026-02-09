package footer

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/theme"
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

			activeView := context.ViewMain
			switch tt.mode {
			case ModeSettings:
				activeView = context.ViewSettings
			case ModeHelp:
				activeView = context.ViewHelp
			case ModeLog:
				activeView = context.ViewLogs
			case ModeClipboard:
				activeView = context.ViewClipboard
			case ModeTrash:
				activeView = context.ViewTrash
			case ModeAnalyze:
				activeView = context.ViewAnalyze
			}

			props := Props{
				Mode:       tt.mode,
				ActiveView: activeView,
				Width:      100,
				Status: StatusInfo{
					Message: "Alert",
				},
				Styles: styles,
				Input: InputContext{
					Active: ti,
				},
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
		Mode:  ModeNormal,
		Width: 80,
		Status: StatusInfo{
			Cursor:     5,
			TotalItems: 10,
		},
		Styles: styles,
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
			Mode:  ModeNormal,
			Width: 100,
			Status: StatusInfo{
				SelectedCount: 1,
			},
			Styles: styles,
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
			Mode:  ModeNormal,
			Width: 20,
			Status: StatusInfo{
				SelectedCount: 1,
			},
			Styles: styles,
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

func TestFooter_UpDirStats(t *testing.T) {
	styles := theme.GetStylesheet(0)

	// Case 1: Cursor on "↑ .."
	props := Props{
		Mode:  ModeNormal,
		Width: 80,
		Status: StatusInfo{
			Cursor:     0,
			TotalItems: 3, // ↑ .., file1, file2
			FilteredItems: []core.Item{
				{Name: "↑ ..", State: core.ItemState{IsUp: true}},
				{Name: "file1"},
				{Name: "file2"},
			},
		},
		Styles: styles,
	}

	v := Render(props)
	stripped := testutil.StripANSI(v)
	if !strings.Contains(stripped, "-/2") {
		t.Errorf("expected stats -/2 for up directory, got %q", stripped)
	}

	// Case 2: Cursor on "file1"
	props.Status.Cursor = 1
	v = Render(props)
	stripped = testutil.StripANSI(v)
	if !strings.Contains(stripped, "1/2") {
		t.Errorf("expected stats 1/2 for first file, got %q", stripped)
	}
}
