package handlers

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	footer_comp "github.com/zulfikawr/fm/internal/tui/components/footer"
	msg_comp "github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/file"
	"github.com/zulfikawr/fm/internal/tui/handlers/integration"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

// HandleMouse handles mouse events
func HandleMouse(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	if !m.Config.UI.EnableMouse {
		return nil
	}

	// 1. App Header is 1 line (y=0)
	if msg.Y == 0 {
		return handleMouseClick(m, msg) // Existing header logic
	}

	// 2. Analyze Handlers (High Priority Modal)
	if m.UI.ActiveView == context.ViewAnalyze {
		return HandleAnalyze(m, msg)
	}

	switch msg.Action {
	case tea.MouseActionPress:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			return handleScrollUp(m)
		case tea.MouseButtonWheelDown:
			return handleScrollDown(m)
		case tea.MouseButtonLeft:
			return handleMousePress(m, msg)
		}
	case tea.MouseActionMotion:
		if msg.Button == tea.MouseButtonLeft {
			// Only allow dragging in the main file view, not in other views
			if m.UI.ActiveView == context.ViewMain {
				return handleMouseDrag(m, msg)
			}
		}
	case tea.MouseActionRelease:
		if msg.Button == tea.MouseButtonLeft {
			// Only allow releasing drag in the main file view, not in other views
			if m.UI.ActiveView == context.ViewMain {
				return handleMouseRelease(m, msg)
			}
		}
	}

	return nil
}

