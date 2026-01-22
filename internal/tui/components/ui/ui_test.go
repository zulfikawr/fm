package ui

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/theme"
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
		v := ProgressBar(ProgressProps{
			Label:   "Copying",
			Percent: 0.5,
			Width:   80,
			Styles:  styles,
		})
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
		v := RenderHeader(HeaderProps{Width: 80, Columns: cols, Gap: 2, Styles: styles})
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
		if Highlight(styles, "x") == "" {
			t.Error("Highlight failed")
		}
	})

	t.Run("Marker", func(t *testing.T) {
		if Marker(ListProps{Selected: false, IsUp: true, IsCursor: false, Styles: styles}) != "    " {
			t.Error("Marker for isUp should be empty space")
		}
		if !strings.Contains(Marker(ListProps{Selected: false, IsUp: false, IsCursor: false, Styles: styles}), "[ ]") {
			t.Error("Marker for unselected should contain [ ]")
		}
		if !strings.Contains(Marker(ListProps{Selected: true, IsUp: false, IsCursor: false, Styles: styles}), "[x]") {
			t.Error("Marker for selected should contain [x]")
		}
		Marker(ListProps{Selected: true, IsUp: false, IsCursor: true, Styles: styles}) // test cursor
	})

	t.Run("ItemRow", func(t *testing.T) {
		v := ItemRow("content", ListProps{Width: 20, IsCursor: false, Styles: styles})
		if !strings.Contains(v, "content") {
			t.Error("ItemRow should contain content")
		}
		ItemRow("content", ListProps{Width: 20, IsCursor: true, Styles: styles}) // test cursor
	})

	t.Run("Toggle", func(t *testing.T) {
		v := Toggle(true, styles)
		if !strings.Contains(testutil.StripANSI(v), "[ON]") {
			t.Error("Toggle ON failed")
		}
		v = Toggle(false, styles)
		if !strings.Contains(testutil.StripANSI(v), "[OFF]") {
			t.Error("Toggle OFF failed")
		}
		ToggleLabeled(ToggleProps{Label: "Label", Value: true, Width: 20, Styles: styles})
	})

	t.Run("Menu/Selectable Rows", func(t *testing.T) {
		v := SelectableRow("content", MenuProps{Width: 20, IsCursor: false, Styles: styles})
		if !strings.Contains(v, "content") {
			t.Error("SelectableRow failed")
		}
		MenuRow(MenuProps{Label: "Label", Value: "Value", Width: 20, IsCursor: false, Styles: styles})
	})
}
