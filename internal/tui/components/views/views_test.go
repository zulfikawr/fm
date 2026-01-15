package views

import (
	"fm/internal/config"
	"fm/internal/files/core"
	"fm/internal/testutil"
	tui_context "fm/internal/tui/context"
	"fm/internal/tui/theme"
	"strings"
	"testing"
)

func TestRenderSettings(t *testing.T) {
	styles := theme.GetStylesheet(0)
	props := SettingsProps{
		Width:  100,
		Height: 20,
		Config: config.DefaultConfig(),
		Style:  styles,
	}

	t.Run("Basic Render", func(t *testing.T) {
		output := RenderSettings(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "File Operations") {
			t.Errorf("Expected 'File Operations' in settings view, got %q", plain)
		}
	})

	t.Run("Footer", func(t *testing.T) {
		output := RenderSettingsFooter(100, 0, styles)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "Navigate") {
			t.Errorf("Expected 'Navigate' in settings footer, got %q", plain)
		}
	})
}

func TestRenderLogs(t *testing.T) {
	styles := theme.GetStylesheet(0)
	props := LogsProps{
		Width:  100,
		Height: 10,
		Style:  styles,
		Logs: []tui_context.LogEntry{
			{ID: "1", Message: "log 1", Type: "Test"},
		},
	}

	t.Run("With Logs", func(t *testing.T) {
		output := RenderLogs(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "log 1") {
			t.Errorf("Expected log message in output, got %q", plain)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		props.Logs = nil
		output := RenderLogs(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "No operations") {
			t.Errorf("Expected empty log message, got %q", plain)
		}
	})

	t.Run("Log Entry with Details", func(t *testing.T) {
		entry := tui_context.LogEntry{
			ID:      "1",
			Message: "msg",
			Type:    "Op",
			Details: "detail line",
		}
		output := renderLogEntry(props, entry, true)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "detail line") {
			t.Errorf("Expected details in selected log entry, got %q", plain)
		}
	})

	t.Run("Logs Footer", func(t *testing.T) {
		output := RenderLogsFooter(100, styles)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "Back") {
			t.Errorf("Expected 'Back' hint in logs footer, got %q", plain)
		}
	})
}

func TestRenderClipboard(t *testing.T) {
	styles := theme.GetStylesheet(0)
	props := ClipboardProps{
		Width:  100,
		Height: 10,
		Style:  styles,
		Paths:  []string{"/test/f1"},
	}

	t.Run("With Paths", func(t *testing.T) {
		output := RenderClipboard(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "/test/f1") {
			t.Errorf("Expected path in clipboard output, got %q", plain)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		props.Paths = nil
		output := RenderClipboard(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "empty") {
			t.Errorf("Expected empty clipboard message, got %q", plain)
		}
	})

	t.Run("Clipboard Footer - Not Empty", func(t *testing.T) {
		output := RenderClipboardFooter(100, false, styles)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "Paste") {
			t.Errorf("Expected 'Paste' hint in clipboard footer, got %q", plain)
		}
	})

	t.Run("Clipboard Footer - Empty", func(t *testing.T) {
		output := RenderClipboardFooter(100, true, styles)
		plain := testutil.StripANSI(output)
		if strings.Contains(plain, "Paste") {
			t.Errorf("Expected NO 'Paste' hint in empty clipboard footer, got %q", plain)
		}
		if !strings.Contains(plain, "Back") {
			t.Errorf("Expected 'Back' hint in empty clipboard footer, got %q", plain)
		}
	})
}

func TestGetSettingsHelp(t *testing.T) {
	t.Run("Valid Index", func(t *testing.T) {
		help := getSettingsHelp(0)
		if !strings.Contains(help, "Show/hide") {
			t.Errorf("Expected help text for index 0, got %q", help)
		}
	})

	t.Run("Invalid Index", func(t *testing.T) {
		help := getSettingsHelp(99)
		if help != "" {
			t.Errorf("Expected empty string for invalid index, got %q", help)
		}
	})
}

func TestRenderSearch(t *testing.T) {
	styles := theme.GetStylesheet(0)
	props := SearchProps{
		Width:  100,
		Height: 10,
		Style:  styles,
		Query:  "test",
		Results: []core.FileResult{
			{FileName: "f1.txt", Matches: []core.Match{{Line: 1, Content: "line 1"}}},
		},
	}

	t.Run("With Results", func(t *testing.T) {
		output := RenderSearch(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "f1.txt") || !strings.Contains(plain, "line 1") {
			t.Errorf("Expected result content in search output, got %q", plain)
		}
	})
}

func TestRenderMatchContent(t *testing.T) {
	styles := theme.GetStylesheet(0)
	content := "hello world"
	matchedIdx := []int{0, 1, 2, 3, 4} // "hello"

	t.Run("Selected", func(t *testing.T) {
		output := renderMatchContent(content, matchedIdx, true, styles)
		plain := testutil.StripANSI(output)
		testutil.AssertEqual(t, "hello world", plain, "Content should match")
	})

	t.Run("Not Selected", func(t *testing.T) {
		output := renderMatchContent(content, matchedIdx, false, styles)
		plain := testutil.StripANSI(output)
		testutil.AssertEqual(t, "hello world", plain, "Content should match")
	})
}
