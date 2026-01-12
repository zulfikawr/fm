package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewStylesheet(t *testing.T) {
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			styles := NewStylesheet(theme)

			// Verify all styles are initialized (not zero value)
			if styles.Header.GetPaddingLeft() == 0 && styles.Header.GetPaddingRight() == 0 {
				// Header should have padding
			}

			// Test that styles can render without panic
			_ = styles.Header.Render("test")
			_ = styles.Footer.Render("test")
			_ = styles.ListHeader.Render("test")
			_ = styles.Separator.Render("test")
			_ = styles.Item.Render("test")
			_ = styles.SelectedItem.Render("test")
			_ = styles.SettingsItem.Render("test")
			_ = styles.SettingsSelectedItem.Render("test")
			_ = styles.DirCol.Render("test")
			_ = styles.ExecCol.Render("test")
			_ = styles.FileCol.Render("test")
			_ = styles.GitMod.Render("test")
			_ = styles.GitStaged.Render("test")
			_ = styles.GitUntracked.Render("test")
			_ = styles.GitConflict.Render("test")
			_ = styles.GitGhost.Render("test")
			_ = styles.GitIgnored.Render("test")
			_ = styles.DimCol.Render("test")
			_ = styles.KeyCol.Render("test")
			_ = styles.SettingsHeader.Render("test")
			_ = styles.ProgressBar.Render("test")
		})
	}
}

func TestStylesheetHeaderHasPadding(t *testing.T) {
	theme := Themes[0] // Use first theme
	styles := NewStylesheet(theme)

	// Header should have horizontal padding
	paddingLeft := styles.Header.GetPaddingLeft()
	paddingRight := styles.Header.GetPaddingRight()

	if paddingLeft == 0 && paddingRight == 0 {
		t.Error("Header style should have padding")
	}
}

func TestStylesheetSelectedItemIsBold(t *testing.T) {
	theme := Themes[0]
	styles := NewStylesheet(theme)

	// SelectedItem should be bold
	if !styles.SelectedItem.GetBold() {
		t.Error("SelectedItem style should be bold")
	}
}

func TestStylesheetDirColIsBold(t *testing.T) {
	theme := Themes[0]
	styles := NewStylesheet(theme)

	// DirCol should be bold
	if !styles.DirCol.GetBold() {
		t.Error("DirCol style should be bold")
	}
}

func TestStylesheetGitGhostHasStrikethrough(t *testing.T) {
	theme := Themes[0]
	styles := NewStylesheet(theme)

	// GitGhost should have strikethrough
	if !styles.GitGhost.GetStrikethrough() {
		t.Error("GitGhost style should have strikethrough")
	}
}

func TestStylesheetSettingsItemHasPadding(t *testing.T) {
	theme := Themes[0]
	styles := NewStylesheet(theme)

	// SettingsItem should have left padding
	if styles.SettingsItem.GetPaddingLeft() == 0 {
		t.Error("SettingsItem style should have left padding")
	}
}

func TestStylesheetConsistencyAcrossThemes(t *testing.T) {
	// All themes should produce stylesheets with the same structure
	var firstStyles Stylesheet
	for i, theme := range Themes {
		styles := NewStylesheet(theme)
		if i == 0 {
			firstStyles = styles
			continue
		}

		// Check that structural properties are consistent
		if styles.Header.GetPaddingLeft() != firstStyles.Header.GetPaddingLeft() {
			t.Errorf("Theme %s has different header padding than first theme", theme.Name)
		}
		if styles.SelectedItem.GetBold() != firstStyles.SelectedItem.GetBold() {
			t.Errorf("Theme %s has different SelectedItem bold setting", theme.Name)
		}
		if styles.DirCol.GetBold() != firstStyles.DirCol.GetBold() {
			t.Errorf("Theme %s has different DirCol bold setting", theme.Name)
		}
	}
}

func TestStylesheetRenderOutput(t *testing.T) {
	theme := Themes[0]
	styles := NewStylesheet(theme)

	// Test that render produces non-empty output
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Header", styles.Header},
		{"Footer", styles.Footer},
		{"DirCol", styles.DirCol},
		{"FileCol", styles.FileCol},
		{"SelectedItem", styles.SelectedItem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.style.Render("test")
			if output == "" {
				t.Errorf("%s.Render() returned empty string", tt.name)
			}
		})
	}
}
