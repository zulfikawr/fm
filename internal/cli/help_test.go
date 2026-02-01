package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

func TestPrintHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	styles := theme.GetStylesheet(0)
	PrintHelp(styles, "Gruvbox")

	if err := w.Close(); err != nil {
		t.Errorf("Failed to close pipe: %v", err)
	}
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout

	// Strip ANSI codes for easier comparison
	output := testutil.StripANSI(buf.String())

	expectedSubstrings := []string{
		"FM - Terminal File Manager",
		"Usage:",
		"Keybindings:",
		"j/down, k/up",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(output, s) {
			t.Errorf("Expected output to contain %q, but got:\n%s", s, output)
		}
	}
}
