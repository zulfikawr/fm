package handlers

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/file"
	"github.com/zulfikawr/fm/internal/tui/handlers/integration"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

// HandleMouse handles mouse events
func HandleMouse(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	if !m.Config.EnableMouse {
		return nil
	}

	if msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			return handleScrollUp(m)
		case tea.MouseButtonWheelDown:
			return handleScrollDown(m)
		case tea.MouseButtonLeft:
			return handleMouseClick(m, msg)
		}
	}

	return nil
}

func handleScrollUp(m *context.Model) tea.Cmd {
	if m.UI.SettingsOpen {
		if m.Settings.Offset > 0 {
			m.Settings.Offset--
		}
		// Clamp cursor to viewport
		if m.Settings.Cursor > m.Settings.Offset+m.Display.ViewportHeight-1 {
			m.Settings.Cursor = m.Settings.Offset + m.Display.ViewportHeight - 1
		}
		return nil
	}
	if m.UI.LogOpen {
		if m.Logs.Cursor > 0 {
			m.Logs.Cursor--
		}
		m.Logs.Offset = app.ScrollLogs(m.Logs.Cursor, m.Logs.Offset, m.Display.ViewportHeight)
		return nil
	}
	if m.UI.ClipboardOpen {
		if m.Operations.Clipboard.Cursor > 0 {
			m.Operations.Clipboard.Cursor--
		}
		m.Operations.Clipboard.Offset = app.ScrollLogs(m.Operations.Clipboard.Cursor, m.Operations.Clipboard.Offset, m.Display.ViewportHeight)
		return nil
	}
	if m.Inputs.Mode == context.InputFuzzySearch || len(m.Search.Results) > 0 {
		if m.Search.Offset > 0 {
			m.Search.Offset--
		}
		return nil
	}

	if m.Navigation.Offset > 0 {
		m.Navigation.Offset--
	}
	// Ensure cursor stays within viewport when scrolling offset
	if m.Navigation.Cursor >= m.Navigation.Offset+m.Display.ViewportHeight {
		m.Navigation.Cursor = m.Navigation.Offset + m.Display.ViewportHeight - 1
	}
	return nil
}

func handleScrollDown(m *context.Model) tea.Cmd {
	if m.UI.SettingsOpen {
		// Calculate total settings lines roughly
		totalSettingsLines := 50 // keybindings adds a lot
		if m.Settings.Offset < totalSettingsLines {
			m.Settings.Offset++
		}
		// Clamp cursor to viewport
		if m.Settings.Cursor < m.Settings.Offset {
			m.Settings.Cursor = m.Settings.Offset
		}
		return nil
	}
	if m.UI.LogOpen {
		if m.Logs.Cursor < len(m.Logs.Entries)-1 {
			m.Logs.Cursor++
		}
		m.Logs.Offset = app.ScrollLogs(m.Logs.Cursor, m.Logs.Offset, m.Display.ViewportHeight)
		return nil
	}
	if m.UI.ClipboardOpen {
		if m.Operations.Clipboard.Cursor < len(m.Operations.Clipboard.Paths)-1 {
			m.Operations.Clipboard.Cursor++
		}
		m.Operations.Clipboard.Offset = app.ScrollLogs(m.Operations.Clipboard.Cursor, m.Operations.Clipboard.Offset, m.Display.ViewportHeight)
		return nil
	}
	if m.Inputs.Mode == context.InputFuzzySearch || len(m.Search.Results) > 0 {
		// Calculate total lines in search results
		totalLines := 0
		for _, res := range m.Search.Results {
			totalLines++ // Header
			if !res.Collapsed {
				totalLines += len(res.Matches)
			}
			totalLines++ // Empty line after file
		}
		if m.Search.Offset < totalLines-1 {
			m.Search.Offset++
		}
		return nil
	}

	if m.Navigation.Offset < len(m.Navigation.FilteredItems)-m.Display.ViewportHeight {
		m.Navigation.Offset++
	}
	// Ensure cursor stays within viewport when scrolling offset
	if m.Navigation.Cursor < m.Navigation.Offset {
		m.Navigation.Cursor = m.Navigation.Offset
	}
	return nil
}

