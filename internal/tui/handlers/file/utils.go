package file

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/files/core"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
)

func GetTargets(m *tui_context.Model) []string {
	var targets []string
	if m.Navigation.SelectedCount > 0 {
		for path, selected := range m.Navigation.SelectedPaths {
			if selected {
				targets = append(targets, path)
			}
		}
	} else if len(m.Navigation.FilteredItems) > 0 {
		cursor := m.Navigation.Cursor
		if cursor < len(m.Navigation.FilteredItems) {
			sel := m.Navigation.FilteredItems[cursor]
			if !sel.IsUp {
				targets = append(targets, sel.Path)
			}
		}
	}
	return targets
}

func FormatDisplayPath(m *tui_context.Model, fs core.FileSystem, path string) string {
	if fs.IsLocal() {
		return path
	}
	host := m.Remote.Host
	user := m.Remote.User
	if host != "" {
		if user != "" {
			return fmt.Sprintf("%s@%s:%s", user, host, path)
		}
		return fmt.Sprintf("%s:%s", host, path)
	}
	return path
}