func handleScrollUp(m *context.Model) tea.Cmd {
	if m.UI.ActiveView == context.ViewSettings {
		if m.Settings.Offset > 0 {
			m.Settings.Offset--
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewHelp {
		if m.Help.Offset > 0 {
			m.Help.Offset--
		}
		// Clamp cursor to viewport
		if m.Help.Cursor > m.Help.Offset+m.Display.ViewportHeight-1 {
			m.Help.Cursor = m.Help.Offset + m.Display.ViewportHeight - 1
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewLogs {
		if m.Logs.Offset > 0 {
			m.Logs.Offset--
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewClipboard {
		if m.Operations.Clipboard.Offset > 0 {
			m.Operations.Clipboard.Offset--
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewAnalyze {
		if m.Analyze.Offset > 0 {
			m.Analyze.Offset--
		}
		return nil
	}
	if m.Inputs.Mode == context.InputFuzzySearch || len(m.Navigation.Search.Results) > 0 {
		if m.Navigation.Search.Offset > 0 {
			m.Navigation.Search.Offset--
		}
		return nil
	}

	if m.Navigation.Offset > 0 {
		m.Navigation.Offset--
	}
	return nil
}

func handleScrollDown(m *context.Model) tea.Cmd {
	if m.UI.ActiveView == context.ViewSettings {
		// Total settings rows = 1 (top empty) + 1 (header) + 6 (opts) + 1 (empty) + 1 (header) + 7 (opts) + 1 (empty) + 1 (header) + 1 (opt) + 1 (empty) + 1 (header) + 25 (keys) = 47
		totalSettingsLines := 47
		if m.Settings.Offset < totalSettingsLines-m.Display.ViewportHeight {
			m.Settings.Offset++
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewHelp {
		// Total help rows = 1 (top) + (7+1 header) + 1 spacer + (4+1) + 1 + (3+1) + 1 + (8+1) + 1 + (4+1) + 1 + (6+1) = 43 approx
		totalHelpLines := 43
		if m.Help.Offset < totalHelpLines-m.Display.ViewportHeight {
			m.Help.Offset++
		}
		// Clamp cursor to viewport
		if m.Help.Cursor < m.Help.Offset {
			m.Help.Cursor = m.Help.Offset
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewLogs {
		if m.Logs.Offset < len(m.Logs.Entries)-m.Display.ViewportHeight {
			m.Logs.Offset++
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewClipboard {
		if m.Operations.Clipboard.Offset < len(m.Operations.Clipboard.Paths)-m.Display.ViewportHeight {
			m.Operations.Clipboard.Offset++
		}
		return nil
	}
	if m.UI.ActiveView == context.ViewAnalyze {
		items := getAnalyzeItems(m, m.Analyze.ActiveNode)
		if m.Analyze.Offset < len(items)-m.Display.ViewportHeight {
			m.Analyze.Offset++
		}
		return nil
	}
	if m.Inputs.Mode == context.InputFuzzySearch || len(m.Navigation.Search.Results) > 0 {
		// Calculate total lines in search results
		totalLines := 0
		for _, res := range m.Navigation.Search.Results {
			totalLines++ // Header
			if !res.Collapsed {
				totalLines += len(res.Matches)
			}
			totalLines++ // Empty line after file
		}
		if m.Navigation.Search.Offset < totalLines-m.Display.ViewportHeight {
			m.Navigation.Search.Offset++
		}
		return nil
	}

	if m.Navigation.Offset < len(m.Navigation.FilteredItems)-m.Display.ViewportHeight {
		m.Navigation.Offset++
	}
	return nil
}

func handleMousePress(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	m.Display.Mouse.IsDragging = true
	m.Display.Mouse.DragStart.X = msg.X
	m.Display.Mouse.DragStart.Y = msg.Y
	m.Display.Mouse.DragEnd.X = msg.X
	m.Display.Mouse.DragEnd.Y = msg.Y
	m.Display.Mouse.DragStartIdx = -1

	// Store initial selection state for dynamic selection
	m.Display.Mouse.InitialSelection = make(map[string]bool)
	for k, v := range m.Navigation.SelectedPaths {
		m.Display.Mouse.InitialSelection[k] = v
	}

	// App Header is 1 line (y=0)
	if msg.Y == 0 {
		return handleMouseClick(m, msg) // Existing header logic
	}

	// Footer is last line
	if msg.Y == m.Display.Height-1 {
		return handleFooterClick(m, msg)
	}

	// Calculate item index
	bodyY := msg.Y - 1
	headerHeight := 0
	if m.Config.UI.ShowHeader {
		headerHeight = 3
	}

	if bodyY < headerHeight {
		return nil
	}

	itemIdx := bodyY - headerHeight + m.Navigation.Offset
	if itemIdx >= 0 && itemIdx < len(m.Navigation.FilteredItems) {
		m.Display.Mouse.DragStartIdx = itemIdx
		item := m.Navigation.FilteredItems[itemIdx]

		// Shift+Click for range selection or toggle
		if msg.Shift {
			if m.Navigation.IsSelected(item.Path) {
				m.Navigation.Deselect(item.Path)
			} else {
				if m.Navigation.Cursor != itemIdx {
					// Range select from current cursor to clicked item
					return nav.HandleShiftSelect(m, itemIdx-m.Navigation.Cursor)
				}
				m.Navigation.Select(item.Path)
			}
			m.Navigation.SelectMode = m.Navigation.SelectedCount() > 0
			return nil
		}

		// Handle selection marker click
		if m.Navigation.SelectMode && msg.X <= 4 {
			if !item.State.IsUp {
				m.Navigation.ToggleSelection(item.Path)
				m.Navigation.SelectMode = m.Navigation.SelectedCount() > 0
				return nil
			}
		}
	}

	return handleMouseClick(m, msg)
}

func handleMouseDrag(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	if !m.Display.Mouse.IsDragging {
		return nil
	}
	m.Display.Mouse.DragEnd.X = msg.X
	m.Display.Mouse.DragEnd.Y = msg.Y

	// If we started on an empty area OR on an item but NOT moving it
	// (for now, let's treat all drag as selection if startIdx was -1)
	if m.Display.Mouse.DragStartIdx == -1 {
		updateDragSelection(m)
	}

	return nil
}

func handleMouseRelease(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	if !m.Display.Mouse.IsDragging {
		return nil
	}

	startIdx := m.Display.Mouse.DragStartIdx
	m.Display.Mouse.IsDragging = false
	m.Display.Mouse.InitialSelection = nil // Clear

	// Check for drag-to-move
	if startIdx != -1 {
		// Calculate target index
		bodyY := msg.Y - 1
		headerHeight := 0
		if m.Config.UI.ShowHeader {
			headerHeight = 3
		}
		targetIdx := bodyY - headerHeight + m.Navigation.Offset

		if targetIdx >= 0 && targetIdx < len(m.Navigation.FilteredItems) && targetIdx != startIdx {
			targetItem := m.Navigation.FilteredItems[targetIdx]
			if targetItem.IsDir && !targetItem.State.IsUp {
				// Dragged onto a directory -> Move
				sourceItem := m.Navigation.FilteredItems[startIdx]
				var sources []string
				if m.Navigation.IsSelected(sourceItem.Path) {
					// Move all selected items
					for path := range m.Navigation.SelectedPaths {
						sources = append(sources, path)
					}
				} else {
					// Just move the dragged item
					sources = []string{sourceItem.Path}
				}

				if len(sources) > 0 {
					destDir := targetItem.Path
					m.Operations.Clipboard.SetCut(m.FS, sources)
					return func() tea.Msg {
						return messages.PerformPasteMsg{
							OpName:  "Move",
							Message: fmt.Sprintf("Moving %d items to %s", len(sources), m.FS.Base(destDir)),
							Paths:   sources,
							DestDir: destDir,
							IsCut:   true,
						}
					}
				}
			}
		}
	}

	return nil
}

func updateDragSelection(m *context.Model) {
	headerHeight := 1 // Header row
	if m.Config.UI.ShowHeader {
		headerHeight += 3
	}

	startY := m.Display.Mouse.DragStart.Y
	endY := m.Display.Mouse.DragEnd.Y

	if startY > endY {
		startY, endY = endY, startY
	}

	// Map screen Y to item indices
	minIdx := startY - headerHeight + m.Navigation.Offset
	maxIdx := endY - headerHeight + m.Navigation.Offset

	// Reset to initial state before applying current drag rectangle
	m.Navigation.ClearSelection()
	if m.Display.Mouse.InitialSelection != nil {
		for path := range m.Display.Mouse.InitialSelection {
			m.Navigation.Select(path)
		}
	}

	for i := range m.Navigation.FilteredItems {
		item := &m.Navigation.FilteredItems[i]
		if item.State.IsUp {
			continue
		}
		if i >= minIdx && i <= maxIdx {
			m.Navigation.Select(item.Path)
		}
	}
	m.Navigation.SelectMode = m.Navigation.SelectedCount() > 0
}

func handleMouseClick(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	// App Header is 1 line (y=0)
	if msg.Y == 0 {
		// 1. Check for breadcrumb clicks (Left side)
		if m.UI.ActiveView == context.ViewMain {
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

	// 1. Check for footer clicks (Bottom line)
	if msg.Y == m.Display.Height-1 {
		return handleFooterClick(m, msg)
	}

	// Body starts at y=1
	if msg.Y < 1 {
		return nil
	}

	// Calculate which item was clicked
	bodyY := msg.Y - 1

	switch m.UI.ActiveView {
	case context.ViewSettings:
		return handleSettingsClick(m, bodyY)
	case context.ViewHelp:
		return handleHelpClick(m, bodyY)
	case context.ViewLogs:
		return handleLogClick(m, bodyY)
	case context.ViewClipboard:
		return handleClipboardClick(m, bodyY)
	case context.ViewAnalyze:
		return handleAnalyzeClick(m, bodyY)
	}

	if m.Inputs.Mode == context.InputFuzzySearch || len(m.Navigation.Search.Results) > 0 {
		return handleSearchClick(m, msg)
	}

	headerHeight := 0
	if m.Config.UI.ShowHeader {
		headerHeight = 3
	}

	if bodyY < headerHeight {
		return nil // Clicked on list header
	}

	itemIdx := bodyY - headerHeight + m.Navigation.Offset

	// Check if clicked on empty part below file list
	if itemIdx >= len(m.Navigation.FilteredItems) {
		// Double click on empty space -> Create
		now := time.Now()
		isDoubleClick := m.Display.Mouse.LastClickIdx == -2 && now.Sub(m.Display.Mouse.LastClickTime) < 500*time.Millisecond
		m.Display.Mouse.LastClickTime = now
		m.Display.Mouse.LastClickIdx = -2
		if isDoubleClick {
			return file.StartCreate(m)
		}
		return nil
	}

	if itemIdx < 0 {
		return nil
	}

	now := time.Now()
	isDoubleClick := itemIdx == m.Display.Mouse.LastClickIdx && now.Sub(m.Display.Mouse.LastClickTime) < 500*time.Millisecond

	m.Display.Mouse.LastClickTime = now
	m.Display.Mouse.LastClickIdx = itemIdx

	if isDoubleClick {
		// Double click -> Action (Navigate or Open)
		return nav.NavigateToSelected(m)
	}

	// Single click -> Just select it
	m.Navigation.Cursor = itemIdx
	return nil
}

func handleHelpClick(m *context.Model, bodyY int) tea.Cmd {
	clickedLine := bodyY + m.Help.Offset
	if clickedLine < 0 {
		return nil
	}

	// This mapping must match buildHelpGroups and renderHelpRows in internal/tui/components/views/help.go
	currentIndex := 0
	currentLine := 1 // Initial empty line

	// Helper to check if a line was clicked and update cursor
	check := func(count int) bool {
		currentLine++ // Header
		for i := 0; i < count; i++ {
			if currentLine == clickedLine {
				m.Help.Cursor = currentIndex + i
				m.Help.Offset = app.ScrollHelp(m)
				return true
			}
			currentLine++
		}
		currentIndex += count
		currentLine++ // Spacer
		return false
	}

	if check(7) {
		return nil
	} // Navigation
	if check(4) {
		return nil
	} // Selection
	if check(3) {
		return nil
	} // Tabs
	if check(8) {
		return nil
	} // File Ops
	if check(4) {
		return nil
	} // Search
	if check(6) {
		return nil
	} // Misc

	return nil
}

func handleLogClick(m *context.Model, bodyY int) tea.Cmd {
	// Header at Y=1 (bodyY=0)
	if bodyY == 0 {
		return nil
	}

	idx := bodyY - 1
	if idx < 0 || idx >= len(m.Logs.Entries) {
		return nil
	}

	m.Logs.Cursor = idx
	m.Logs.Offset = app.ScrollLogs(m.Logs.Cursor, m.Logs.Offset, m.Display.ViewportHeight)
	return nil
}

func handleClipboardClick(m *context.Model, bodyY int) tea.Cmd {
	itemIdx := bodyY + m.Operations.Clipboard.Offset
	if itemIdx < 0 || itemIdx >= len(m.Operations.Clipboard.Paths) {
		return nil
	}

	m.Operations.Clipboard.Cursor = itemIdx
	m.Operations.Clipboard.Offset = app.ScrollLogs(m.Operations.Clipboard.Cursor, m.Operations.Clipboard.Offset, m.Display.ViewportHeight)
	return nil
}

func handleAnalyzeClick(m *context.Model, bodyY int) tea.Cmd {
	idx := bodyY + m.Analyze.Offset
	if idx < 0 {
		return nil
	}

	items := getAnalyzeItems(m, m.Analyze.ActiveNode)
	if idx >= len(items) {
		return nil
	}

	now := time.Now()
	// Use a high offset for analyze clicks
	clickID := 0xAC1D0000 | idx
	isDoubleClick := clickID == m.Display.Mouse.LastClickIdx && now.Sub(m.Display.Mouse.LastClickTime) < 500*time.Millisecond

	m.Display.Mouse.LastClickTime = now
	m.Display.Mouse.LastClickIdx = clickID

	if isDoubleClick {
		selected := items[idx]
		if selected.IsDirectory {
			if selected.Name == ".." {
				if m.Analyze.ActiveNode.Parent != nil {
					m.Analyze.ActiveNode = m.Analyze.ActiveNode.Parent
					m.Navigation.Path = m.Analyze.ActiveNode.Path
				} else {
					return StartAnalysisAtPath(m, m.FS.Dir(m.Analyze.ActiveNode.Path))
				}
			} else {
				m.Analyze.ActiveNode = selected
				m.Navigation.Path = m.Analyze.ActiveNode.Path
			}
			m.Analyze.Cursor = 0
			m.Analyze.Offset = 0
			return nil
		}
	}

	m.Analyze.Cursor = idx
	return nil
}

func handleFooterClick(m *context.Model, msg tea.MouseMsg) tea.Cmd {
	mode := utils.DetermineFooterMode(m)

	switch mode {
	case footer_comp.ModeConfirming:
		props := msg_comp.Props{
			Confirm: msg_comp.ConfirmContext{
				ActionType:     m.Operations.ActionType,
				ClipboardCount: len(m.Operations.Clipboard.Paths),
				ClipboardPaths: m.Operations.Clipboard.Paths,
				ConflictDst:    m.Operations.Conflict.Destination,
				ConflictCount:  len(m.Operations.Conflict.PendingItems),
				LatestVersion:  m.UI.LatestVersion,
			},
		}
		prompt := msg_comp.BuildConfirmationPrompt(props)
		action := findActionInPrompt(msg.X, prompt)
		if action != "" {
			return HandleUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(action)})
		}

	case footer_comp.ModeHostConfirm:
		hostname := ""
		if m.Navigation.Remote.HostConfirmReq != nil {
			hostname = m.Navigation.Remote.HostConfirmReq.Hostname
		}
		prompt := "Add host '" + hostname + "' to known_hosts? [y] Yes | [n] No"
		action := findActionInPrompt(msg.X, prompt)
		if action != "" {
			return HandleUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(action)})
		}

	case footer_comp.ModeNormal:
		action := footer_comp.GetActionAt(msg.X, footer_comp.Props{
			Width: m.Display.Width,
			Status: footer_comp.StatusInfo{
				SortMode:      m.Display.SortMode,
				Cursor:        m.Navigation.Cursor,
				TotalItems:    len(m.Navigation.FilteredItems),
				SelectedCount: m.Navigation.SelectedCount(),
				Items:         m.Navigation.Items,
				FilteredItems: m.Navigation.FilteredItems,
			},
			Confirm: footer_comp.ConfirmContext{
				ClipboardCount: len(m.Operations.Clipboard.Paths),
			},
			Styles: m.Display.Styles,
		})

		if action != "" {
			return HandleUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(action)})
		}

	case footer_comp.ModeSearching, footer_comp.ModeRenaming, footer_comp.ModeGoto, footer_comp.ModeAuth, footer_comp.ModeFuzzySearch, footer_comp.ModeZip, footer_comp.ModeUnzip, footer_comp.ModeCreate, footer_comp.ModeConflictRename:
		promptLen := utils.GetPromptLength(m)
		// Click is after leading space (1) and prompt
		relativeX := msg.X - 1 - promptLen

		// Check for right-side tab hints
		// Re-calculate unstyled widths roughly
		// FuzzySearch has a very long right side. Others are shorter.
		rightWidth := 0
		switch mode {
		case footer_comp.ModeFuzzySearch:
			rightWidth = 45 // "[Tab] Collapse | [Alt+n/m] Files | [Alt+j/k] Matches "
		case footer_comp.ModeCreate:
			rightWidth = 15 // "[Tab] File/Folder "
		}

		if msg.X >= m.Display.Width-rightWidth {
			return HandleUpdate(m, tea.KeyMsg{Type: tea.KeyTab})
		}

		if msg.X >= 1 && msg.X < 1+promptLen {
			m.Inputs.ActiveInput.SetCursor(0)
			return nil
		}

		if relativeX >= 0 {
			m.Inputs.ActiveInput.SetCursorFromX(relativeX)
			return nil
		}

	case footer_comp.ModeSettings, footer_comp.ModeLog, footer_comp.ModeClipboard:
		// Simple check for [Esc] or [.] or [Alt+L/C]
		if msg.X > 1 && msg.X < 20 { // Typical position for [Esc/...] Back
			return HandleUpdate(m, tea.KeyMsg{Type: tea.KeyEsc})
		}
	}

	return nil
}

func findActionInPrompt(x int, prompt string) string {
	// Prompt example: "Delete selected items? [y] Yes | [n] No"
	// Indices are 0-based, and we have a leading space in footer rendering.
	targetX := x - 1
	if targetX < 0 {
		return ""
	}

	// Look for [k] pattern in prompt
	for i := 0; i < len(prompt)-2; i++ {
		if prompt[i] == '[' && prompt[i+2] == ']' {
			// Found a key indicator
			if targetX >= i && targetX <= i+2 {
				return string(prompt[i+1])
			}
		}
	}
	return ""
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
	isDoubleClick := clickID == m.Display.Mouse.LastClickIdx && now.Sub(m.Display.Mouse.LastClickTime) < 500*time.Millisecond

	m.Display.Mouse.LastClickTime = now
	m.Display.Mouse.LastClickIdx = clickID

	m.Settings.Cursor = idx
	m.Settings.Offset = app.ScrollSettings(m)

	if isDoubleClick {
		if reload, cmd := app.ToggleSetting(idx, m); reload {
			return tea.Batch(cmd, func() tea.Msg { return messages.ReloadMsg{} })
		} else {
			return cmd
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

	clickedLine := bodyY - 2 + m.Navigation.Search.Offset
	if clickedLine < 0 {
		return nil
	}

	currentLine := 0
	for fIdx := range m.Navigation.Search.Results {
		res := &m.Navigation.Search.Results[fIdx]

		// File header
		if currentLine == clickedLine {
			// Clicked on file header
			if msg.X >= 0 && msg.X <= 2 {
				// Clicked on arrow
				res.Collapsed = !res.Collapsed
				m.Navigation.Search.Offset = integration.ScrollSearch(m)
				return nil
			}
			m.Navigation.Search.CursorFile = fIdx
			m.Navigation.Search.CursorMatch = -1
			m.Navigation.Search.Offset = integration.ScrollSearch(m)
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
					isDoubleClick := clickID == m.Display.Mouse.LastClickIdx && now.Sub(m.Display.Mouse.LastClickTime) < 500*time.Millisecond

					m.Display.Mouse.LastClickTime = now
					m.Display.Mouse.LastClickIdx = clickID

					m.Navigation.Search.CursorFile = fIdx
					m.Navigation.Search.CursorMatch = mIdx
					m.Navigation.Search.Offset = integration.ScrollSearch(m)

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
