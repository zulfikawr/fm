package view

import (
	"github.com/zulfikawr/fm/internal/tui/context"
)

func formatRemoteStr(m *context.Model) string {
	if m.FS.IsLocal() {
		return ""
	}
	user := m.FS.User()
	addr := m.FS.Address()

	if user != "" && addr != "" {
		return user + "@" + addr
	}
	if addr != "" {
		return addr
	}
	return "Remote"
}

func formatArchiveRoot(m *context.Model) string {
	if m.Navigation.ParentFS == nil {
		return ""
	}
	return m.FS.Address()
}
