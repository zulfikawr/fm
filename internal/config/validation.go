package config

import (
	"fm/internal/files/format"
	"fm/internal/files/ops"
	"fm/internal/tui/theme"
	"fmt"
)

// ValidationBounds holds the valid ranges for config fields
// These are derived from actual data lengths
type ValidationBounds struct {
	MaxThemeIndex      int
	MaxDateFormatIndex int
	MaxSizeFormatIndex int
	MaxEditorIndex     int
}

// GetValidationBounds returns bounds derived from actual data
func GetValidationBounds() ValidationBounds {
	return ValidationBounds{
		MaxThemeIndex:      len(theme.Themes) - 1,
		MaxDateFormatIndex: len(format.DateFormats) - 1,
		MaxSizeFormatIndex: len(format.SizeFormats) - 1,
		MaxEditorIndex:     len(ops.Editors) - 1,
	}
}

// Validate checks if config values are within valid ranges
func (c *Config) Validate() error {
	bounds := GetValidationBounds()

	if c.ThemeIndex < 0 || c.ThemeIndex > bounds.MaxThemeIndex {
		return &ValidationError{
			Field:   "theme_index",
			Value:   c.ThemeIndex,
			Message: fmt.Sprintf("must be 0-%d", bounds.MaxThemeIndex),
		}
	}
	if c.DateFormatIndex < 0 || c.DateFormatIndex > bounds.MaxDateFormatIndex {
		return &ValidationError{
			Field:   "date_format_index",
			Value:   c.DateFormatIndex,
			Message: fmt.Sprintf("must be 0-%d", bounds.MaxDateFormatIndex),
		}
	}
	if c.SizeFormatIndex < 0 || c.SizeFormatIndex > bounds.MaxSizeFormatIndex {
		return &ValidationError{
			Field:   "size_format_index",
			Value:   c.SizeFormatIndex,
			Message: fmt.Sprintf("must be 0-%d", bounds.MaxSizeFormatIndex),
		}
	}
	if c.EditorIndex < 0 || c.EditorIndex > bounds.MaxEditorIndex {
		return &ValidationError{
			Field:   "editor_index",
			Value:   c.EditorIndex,
			Message: fmt.Sprintf("must be 0-%d", bounds.MaxEditorIndex),
		}
	}
	return nil
}

// ValidationError indicates invalid input
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s (value: %v): %s", e.Field, e.Value, e.Message)
}
