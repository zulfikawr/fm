package ui

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Picker renders a cycling option selection indicator.
func Picker(value string, styles theme.Stylesheet) string {
	content := fmt.Sprintf("< %s >", value)
	return styles.SecondaryCol.Render(content)
}
