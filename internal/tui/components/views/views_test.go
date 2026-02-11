package views

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/theme"
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
		items := []SettingHelpItem{{HelpText: "Test help text"}}
		output := RenderSettingsFooter(100, 0, items, styles)
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
		Logs: []tuictx.LogEntry{
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
		entry := tuictx.LogEntry{
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
		output := renderMatchContent(MatchProps{
			Content:    content,
			MatchedIdx: matchedIdx,
			IsSelected: true,
			Style:      styles,
		})
		plain := testutil.StripANSI(output)
		testutil.AssertEqual(t, "hello world", plain, "Content should match")
	})

	t.Run("Not Selected", func(t *testing.T) {
		output := renderMatchContent(MatchProps{
			Content:    content,
			MatchedIdx: matchedIdx,
			IsSelected: false,
			Style:      styles,
		})
		plain := testutil.StripANSI(output)
		testutil.AssertEqual(t, "hello world", plain, "Content should match")
	})
}

func TestRenderAnalyze(t *testing.T) {
	styles := theme.GetStylesheet(0)
	root := &core.AnalysisResult{
		Name:        "root",
		Path:        "/root",
		IsDirectory: true,
		Size:        1000,
		Percentage:  1.0,
		Children: []*core.AnalysisResult{
			{
				Name:        "child1",
				Path:        "/root/child1",
				IsDirectory: false,
				Size:        500,
				Percentage:  0.5,
			},
		},
	}

	props := AnalyzeProps{
		Width:           100,
		Height:          20,
		ActiveNode:      root,
		Style:           styles,
		EnableIcons:     true,
		SizeFormatIndex: 0,
	}

	t.Run("Basic Render", func(t *testing.T) {
		output := RenderAnalyze(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "child1") {
			t.Errorf("Expected 'child1' in analyze view, got %q", plain)
		}
		if !strings.Contains(plain, "50.0%") {
			t.Errorf("Expected percentage in analyze view, got %q", plain)
		}
	})

	t.Run("Empty Node", func(t *testing.T) {
		props.ActiveNode = nil
		output := RenderAnalyze(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "No analysis data") {
			t.Errorf("Expected no data message, got %q", plain)
		}
	})

	t.Run("With Up Entry", func(t *testing.T) {
		parent := &core.AnalysisResult{Name: "parent", Path: "/"}
		root.Parent = parent
		props.ActiveNode = root
		props.IsRoot = false
		output := RenderAnalyze(props)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "↑ ..") {
			t.Errorf("Expected '↑ ..' in analyze view, got %q", plain)
		}
	})

	t.Run("Analyze Footer", func(t *testing.T) {
		output := RenderAnalyzeFooter(100, styles)
		plain := testutil.StripANSI(output)
		if !strings.Contains(plain, "Back") || !strings.Contains(plain, "Delete") {
			t.Errorf("Expected navigation hints in analyze footer, got %q", plain)
		}
	})
}
