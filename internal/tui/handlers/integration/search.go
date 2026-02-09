package integration

import (
	"context"
	"time"

	"github.com/zulfikawr/fm/internal/files/ops"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// SearchDebounceDuration is the wait time before triggering a search
const SearchDebounceDuration = 300 * time.Millisecond

// HandleSearch handles search-related messages
func HandleSearch(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case messages.SearchMsg:
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
		m.Navigation.Search.Offset = ScrollSearch(m)
	case "down", "alt+j":
		moveSearchCursor(m, 1)
		m.Navigation.Search.Offset = ScrollSearch(m)
	case "alt+m":
		if m.Navigation.Search.CursorFile > 0 {
			m.Navigation.Search.CursorFile--
			m.Navigation.Search.CursorMatch = -1
			m.Navigation.Search.Offset = ScrollSearch(m)
		}
	case "alt+n":
		if m.Navigation.Search.CursorFile < len(m.Navigation.Search.Results)-1 {
			m.Navigation.Search.CursorFile++
			m.Navigation.Search.CursorMatch = -1
			m.Navigation.Search.Offset = ScrollSearch(m)
		}
	case "tab":
		if len(m.Navigation.Search.Results) > 0 {
			res := &m.Navigation.Search.Results[m.Navigation.Search.CursorFile]
			res.Collapsed = !res.Collapsed
			m.Navigation.Search.Offset = ScrollSearch(m)
		}
	}
	return nil
}

func moveSearchCursor(m *tui_context.Model, dir int) {
	if len(m.Navigation.Search.Results) == 0 {
		return
	}

	if m.Navigation.Search.CursorFile < 0 {
		m.Navigation.Search.CursorFile = 0
	}
	if m.Navigation.Search.CursorFile >= len(m.Navigation.Search.Results) {
		m.Navigation.Search.CursorFile = len(m.Navigation.Search.Results) - 1
	}

	if dir > 0 {
		res := m.Navigation.Search.Results[m.Navigation.Search.CursorFile]
		if m.Navigation.Search.CursorMatch == -1 {
			if !res.Collapsed && len(res.Matches) > 0 {
				m.Navigation.Search.CursorMatch = 0
			} else if m.Navigation.Search.CursorFile < len(m.Navigation.Search.Results)-1 {
				m.Navigation.Search.CursorFile++
				m.Navigation.Search.CursorMatch = -1
			}
		} else if m.Navigation.Search.CursorMatch < len(res.Matches)-1 {
			m.Navigation.Search.CursorMatch++
		} else if m.Navigation.Search.CursorFile < len(m.Navigation.Search.Results)-1 {
			m.Navigation.Search.CursorFile++
			m.Navigation.Search.CursorMatch = -1
		}
	} else {
		if m.Navigation.Search.CursorMatch > 0 {
			m.Navigation.Search.CursorMatch--
		} else if m.Navigation.Search.CursorMatch == 0 {
			m.Navigation.Search.CursorMatch = -1
		} else if m.Navigation.Search.CursorFile > 0 {
			m.Navigation.Search.CursorFile--
			res := m.Navigation.Search.Results[m.Navigation.Search.CursorFile]
			if res.Collapsed || len(res.Matches) == 0 {
				m.Navigation.Search.CursorMatch = -1
			} else {
				m.Navigation.Search.CursorMatch = len(res.Matches) - 1
			}
		}
	}
}

func ScrollSearch(m *tui_context.Model) int {
	cursorLine := CalculateSearchCursorLine(m)
	viewportHeight := m.Display.ViewportHeight
	if viewportHeight <= 0 {
		viewportHeight = m.Display.Height - 2
	}

	effectiveHeight := viewportHeight - 2
	if effectiveHeight < 1 {
		effectiveHeight = 1
	}

	offset := m.Navigation.Search.Offset
	if cursorLine < offset {
		return cursorLine
	}
	if cursorLine >= offset+effectiveHeight {
		return cursorLine - effectiveHeight + 1
	}
	return offset
}

func CalculateSearchCursorLine(m *tui_context.Model) int {
	if len(m.Navigation.Search.Results) == 0 {
		return 0
	}

	line := 0
	for fIdx, res := range m.Navigation.Search.Results {
		if fIdx == m.Navigation.Search.CursorFile {
			if m.Navigation.Search.CursorMatch == -1 || res.Collapsed {
				return line
			}
			line++
			for mIdx := range res.Matches {
				if mIdx == m.Navigation.Search.CursorMatch {
					return line
				}
				line++
			}
		} else {
			line++
			if !res.Collapsed {
				line += len(res.Matches)
			}
		}
		line++
	}
	return line
}

func TriggerSearch(m *tui_context.Model, query string) tea.Cmd {
	if query == "" {
		StopSearch(m)
		return nil
	}

	m.Navigation.Search.Query = query
	m.Navigation.Search.IsSearching = true

	return tea.Tick(SearchDebounceDuration, func(t time.Time) tea.Msg {
		return messages.SearchMsg{
			Query: query,
		}
	})
}

func StopSearch(m *tui_context.Model) {
	if m.Navigation.Search.CancelFunc != nil {
		m.Navigation.Search.CancelFunc()
		m.Navigation.Search.CancelFunc = nil
	}
	m.Navigation.Search.Results = nil
	m.Navigation.Search.IsSearching = false
	m.Navigation.Search.Query = ""
}

func finalizeSearch(m *tui_context.Model, msg messages.SearchMsg) tea.Cmd {
	if msg.Query != m.Navigation.Search.Query {
		return nil
	}

	if msg.Results == nil && msg.Err == nil {
		return performSearch(m, msg.Query)
	}

	m.Navigation.Search.IsSearching = false
	if m.Navigation.Search.CancelFunc != nil {
		m.Navigation.Search.CancelFunc = nil
	}

	if msg.Err != nil {
		return utils.LogError(m, msg.Err, "Search failed")
	}

	m.Navigation.Search.Results = msg.Results
	m.Navigation.Search.CursorFile = 0
	m.Navigation.Search.CursorMatch = 0
	m.Navigation.Search.Offset = 0
	return nil
}

func performSearch(m *tui_context.Model, query string) tea.Cmd {
	if len(query) < 1 {
		m.Navigation.Search.IsSearching = false
		return nil
	}

	if m.Navigation.Search.CancelFunc != nil {
		m.Navigation.Search.CancelFunc()
	}

	fs := m.FS
	gs := m.GS
	path := m.Navigation.Path
	isRegex := m.Config.Ops.EnableRegexSearch

	ctx, cancel := context.WithTimeout(m.Context, 30*time.Second)
	m.Navigation.Search.CancelFunc = cancel

	return func() tea.Msg {
		results, err := ops.Search(ops.SearchOptions{
			OpCtx: ops.OpContext{Context: ctx, FS: fs},
			Git:   gs,
			Root:  path,
			Query: query,
			Regex: isRegex,
		})
		return messages.SearchMsg{
			Query:   query,
			Results: results,
			Err:     err,
		}
	}
}
