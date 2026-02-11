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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	styles := theme.GetStylesheet(0)
	PrintHelp(styles, "Gruvbox")

	if err := w.Close(); err != nil {
		t.Errorf("Failed to close pipe: %v", err)
	}
	var buf bytes.Buffer
	n, err := io.Copy(&buf, r)
	if err != nil {
		t.Errorf("failed to copy from pipe (copied %d bytes): %v", n, err)
	}
	os.Stdout = oldStdout

	// Strip ANSI codes for easier comparison
	output := testutil.StripANSI(buf.String())

	expectedSubstrings := []string{
		"fm - Terminal File Manager",
		"Usage:",
		"General:",
		"Navigation:",
		"File Operations:",
		"Selection:",
		"Search & Filter:",
		"Tabs:",
	}

	for i := range expectedSubstrings {
		s := expectedSubstrings[i]
		if !strings.Contains(output, s) {
			t.Errorf("Expected output to contain %q, but got:\n%s", s, output)
		}
	}
}
