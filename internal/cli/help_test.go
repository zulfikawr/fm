package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestPrintHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	styles := theme.GetStylesheet(0)
	PrintHelp(styles, "Gruvbox")

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout

	output := buf.String()

	expectedSubstrings := []string{
		"FM - Terminal File Manager",
		"Active Theme: Gruvbox",
		"Usage:",
		"Keybindings:",
		"j/down, k/up",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(output, s) {
			t.Errorf("Expected output to contain %q", s)
		}
	}
}
