package settings

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderSettingRow renders a single setting or keybinding row
func renderSettingRow(props Props, sItem SettingItem, isCursor bool) string {
	style := props.Styles.SettingsItem
	if isCursor {
		style = props.Styles.SettingsSelectedItem
	}

	val := sItem.Value
	if sItem.Inactive {
		style = props.Styles.DimCol.PaddingLeft(2)
		val = props.Styles.DimCol.Render(sItem.Value)
	}

	labelWidth := 35
	if props.Width < 60 {
		labelWidth = props.Width - 20
	}
	if labelWidth < 10 {
		labelWidth = 10
	}

	label := sItem.Label + ":"
	if len(label) > labelWidth {
		label = label[:labelWidth-1] + "…"
	}

	// Calculate width without ANSI codes for proper alignment
	valWidth := lipgloss.Width(sItem.Value)
	content := fmt.Sprintf("% -*s %s", labelWidth, label, val)

	if labelWidth+1+valWidth > props.Width-2 {
		// If still too long, prioritize showing the value
		availableLabelWidth := props.Width - 2 - valWidth - 1
		if availableLabelWidth > 0 {
			label = sItem.Label + ":"
			if len(label) > availableLabelWidth {
				label = label[:availableLabelWidth-1] + "…"
			}
			content = fmt.Sprintf("% -*s %s", availableLabelWidth, label, val)
		}
	}

	return style.Width(props.Width).Render(content)
}
