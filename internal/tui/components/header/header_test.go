package header

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRender(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	t.Run("Settings Header", func(t *testing.T) {
		props := Props{
			Width:        80,
			SettingsOpen: true,
			Styles:       styles,
		}

		result := Render(props)
		if !strings.Contains(result, "Settings") {
			t.Error("Settings header should contain 'Settings'")
		}
	})

	t.Run("File Header Without Tabs", func(t *testing.T) {
		props := Props{
			Width:     80,
			Path:      "/home/user/docs",
			Separator: "/",
			GitBranch: "",
			ReadOnly:  false,
			TabCount:  1,
			ActiveTab: 0,
			Styles:    styles,
		}

		result := Render(props)
		if !strings.Contains(result, "home") || !strings.Contains(result, "docs") {
			t.Error("Header should contain path components")
		}
		if strings.Contains(result, "[1]") {
			t.Error("Single tab should not show tab indicators")
		}
	})

	t.Run("File Header With Multiple Tabs", func(t *testing.T) {
		props := Props{
			Width:     80,
			Path:      "/home/user",
			Separator: "/",
			TabCount:  3,
			ActiveTab: 1,
			Styles:    styles,
		}

		result := Render(props)
		if !strings.Contains(result, "[1]") || !strings.Contains(result, "[2]") || !strings.Contains(result, "[3]") {
			t.Error("Multiple tabs should show tab indicators")
		}
	})

	t.Run("File Header With Remote", func(t *testing.T) {
		props := Props{
			Width:           80,
			Path:            "/home",
			Separator:       "/",
			RemoteConnected: true,
			RemoteUser:      "u",
			RemoteHost:      "h",
			Styles:          styles,
		}

		result := Render(props)
		if !strings.Contains(result, "u@h") {
			t.Error("Header should contain remote info")
		}
	})
}
