package tui

import (
	"fm/internal/files"
	"os"
	"testing"
)

func TestHelp(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-help-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(&files.LocalFS{}, tmpDir)

	t.Run("Help Screen", func(t *testing.T) {
		PrintHelp(m.styles, "Gruvbox")
	})
}
