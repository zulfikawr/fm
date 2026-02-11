package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

func TestRunAnalyze(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutil.TempDir(t)

	// Create some files and directories
	dir1 := filepath.Join(tempDir, "dir1")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("failed to create dir1: %v", err)
	}
	file1 := filepath.Join(dir1, "file1.txt")
	if err := os.WriteFile(file1, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	file2 := filepath.Join(tempDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("large file content here..."), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	t.Run("Analyze specific path", func(t *testing.T) {
		// Mock args
		args := &Args{
			Args: []string{tempDir},
		}

		err := RunAnalyze(args)
		if err != nil {
			t.Errorf("RunAnalyze failed: %v", err)
		}
	})

	t.Run("Analyze current directory", func(t *testing.T) {
		// Change to tempDir to test "." behavior
		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get working directory: %v", err)
		}
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(oldWd); err != nil {
				t.Errorf("failed to restore working directory: %v", err)
			}
		}()

		args := &Args{
			Args: []string{"."},
		}
		err = RunAnalyze(args)
		if err != nil {
			t.Errorf("RunAnalyze failed: %v", err)
		}
	})

	// Restore stdout
	if err := w.Close(); err != nil {
		t.Errorf("failed to close pipe writer: %v", err)
	}
	os.Stdout = oldStdout
	outputBytes, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Errorf("failed to read from pipe reader: %v", readErr)
	}
	if len(outputBytes) == 0 {
		// Log or check
	}
}

func TestRenderCLIBar(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.Themes[cfg.UI.ThemeIndex]
	styles := theme.NewStylesheet(th)

	tests := []struct {
		name    string
		percent float64
		width   int
		want    string
	}{
		{"Empty", 0.0, 10, "[..........]"},
		{"Full", 1.0, 10, "[##########]"},
		{"Half", 0.5, 10, "[#####.....]"},
		{"Over", 1.5, 10, "[##########]"},
		{"Under", -0.5, 10, "[..........]"},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got := renderCLIBar(tt.percent, tt.width, &styles)
			// Strip ANSI codes for comparison
			cleanGot := stripANSI(got)
			if cleanGot != tt.want {
				t.Errorf("renderCLIBar() = %v, want %v", cleanGot, tt.want)
			}
		})
	}
}

func stripANSI(str string) string {
	var b strings.Builder
	skip := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\x1b' {
			skip = true
			continue
		}
		if skip {
			if (str[i] >= 'a' && str[i] <= 'z') || (str[i] >= 'A' && str[i] <= 'Z') {
				skip = false
			}
			continue
		}
		b.WriteByte(str[i])
	}
	return b.String()
}