func handleMouseClick(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	// App Header is 1 line (y=0)
	if msg.Y == 0 {
		// 1. Check for breadcrumb clicks (Left side)
		if !m.UI.SettingsOpen && !m.UI.LogOpen && !m.UI.ClipboardOpen {
			cmd := handleBreadcrumbClick(m, msg.X)
			if cmd != nil {
				return cmd
			}
		}

		// 2. Handle tab clicks
		tabCount := len(m.Tabs)
		if tabCount <= 1 {
			return nil
		}

		// Each tab is "[n]" (3 chars) + " " (1 char)
		tabsWidth := (tabCount * 3) + (tabCount - 1)
		// Tabs are right-aligned. Header has Padding(0, 1)
		// availableWidth = m.Display.Width - 2
		// tabs start at: 1 + (availableWidth - tabsWidth)
		startX := 1 + (m.Display.Width - 2 - tabsWidth)

		if msg.X >= startX && msg.X < startX+tabsWidth {
			tabIdx := (msg.X - startX) / 4
			if tabIdx >= 0 && tabIdx < tabCount {
				return nav.SwitchTab(m, tabIdx+1)
			}
		}
		return nil
	}

	// Body starts at y=1
	if msg.Y < 1 {
		return nil
	}

	// Calculate which item was clicked
	bodyY := msg.Y - 1

	if m.UI.SettingsOpen {
		return handleSettingsClick(m, bodyY)
	}

	if m.UI.LogOpen || m.UI.ClipboardOpen {
		// Handle clicks in these views if needed
		return nil
	}

	if m.Inputs.Mode == context.InputFuzzySearch || len(m.Search.Results) > 0 {
		return handleSearchClick(m, msg)
	}

	headerHeight := 0
	if m.Config.ShowHeader {
		headerHeight = 3
	}

	if bodyY < headerHeight {
		return nil // Clicked on list header
	}

	itemIdx := bodyY - headerHeight + m.Navigation.Offset
	if itemIdx < 0 || itemIdx >= len(m.Navigation.FilteredItems) {
		return nil
	}

	now := time.Now()
	isDoubleClick := itemIdx == m.Display.LastClickIdx && now.Sub(m.Display.LastClickTime) < 500*time.Millisecond

	m.Display.LastClickTime = now
	m.Display.LastClickIdx = itemIdx

	if isDoubleClick {
		// Double click -> Action (Navigate or Open)
		return nav.NavigateToSelected(m)
	}

	// Single click -> Just select it
	m.Navigation.Cursor = itemIdx
	return nil
}

func handleSettingsClick(m *context.Model, bodyY int) tea.Cmd {
	clickedLine := bodyY + m.Settings.Offset
	if clickedLine < 0 {
		return nil
	}

	// Calculate which setting index corresponds to the clicked line
	// This mapping must match buildSettingGroups and renderGroups in internal/tui/components/views/settings.go
	currentIndex := 0
	currentLine := 1 // Initial empty line

	// Group 1: File Operations (6 settings)
	currentLine++ // Header: File Operations
	for i := 0; i < 6; i++ {
		if currentLine == clickedLine {
			return selectAndToggleSetting(m, currentIndex+i)
		}
		currentLine++
	}
	currentIndex += 6

	currentLine++ // Empty line between groups
	// Group 2: Display Options (7 settings)
	currentLine++ // Header: Display Options
	for i := 0; i < 7; i++ {
		if currentLine == clickedLine {
			return selectAndToggleSetting(m, currentIndex+i)
		}
		currentLine++
	}
	currentIndex += 7

	currentLine++ // Empty line between groups
	// Group 3: Appearance (1 setting)
	currentLine++ // Header: Appearance
	for i := 0; i < 1; i++ {
		if currentLine == clickedLine {
			return selectAndToggleSetting(m, currentIndex+i)
		}
		currentLine++
	}

	return nil
}

func selectAndToggleSetting(m *context.Model, idx int) tea.Cmd {
	now := time.Now()
	// Using a high offset for settings click IDs to avoid collisions
	clickID := 0x5E771000 | idx
	isDoubleClick := clickID == m.Display.LastClickIdx && now.Sub(m.Display.LastClickTime) < 500*time.Millisecond

	m.Display.LastClickTime = now
	m.Display.LastClickIdx = clickID

	m.Settings.Cursor = idx
	m.Settings.Offset = app.ScrollSettings(m)

	if isDoubleClick {
		if reload := app.ToggleSetting(idx, m); reload {
			return func() tea.Msg { return messages.ReloadMsg{} }
		}
	}
	return nil
}

