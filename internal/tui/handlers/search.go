package handlers

import (
	"context"
	"time"

	"github.com/zulfikawr/fm/internal/files/ops"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// SearchDebounceDuration is the wait time before triggering a search
const SearchDebounceDuration = 300 * time.Millisecond

// HandleSearch handles search-related messages
func HandleSearch(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case SearchMsg:
		return finalizeSearch(m, msg)
	case tea.KeyMsg:
		if m.Inputs.Mode == tui_context.InputFuzzySearch {
			return handleSearchKeys(m, msg)
		}
	}
	return nil
}

func handleSearchKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	switch key {
	case "up", "alt+k":
		moveSearchCursor(m, -1)
		m.Search.Offset = scrollSearch(m)
	case "down", "alt+j":
		moveSearchCursor(m, 1)
		m.Search.Offset = scrollSearch(m)
	case "alt+m":
		if m.Search.CursorFile > 0 {
			m.Search.CursorFile--
			m.Search.CursorMatch = -1
			m.Search.Offset = scrollSearch(m)
		}
	case "alt+n":
		if m.Search.CursorFile < len(m.Search.Results)-1 {
			m.Search.CursorFile++
			m.Search.CursorMatch = -1
			m.Search.Offset = scrollSearch(m)
		}
	case "tab":
		if len(m.Search.Results) > 0 {
			res := &m.Search.Results[m.Search.CursorFile]
			res.Collapsed = !res.Collapsed
			m.Search.Offset = scrollSearch(m)
		}
	}
	return nil
}

func moveSearchCursor(m *tui_context.Model, dir int) {
	if len(m.Search.Results) == 0 {
		return
	}

	if m.Search.CursorFile < 0 {
		m.Search.CursorFile = 0
	}
	if m.Search.CursorFile >= len(m.Search.Results) {
		m.Search.CursorFile = len(m.Search.Results) - 1
	}

	if dir > 0 {
		res := m.Search.Results[m.Search.CursorFile]
		if m.Search.CursorMatch == -1 {
			if !res.Collapsed && len(res.Matches) > 0 {
				m.Search.CursorMatch = 0
			} else if m.Search.CursorFile < len(m.Search.Results)-1 {
				m.Search.CursorFile++
				m.Search.CursorMatch = -1
			}
		} else if m.Search.CursorMatch < len(res.Matches)-1 {
			m.Search.CursorMatch++
		} else if m.Search.CursorFile < len(m.Search.Results)-1 {
			m.Search.CursorFile++
			m.Search.CursorMatch = -1
		}
	} else {
		if m.Search.CursorMatch > 0 {
			m.Search.CursorMatch--
		} else if m.Search.CursorMatch == 0 {
			m.Search.CursorMatch = -1
		} else if m.Search.CursorFile > 0 {
			m.Search.CursorFile--
			res := m.Search.Results[m.Search.CursorFile]
			if res.Collapsed || len(res.Matches) == 0 {
				m.Search.CursorMatch = -1
			} else {
				m.Search.CursorMatch = len(res.Matches) - 1
			}
		}
	}
}

func scrollSearch(m *tui_context.Model) int {
	cursorLine := calculateSearchCursorLine(m)
	viewportHeight := m.Display.ViewportHeight
	if viewportHeight <= 0 {
		viewportHeight = m.Display.Height - 2
	}

	// Adjust for fixed header (2 lines)
	effectiveHeight := viewportHeight - 2
	if effectiveHeight < 1 {
		effectiveHeight = 1
	}

	offset := m.Search.Offset
	if cursorLine < offset {
		return cursorLine
	}
	if cursorLine >= offset+effectiveHeight {
		return cursorLine - effectiveHeight + 1
	}
	return offset
}

func calculateSearchCursorLine(m *tui_context.Model) int {
	if len(m.Search.Results) == 0 {
		return 0
	}

	line := 0 // Header is fixed now
	for fIdx, res := range m.Search.Results {
		if fIdx == m.Search.CursorFile {
			if m.Search.CursorMatch == -1 || res.Collapsed {
				return line
			}
			line++ // File header line
			for mIdx := range res.Matches {
				if mIdx == m.Search.CursorMatch {
					return line
				}
				line++
			}
		} else {
			line++ // File header
			if !res.Collapsed {
				line += len(res.Matches)
			}
		}
		line++ // Empty line between files
	}
	return line
}

// TriggerSearch triggers a debounced fuzzy content search
func TriggerSearch(m *tui_context.Model, query string) tea.Cmd {
	if query == "" {
		StopSearch(m)
		return nil
	}

	m.Search.Query = query
	m.Search.IsSearching = true

	// Use a timer to debounce
	return tea.Tick(SearchDebounceDuration, func(t time.Time) tea.Msg {
		return SearchMsg{
			Query: query,
		}
	})
}

// StopSearch cancels any active search and clears results
func StopSearch(m *tui_context.Model) {
	if m.Search.CancelFunc != nil {
		m.Search.CancelFunc()
		m.Search.CancelFunc = nil
	}
	m.Search.Results = nil
	m.Search.IsSearching = false
	m.Search.Query = ""
}

func finalizeSearch(m *tui_context.Model, msg SearchMsg) tea.Cmd {
	// If the query has changed since this search was triggered, ignore it
	if msg.Query != m.Search.Query {
		return nil
	}

	if msg.Results == nil && msg.Err == nil {
		// This was the debounce tick, now trigger the actual search
		return performSearch(m, msg.Query)
	}

	m.Search.IsSearching = false
	if m.Search.CancelFunc != nil {
		m.Search.CancelFunc = nil
	}

	if msg.Err != nil {
		return LogError(m, msg.Err, "Search failed")
	}

	m.Search.Results = msg.Results
	m.Search.CursorFile = 0
	m.Search.CursorMatch = 0
	m.Search.Offset = 0
	return nil
}

func performSearch(m *tui_context.Model, query string) tea.Cmd {
	// Minimum query length to avoid huge results
	if len(query) < 1 {
		m.Search.IsSearching = false
		return nil
	}

	// Cancel previous search if running
	if m.Search.CancelFunc != nil {
		m.Search.CancelFunc()
	}

	fs := m.FS
	gs := m.GS
	path := m.Navigation.Path

	ctx, cancel := context.WithTimeout(m.Context, 30*time.Second)
	m.Search.CancelFunc = cancel

	return func() tea.Msg {
		results, err := ops.Search(ctx, fs, gs, path, query)
		return SearchMsg{
			Query:   query,
			Results: results,
			Err:     err,
		}
	}
}
