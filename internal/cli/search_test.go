package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestHighlightMatches(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	style := lipgloss.NewStyle().Bold(true)

	t.Run("No indices", func(t *testing.T) {
		content := "hello world"
		got := highlightMatches(content, nil, style)
		if got != content {
			t.Errorf("expected %q, got %q", content, got)
		}
	})

	t.Run("With indices", func(t *testing.T) {
		content := "hello"
		got := highlightMatches(content, []int{0, 4}, style)
		stripped := testutil.StripANSI(got)
		if stripped != content {
			t.Errorf("stripped content mismatch: got %q, want %q", stripped, content)
		}
		if !strings.Contains(got, "\x1b[") {
			t.Error("expected ANSI escape codes in output")
		}
	})
}

func TestHighlightMatchesWithBase(t *testing.T) {
	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	matchStyle := lipgloss.NewStyle().Bold(true)

	t.Run("No indices", func(t *testing.T) {
		content := "test"
		got := highlightMatchesWithBase(content, nil, matchStyle, baseStyle)
		stripped := testutil.StripANSI(got)
		if stripped != content {
			t.Errorf("expected %q, got %q", content, stripped)
		}
	})

	t.Run("With indices", func(t *testing.T) {
		content := "test"
		got := highlightMatchesWithBase(content, []int{1}, matchStyle, baseStyle)
		stripped := testutil.StripANSI(got)
		if stripped != content {
			t.Errorf("expected %q, got %q", content, stripped)
		}
	})
}

func TestRunSearch(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	filePath := filepath.Join(tmpDir, "test.txt")
	content := "This is a search test\nLine with match\nAnother line"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("Empty Query", func(t *testing.T) {
		args := &Args{SearchQuery: ""}
		err := RunSearch(args)
		if err != nil {
			t.Errorf("RunSearch with empty query should not return error, got %v", err)
		}
	})

	t.Run("Successful search", func(t *testing.T) {
		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		args := &Args{
			SearchQuery: "match",
			Args:        []string{tmpDir},
			IsSearch:    true,
		}
		err := RunSearch(args)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := testutil.StripANSI(buf.String())

		if err != nil {
			t.Errorf("RunSearch failed: %v", err)
		}
		if !strings.Contains(output, "test.txt") {
			t.Error("expected filename in output")
		}
		if !strings.Contains(output, "Line with match") {
			t.Error("expected matching line in output")
		}
	})

	t.Run("No matches", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		args := &Args{
			SearchQuery: "nonexistent",
			Args:        []string{tmpDir},
			IsSearch:    true,
		}
		_ = RunSearch(args)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := testutil.StripANSI(buf.String())

		if !strings.Contains(output, "No matches found") {
			t.Error("expected 'No matches found' in output")
		}
	})
}