func handleSearchClick(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	// Header is at Y=1 (bodyY=0). Data starts at bodyY=2
	bodyY := msg.Y - 1
	if bodyY < 2 {
		return nil
	}

	clickedLine := bodyY - 2 + m.Search.Offset
	if clickedLine < 0 {
		return nil
	}

	currentLine := 0
	for fIdx := range m.Search.Results {
		res := &m.Search.Results[fIdx]

		// File header
		if currentLine == clickedLine {
			// Clicked on file header
			if msg.X >= 0 && msg.X <= 2 {
				// Clicked on arrow
				res.Collapsed = !res.Collapsed
				m.Search.Offset = integration.ScrollSearch(m)
				return nil
			}
			m.Search.CursorFile = fIdx
			m.Search.CursorMatch = -1
			m.Search.Offset = integration.ScrollSearch(m)
			return nil
		}
		currentLine++

		if !res.Collapsed {
			for mIdx := range res.Matches {
				if currentLine == clickedLine {
					// Clicked on a match
					now := time.Now()
					// Reuse LastClickIdx/LastClickTime for search too.
					// Use a unique encoding for search clicks to avoid interference with file list clicks.
					// e.g. (fIdx << 16) | mIdx
					clickID := (fIdx << 16) | mIdx
					isDoubleClick := clickID == m.Display.LastClickIdx && now.Sub(m.Display.LastClickTime) < 500*time.Millisecond

					m.Display.LastClickTime = now
					m.Display.LastClickIdx = clickID

					m.Search.CursorFile = fIdx
					m.Search.CursorMatch = mIdx
					m.Search.Offset = integration.ScrollSearch(m)

					if isDoubleClick {
						// Open file at line
						match := res.Matches[mIdx]
						return file.OpenFileAtLine(m, res.Path, match.Line)
					}
					return nil
				}
				currentLine++
			}
		}

		// Empty line after file results
		if currentLine == clickedLine {
			return nil
		}
		currentLine++
	}

	return nil
}

func handleBreadcrumbClick(m *context.Model, x int) tea.Cmd {
	sep := m.FS.Separator()
	path := m.Navigation.Path
	if path == "" {
		return nil
	}

	// Basic breadcrumb parsing logic (mirrors internal/tui/components/header/breadcrumbs.go)
	parts := strings.Split(path, sep)
	var cleanParts []string
	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}

	currentX := 1 // Header has Padding(0, 1)

	// Determine the root indicator
	rootIndicator := sep
	isArchive := m.Navigation.ParentFS != nil

	if isArchive {
		rootIndicator = m.FS.Address()
	} else {
		// Handle Windows drive
		if len(cleanParts) > 0 && strings.Contains(cleanParts[0], ":") && sep == "\\" {
			rootIndicator = cleanParts[0]
			cleanParts = cleanParts[1:]
		}

		if !m.FS.IsLocal() {
			user := m.FS.User()
			addr := m.FS.Address()
			if user != "" && addr != "" {
				rootIndicator = user + "@" + addr
			} else if addr != "" {
				rootIndicator = addr
			} else {
				rootIndicator = "Remote"
			}
		}
	}

	// Check if root was clicked
	if x >= currentX && x < currentX+len(rootIndicator) {
		if isArchive || !m.FS.IsLocal() {
			return nav.NavigateToPath(m, "/")
		}
		return nav.NavigateToPath(m, rootIndicator)
	}
	currentX += len(rootIndicator)

	targetPath := rootIndicator
	if isArchive || !m.FS.IsLocal() {
		targetPath = "/"
	}

	for _, p := range cleanParts {
		// Separator " > "
		currentX += 3
		if x >= currentX && x < currentX+len(p) {
			newPath := m.FS.Join(targetPath, p)
			return nav.NavigateToPath(m, newPath)
		}
		targetPath = m.FS.Join(targetPath, p)
		currentX += len(p)
	}

	return nil
}
