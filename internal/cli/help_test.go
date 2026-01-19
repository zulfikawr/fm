package cli

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/tui/theme"
)

var ansiRegex = regexp.MustCompile("[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")

func stripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

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

	// Strip ANSI codes for easier comparison
	output := stripANSI(buf.String())

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
