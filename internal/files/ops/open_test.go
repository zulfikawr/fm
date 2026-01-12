package ops

import (
	"testing"
)

func TestIsTextFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.txt", true},
		{"test.go", true},
		{"test.md", true},
		{"test.png", false},
		{"test.jpg", false},
		{"test.exe", false},
		{"LICENSE", true}, // no extension
		{"README", true},
	}

	for _, tt := range tests {
		if got := IsTextFile(tt.path); got != tt.expected {
			t.Errorf("IsTextFile(%s) = %v; want %v", tt.path, got, tt.expected)
		}
	}
}

func TestGetOpenCmd(t *testing.T) {
	t.Run("Text file with editor", func(t *testing.T) {
		cmd, isTerminal, err := GetOpenCmd("test.txt", 0) // vim
		if err != nil {
			t.Fatalf("GetOpenCmd failed: %v", err)
		}
		if cmd.Path == "" {
			t.Error("Expected non-empty command path")
		}
		if !isTerminal {
			t.Error("Expected isTerminal to be true for vim")
		}
	})

	t.Run("Non-text file", func(t *testing.T) {
		cmd, isTerminal, err := GetOpenCmd("test.png", 0)
		if err != nil {
			// Might fail on unsupported OS, but linux/darwin/windows should return a cmd
			return
		}
		if isTerminal {
			t.Error("Expected isTerminal to be false for png")
		}
		if cmd == nil {
			t.Error("Expected non-nil cmd")
		}
	})
}

func TestIsTerminalEditor(t *testing.T) {
	if !isTerminalEditor("vim") {
		t.Error("Expected vim to be terminal editor")
	}
	if isTerminalEditor("code") {
		t.Error("Expected code NOT to be terminal editor")
	}
}
