package ui

import (
	"fm/internal/tui/theme"
	"testing"
)

func TestInput_Reset(t *testing.T) {
	styles := theme.GetStylesheet(0)
	input := NewInput(styles)

	// Set some state
	input.SetValue("secret")
	input.EchoMode = EchoPassword
	input.Placeholder = "Search..."
	input.SetCursor(3)

	if input.Value() != "secret" {
		t.Errorf("expected value 'secret', got %q", input.Value())
	}
	if input.EchoMode != EchoPassword {
		t.Error("expected EchoPassword mode")
	}

	// Reset
	input.Reset()

	if input.Value() != "" {
		t.Errorf("expected empty value after reset, got %q", input.Value())
	}
	if input.EchoMode != EchoNormal {
		t.Errorf("expected EchoNormal mode after reset, got %v", input.EchoMode)
	}
	if input.Placeholder != "" {
		t.Errorf("expected empty placeholder after reset, got %q", input.Placeholder)
	}
	if input.cursor != 0 {
		t.Errorf("expected cursor at 0 after reset, got %d", input.cursor)
	}
}

func TestInput_EchoModes(t *testing.T) {
	styles := theme.GetStylesheet(0)
	input := NewInput(styles)
	input.SetValue("hello")
	input.focused = false

	// Normal mode
	viewNormal := input.View()
	if viewNormal != "hello" {
		t.Errorf("EchoNormal: expected 'hello', got %q", viewNormal)
	}

	// Password mode
	input.EchoMode = EchoPassword
	input.EchoCharacter = '*'
	viewPassword := input.View()
	if viewPassword != "*****" {
		t.Errorf("EchoPassword: expected '*****', got %q", viewPassword)
	}

	// None mode
	input.EchoMode = EchoNone
	viewNone := input.View()
	if viewNone != "" {
		t.Errorf("EchoNone: expected empty, got %q", viewNone)
	}
}
