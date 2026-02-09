package config

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/constants"
)

// ValidationBounds holds the valid ranges for config fields
type ValidationBounds struct {
	MaxThemeIndex      int
	MaxDateFormatIndex int
	MaxSizeFormatIndex int
	MaxEditorIndex     int
}

// GetValidationBounds returns bounds derived from constants
func GetValidationBounds() ValidationBounds {
	return ValidationBounds{
		MaxThemeIndex:      constants.ThemeCount - 1,
		MaxDateFormatIndex: constants.DateFormatCount - 1,
		MaxSizeFormatIndex: constants.SizeFormatCount - 1,
		MaxEditorIndex:     len(constants.Editors) - 1,
	}
}

// Validate checks if config values are within valid ranges
func (c *Config) Validate() error {
	bounds := GetValidationBounds()

	if c.UI.ThemeIndex < 0 || c.UI.ThemeIndex > bounds.MaxThemeIndex {
		return &ValidationError{
			Field:   "theme_index",
			Value:   c.UI.ThemeIndex,
			Message: fmt.Sprintf("must be 0-%d", bounds.MaxThemeIndex),
		}
	}
	if c.UI.DateFormatIndex < 0 || c.UI.DateFormatIndex > bounds.MaxDateFormatIndex {
		return &ValidationError{
			Field:   "date_format_index",
			Value:   c.UI.DateFormatIndex,
			Message: fmt.Sprintf("must be 0-%d", bounds.MaxDateFormatIndex),
		}
	}
	if c.UI.SizeFormatIndex < 0 || c.UI.SizeFormatIndex > bounds.MaxSizeFormatIndex {
		return &ValidationError{
			Field:   "size_format_index",
			Value:   c.UI.SizeFormatIndex,
			Message: fmt.Sprintf("must be 0-%d", bounds.MaxSizeFormatIndex),
		}
	}
	if c.External.EditorIndex < 0 || c.External.EditorIndex > bounds.MaxEditorIndex {
		return &ValidationError{
			Field:   "editor_index",
			Value:   c.External.EditorIndex,
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
