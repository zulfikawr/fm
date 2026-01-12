package config

import (
	"fm/internal/files/format"
	"fm/internal/files/ops"
	"fm/internal/tui/theme"
	"pgregory.net/rapid"
	"strings"
	"testing"
)

func TestValidationBounds_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 3: Config Validation Bounds Derived from Data
	rapid.Check(t, func(t *rapid.T) {
		bounds := GetValidationBounds()

		if bounds.MaxThemeIndex != len(theme.Themes)-1 {
			t.Errorf("MaxThemeIndex expected %d, got %d", len(theme.Themes)-1, bounds.MaxThemeIndex)
		}
		if bounds.MaxDateFormatIndex != len(format.DateFormats)-1 {
			t.Errorf("MaxDateFormatIndex expected %d, got %d", len(format.DateFormats)-1, bounds.MaxDateFormatIndex)
		}
		if bounds.MaxSizeFormatIndex != len(format.SizeFormats)-1 {
			t.Errorf("MaxSizeFormatIndex expected %d, got %d", len(format.SizeFormats)-1, bounds.MaxSizeFormatIndex)
		}
		if bounds.MaxEditorIndex != len(ops.Editors)-1 {
			t.Errorf("MaxEditorIndex expected %d, got %d", len(ops.Editors)-1, bounds.MaxEditorIndex)
		}
	})
}

func TestValidationErrorFieldNames_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 11: Validation Error Field Names
	rapid.Check(t, func(t *rapid.T) {
		bounds := GetValidationBounds()
		cfg := DefaultConfig()

		// Randomly pick one field to make invalid
		field := rapid.SampledFrom([]string{"theme_index", "date_format_index", "size_format_index", "editor_index"}).Draw(t, "field")

		switch field {
		case "theme_index":
			cfg.ThemeIndex = bounds.MaxThemeIndex + 1 + rapid.IntRange(0, 1000).Draw(t, "val")
		case "date_format_index":
			cfg.DateFormatIndex = bounds.MaxDateFormatIndex + 1 + rapid.IntRange(0, 1000).Draw(t, "val")
		case "size_format_index":
			cfg.SizeFormatIndex = bounds.MaxSizeFormatIndex + 1 + rapid.IntRange(0, 1000).Draw(t, "val")
		case "editor_index":
			cfg.EditorIndex = bounds.MaxEditorIndex + 1 + rapid.IntRange(0, 1000).Draw(t, "val")
		}

		err := cfg.Validate()
		if err == nil {
			t.Fatalf("Expected validation error for field %s", field)
		}

		if !strings.Contains(err.Error(), field) {
			t.Errorf("Error message should contain field name %s, got: %v", field, err)
		}
	})
}
