package footer

import (
	"strings"
	"testing"

	"fm/internal/constants"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestBuildConfirmationPrompt(t *testing.T) {
	tests := []struct {
		name       string
		actionType constants.ActionType
		expected   string
	}{
		{"delete", constants.ActionDelete, "Delete selected items?"},
		{"paste", constants.ActionPaste, "Paste"},
		{"conflict", constants.ActionConflict, "exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := Props{
				ActionType:     tt.actionType,
				ClipboardCount: 5,
				ConflictDst:    "/path/to/file.txt",
			}

			result := BuildConfirmationPrompt(props)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected prompt to contain '%s', got: %s", tt.expected, result)
			}
		})
	}
}

func TestColorizeKeys(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{Styles: styles}

	res := ColorizeKeys(props, "Press [q] to quit")
	if !strings.Contains(res, "q") {
		t.Error("Expected colorized key")
	}
}

func TestRenderInputPrompt(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	input := textinput.New()
	input.SetValue("test val")
	props := Props{Width: 80, Styles: styles}

	res := renderInputPrompt(props, input)
	if !strings.Contains(res, "test val") {
		t.Error("Expected input value in prompt")
	}
}

func TestRenderConfirmationPrompt(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Mode:       ModeConfirming,
		Width:      80,
		ActionType: constants.ActionDelete,
		Styles:     styles,
	}

	res := renderConfirmationPrompt(props)
	if !strings.Contains(res, "Delete") {
		t.Error("Expected Delete in confirmation prompt")
	}

	// Test cache
	props.PromptCache = map[string]string{
		"confirm-delete-0-": "Cached Prompt",
	}
	res = renderConfirmationPrompt(props)
	if !strings.Contains(res, "Cached Prompt") {
		t.Error("Expected cached prompt")
	}
}

func TestRenderHostConfirmPrompt(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Mode:   ModeHostConfirm,
		Width:  80,
		Styles: styles,
	}

	res := renderHostConfirmPrompt(props)
	if !strings.Contains(res, "Add host") {
		t.Error("Expected Add host in host confirm prompt")
	}
}
