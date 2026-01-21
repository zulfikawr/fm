package nav

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/files/local"
	"github.com/zulfikawr/fm/internal/logger"
	"github.com/zulfikawr/fm/internal/tui/components/file"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleNavigation handles navigation-related messages
func HandleNavigation(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return HandleNavKeys(m, msg)
	case messages.PartialItemsMsg:
		return HandlePartialItems(m, msg)
	case messages.LoadedItemsMsg:
		return FinalizeDirectoryLoad(m, msg)
	case messages.ArchiveEnteredMsg:
		m.UI.Loading = false
		m.Navigation.ParentFS = msg.ParentFS
		m.Navigation.ParentPath = msg.ParentPath
		m.FS = msg.FS
		m.Navigation.Path = "/"
		m.Navigation.PathGen++
		m.Navigation.Cursor = 0
		m.Navigation.Offset = 0
		return Reload(m, false)
	}
	return nil
}

// NavigateToPath handles navigation to a specific directory path
func NavigateToPath(m *tui_context.Model, path string) tea.Cmd {
	if m.Navigation.Path != "" {
		if len(m.Navigation.BackHistory) == 0 || m.Navigation.BackHistory[len(m.Navigation.BackHistory)-1] != m.Navigation.Path {
			m.Navigation.BackHistory = append(m.Navigation.BackHistory, m.Navigation.Path)
			if len(m.Navigation.BackHistory) > 100 {
				m.Navigation.BackHistory = m.Navigation.BackHistory[1:]
			}
		}
	}
	m.Navigation.ForwardHistory = nil

	return NavigateToPathInternal(m, path)
}

// NavigateToPathInternal is a helper for history navigation that skips history pushing
func NavigateToPathInternal(m *tui_context.Model, path string) tea.Cmd {
	info, err := m.FS.Stat(m.Context, path)
	if err != nil {
		return func() tea.Msg { return messages.ErrorMsg{Err: err} }
	}

	if !info.IsDir() {
		return func() tea.Msg { return messages.ErrorMsg{Err: fmt.Errorf("not a directory")} }
	}

	m.Cache.CursorMemory.Put(m.Navigation.Path, m.Navigation.Cursor)
	m.Cache.OffsetMemory.Put(m.Navigation.Path, m.Navigation.Offset)

	oldParent := m.FS.Dir(m.Navigation.Path)
	m.Cache.ItemCache.Unprotect(oldParent)

	m.Navigation.Path = path
	m.Navigation.PathGen++

	newParent := m.FS.Dir(path)
	m.Cache.ItemCache.Protect(newParent)

	m.ClearSelection()

	return Reload(m, false)
}

// SwitchToLocal switches the current filesystem back to local
func SwitchToLocal(m *tui_context.Model, path string) tea.Cmd {
	if m.FS.IsLocal() {
		return NavigateToPath(m, path)
	}

	isShared := false
	for i, t := range m.Tabs {
		if i == m.ActiveTab {
			continue
		}
		if t.FS == m.FS {
			isShared = true
			break
		}
	}

	if !isShared {
		m.FS.Close()
	}

	m.FS = local.NewLocalFS()

	targetPath := path
	if targetPath == "" {
		home, err := m.FS.GetHomeDir()
		if err == nil {
			targetPath = home
		} else {
			targetPath = "/"
		}
	}

	oldParent := m.FS.Dir(m.Navigation.Path)
	m.Cache.ItemCache.Unprotect(oldParent)

	m.Navigation.Path = targetPath
	m.Navigation.PathGen++

	newParent := m.FS.Dir(targetPath)
	m.Cache.ItemCache.Protect(newParent)

	m.Navigation.Offset = 0
	m.Navigation.FilterQuery = ""
	m.Inputs.ActiveInput.Reset()

	return Reload(m, false)
}

// ApplyFilter filters the navigation items based on current search query
func ApplyFilter(m *tui_context.Model) {
	if m.UI.InputActive && m.Inputs.Mode == tui_context.InputSearch {
		m.Navigation.FilterQuery = strings.ToLower(m.Inputs.ActiveInput.Value())
	}
	file.ApplyFilter(m)
	SyncOffset(m)
}

func SyncOffset(m *tui_context.Model) {
	if m.Display.ViewportHeight == 0 {
		return
	}

	cursor := m.Navigation.Cursor
	offset := m.Navigation.Offset
	height := m.Display.ViewportHeight

	if cursor < offset {
		m.Navigation.Offset = cursor
	} else if cursor >= offset+height {
		m.Navigation.Offset = cursor - height + 1
	}
}

func ExitArchive(m *tui_context.Model) tea.Cmd {

	if m.Navigation.ParentFS == nil {

		return nil

	}

	oldFS := m.FS

	m.FS = m.Navigation.ParentFS

	targetPath := m.Navigation.ParentPath

	m.Navigation.ParentFS = nil

	m.Navigation.ParentPath = ""

	oldFS.Close()

	return NavigateToPath(m, targetPath)

}

func HandleGotoFinalize(m *tui_context.Model, input string) tea.Cmd {

	if !m.FS.IsLocal() {

		if m.Inputs.AltMode {

			return SwitchToLocal(m, input)

		}

		isPath := strings.HasPrefix(input, "/") || strings.HasPrefix(input, ".") || strings.HasPrefix(input, "~") || input == ""

		isConnection := strings.Contains(input, "@")

		if isPath && !isConnection {

			return NavigateToPath(m, input)

		}

		return func() tea.Msg { return messages.RemoteGotoMsg{Input: input} }

	}

	isRemote := m.Inputs.AltMode

	if !isRemote {

		isRemote = strings.Contains(input, "@") || (!strings.HasPrefix(input, "/") && !strings.HasPrefix(input, "./") && !strings.HasPrefix(input, "../") && !strings.HasPrefix(input, "~") && strings.Contains(input, "."))

	}

	if isRemote {

		return func() tea.Msg { return messages.RemoteGotoMsg{Input: input} }

	}

	return NavigateToPath(m, input)

}

// WatchDirAction starts watching the current directory

func WatchDirAction(m *tui_context.Model) tea.Cmd {

	if m.Navigation.Path == "" {

		return nil

	}

	// If we switched between local and remote, reset listening state

	if m.FS.IsLocal() == m.Watcher.IsRemote {

		m.Watcher.IsListening = false

		m.Watcher.IsRemote = !m.FS.IsLocal()

	}

	if !m.FS.IsLocal() {

		if m.Watcher.IsListening {

			return nil

		}

		m.Watcher.IsListening = true

		return utils.WatchRemoteDir()

	}

	if m.Watcher.Watcher == nil {

		return nil

	}

	// Update the watched directory if it changed

	if m.Watcher.LastWatched != m.Navigation.Path {

		if m.Watcher.LastWatched != "" {

			if err := m.Watcher.Watcher.Remove(m.Watcher.LastWatched); err != nil {

				logger.Debugf("Failed to remove directory from watcher: %v", err)

			}

		}

		if err := m.Watcher.Watcher.Add(m.Navigation.Path); err != nil {

			return nil

		}

		m.Watcher.LastWatched = m.Navigation.Path

	}

	if m.Watcher.IsListening {

		return nil

	}

	m.Watcher.IsListening = true

	return utils.WatchDir(m.Watcher.Watcher)

}
