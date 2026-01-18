package ui

import (
	"strings"
	"testing"

	"fm/internal/testutil"
	"fm/internal/tui/theme"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		s        string
		max      int
		expected string
	}{
		{"hello world", 11, "hello world"},
		{"hello world", 5, "hell…"},
		{"hello world", 2, "he"},
		{"hello world", 0, ""},
	}

	for _, tt := range tests {
		result := Truncate(tt.s, tt.max)
		if result != tt.expected {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.max, result, tt.expected)
		}
	}
}

func TestRender_UI(t *testing.T) {
	styles := theme.GetStylesheet(0)

	t.Run("ProgressBar", func(t *testing.T) {
		v := ProgressBar("Copying", 0.5, 80, styles)
		stripped := testutil.StripANSI(v)
		if !strings.Contains(stripped, "Copying") || !strings.Contains(stripped, "50%") {
			t.Errorf("ProgressBar failed, got: %q", stripped)
		}
	})

	t.Run("Spinner", func(t *testing.T) {
		s := NewSpinner(styles)
		v := s.View()
		if v == "" {
			t.Error("Spinner view should not be empty")
		}
	})

	t.Run("FlexRow and Spacer", func(t *testing.T) {
		v := FlexRow(10, "a", Spacer(2), "b")
		stripped := testutil.StripANSI(v)
		if !strings.Contains(stripped, "a  b") {
			t.Errorf("FlexRow failed, got: %q", stripped)
		}
	})

	t.Run("Window", func(t *testing.T) {
		v := Window("content", WindowProps{Width: 20, Height: 5, Styles: styles})
		if !strings.Contains(v, "content") {
			t.Error("Window should contain content")
		}
	})

	t.Run("JoinWithGaps", func(t *testing.T) {
		v := JoinWithGaps(3, "a", "b")
		if v != "a   b" {
			t.Errorf("JoinWithGaps failed, got %q", v)
		}
	})

	t.Run("Table Header", func(t *testing.T) {
		cols := []Column{{Title: "Col1", Width: 10}}
		v := RenderHeader(80, cols, 2, styles)
		if !strings.Contains(testutil.StripANSI(v), "Col1") {
			t.Error("Table header should contain title")
		}
	})

	t.Run("Text Helpers", func(t *testing.T) {
		if Dim(styles, "x") == "" {
			t.Error("Dim failed")
		}
		if Bold(styles, "x") == "" {
			t.Error("Bold failed")
		}
		if Success(styles, "x") == "" {
			t.Error("Success failed")
		}
		if Error(styles, "x") == "" {
			t.Error("Error failed")
		}
	})
}
