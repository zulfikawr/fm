package footer

import (
	"strings"
	"testing"

	"fm/internal/constants"
	"fm/internal/files"
	"fm/internal/files/sorting"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestRender_Message(t *testing.T) {
	props := Props{
		Mode:    ModeMessage,
		Width:   80,
		Message: "Test message",
		Styles:  theme.NewStylesheet(theme.Themes[0]),
	}

	result := Render(props)

	if !strings.Contains(result, "Test message") {
		t.Error("Expected footer to contain message")
	}
}

func TestRender_Normal(t *testing.T) {
	props := Props{
		Mode:          ModeNormal,
		Width:         80,
		SortMode:      sorting.SortDefault,
		Cursor:        0,
		FilteredItems: []files.Item{{Name: "test.txt"}},
		Items:         []files.Item{{Name: "test.txt"}},
		Styles:        theme.NewStylesheet(theme.Themes[0]),
	}

	result := Render(props)

	if result == "" {
		t.Error("Expected non-empty footer")
	}
}

func TestRender_Input(t *testing.T) {
	input := textinput.New()
	input.SetValue("search term")

	props := Props{
		Mode:        ModeSearching,
		Width:       80,
		ActiveInput: input,
		Styles:      theme.NewStylesheet(theme.Themes[0]),
	}

	result := Render(props)

	if !strings.Contains(result, "search term") {
		t.Error("Expected footer to contain search input")
	}
}

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
