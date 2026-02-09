package handlers

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func HandleAnalyze(m *tuictx.Model, msg tea.Msg) tea.Cmd {
	if !m.UI.AnalyzeOpen {
		return nil
	}

	// Capture Delete confirmation responses
	if m.UI.Confirming && m.Operations.ActionType == constants.ActionDelete {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "y", "Y":
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				return PerformDeleteFromAnalyze(m)
			case "n", "N", "esc":
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				return func() tea.Msg { return nil }
			}
		}
		return func() tea.Msg { return nil }
	}

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonLeft:
				// Body starts at y=1
				row := msg.Y - 1
				if row >= 0 && m.Analyze.ActiveNode != nil {
					idx := row + m.Analyze.Offset
					items := getAnalyzeItems(m, m.Analyze.ActiveNode)
					if idx >= 0 && idx < len(items) {
						m.Analyze.Cursor = idx
						saveAnalyzeState(m)
					}
				}
			}
		}
		return func() tea.Msg { return nil }

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "alt+u":
			m.UI.AnalyzeOpen = false
			return func() tea.Msg { return nil }

		case "up", "k":
			if m.Analyze.Cursor > 0 {
				m.Analyze.Cursor--
			}
			syncAnalyzeOffset(m)
			saveAnalyzeState(m)
			return func() tea.Msg { return nil }

		case "down", "j":
			items := getAnalyzeItems(m, m.Analyze.ActiveNode)
			max := len(items)
			if m.Analyze.Cursor < max-1 {
				m.Analyze.Cursor++
			}
			syncAnalyzeOffset(m)
			saveAnalyzeState(m)
			return func() tea.Msg { return nil }

		case "enter", "right", "l":
			items := getAnalyzeItems(m, m.Analyze.ActiveNode)
			if m.Analyze.Cursor < len(items) {
				selected := items[m.Analyze.Cursor]
				// ONLY allow navigation if it's a directory. Disable Open File.
				if selected.IsDirectory {
					saveAnalyzeState(m)
					if selected.Name == "↑ .." {
						if m.Analyze.ActiveNode.Parent != nil {
							m.Analyze.ActiveNode = m.Analyze.ActiveNode.Parent
							m.Navigation.Path = m.Analyze.ActiveNode.Path
							restoreAnalyzeState(m)
							return func() tea.Msg { return nil }
						}
						return StartAnalysisAtPath(m, m.FS.Dir(m.Analyze.ActiveNode.Path))
					}
					m.Analyze.ActiveNode = selected
					m.Navigation.Path = m.Analyze.ActiveNode.Path
					restoreAnalyzeState(m)
				}
			}
			return func() tea.Msg { return nil }

		case "backspace", "left", "h":
			if m.Analyze.ActiveNode != nil {
				saveAnalyzeState(m)
				if m.Analyze.ActiveNode.Parent != nil {
					m.Analyze.ActiveNode = m.Analyze.ActiveNode.Parent
					m.Navigation.Path = m.Analyze.ActiveNode.Path
					restoreAnalyzeState(m)
					return func() tea.Msg { return nil }
				}
				parentPath := m.FS.Dir(m.Analyze.ActiveNode.Path)
				if parentPath == m.Analyze.ActiveNode.Path {
					m.UI.AnalyzeOpen = false
					return func() tea.Msg { return nil }
				}
				return StartAnalysisAtPath(m, parentPath)
			}
			m.UI.AnalyzeOpen = false
			return func() tea.Msg { return nil }

		case "d":
			items := getAnalyzeItems(m, m.Analyze.ActiveNode)
			if m.Analyze.Cursor < len(items) {
				selected := items[m.Analyze.Cursor]
				if selected.Name == "↑ .." {
					return func() tea.Msg { return nil }
				}
				if m.Config.ConfirmOperations {
					m.UI.StartConfirming()
					m.Operations.ActionType = constants.ActionDelete
					return func() tea.Msg { return nil }
				}
				return PerformDeleteFromAnalyze(m)
			}
			return func() tea.Msg { return nil }

		default:
			// Suppress all other global keys
			return func() tea.Msg { return nil }
		}

	case messages.AnalyzeFinishedMsg:
		m.UI.Loading = false
		if msg.Err != nil {
			return func() tea.Msg { return nil }
		}
		m.Analyze.Result = msg.Result
		m.Analyze.ActiveNode = msg.Result
		m.Navigation.Path = msg.Result.Path
		restoreAnalyzeState(m)
		return func() tea.Msg { return nil }
	}

	return nil
}

func getAnalyzeItems(m *tuictx.Model, node *core.AnalysisResult) []*core.AnalysisResult {
	if node == nil {
		return nil
	}
	var items []*core.AnalysisResult

	parentPath := m.FS.Dir(node.Path)
	if parentPath != node.Path {
		items = append(items, &core.AnalysisResult{
			Name:        "↑ ..",
			Path:        parentPath,
			IsDirectory: true,
			Parent:      node.Parent,
		})
	}

	items = append(items, node.Children...)
	return items
}

func StartAnalysis(m *tuictx.Model) tea.Cmd {
	return StartAnalysisAtPath(m, m.Navigation.Path)
}

func StartAnalysisAtPath(m *tuictx.Model, path string) tea.Cmd {
	m.UI.AnalyzeOpen = true
	m.UI.Loading = true
	m.Navigation.Path = path

	return func() tea.Msg {
		analyzer := files.NewAnalyzer(m.FS)
		res, err := analyzer.AnalyzeConcurrent(context.Background(), path, nil)
		return messages.AnalyzeFinishedMsg{
			Result: res,
			Err:    err,
		}
	}
}

func PerformDeleteFromAnalyze(m *tuictx.Model) tea.Cmd {
	items := getAnalyzeItems(m, m.Analyze.ActiveNode)
	if m.Analyze.Cursor >= len(items) {
		return nil
	}
	selected := items[m.Analyze.Cursor]
	path := selected.Path

	return func() tea.Msg {
		err := ops.DeleteMultiple(ops.DeleteOptions{
			OpCtx:    ops.OpContext{Context: context.Background(), FS: m.FS},
			Paths:    []string{path},
			UseTrash: m.Config.UseTrash,
		})
		if err != nil {
			return messages.StatusMsg{Message: "Failed to delete: " + err.Error(), IsError: true}
		}
		return messages.StartAnalyzeMsg{}
	}
}

func saveAnalyzeState(m *tuictx.Model) {
	if m.Analyze.ActiveNode != nil {
		m.Cache.AnalyzeCursorMemory.Put(m.Analyze.ActiveNode.Path, m.Analyze.Cursor)
		m.Cache.AnalyzeOffsetMemory.Put(m.Analyze.ActiveNode.Path, m.Analyze.Offset)
	}
}

func restoreAnalyzeState(m *tuictx.Model) {
	if m.Analyze.ActiveNode != nil {
		m.Analyze.Cursor, _ = m.Cache.AnalyzeCursorMemory.Get(m.Analyze.ActiveNode.Path)
		m.Analyze.Offset, _ = m.Cache.AnalyzeOffsetMemory.Get(m.Analyze.ActiveNode.Path)

		// Bounds check
		items := getAnalyzeItems(m, m.Analyze.ActiveNode)
		max := len(items)
		if m.Analyze.Cursor >= max && max > 0 {
			m.Analyze.Cursor = max - 1
		}
		if m.Analyze.Cursor < 0 {
			m.Analyze.Cursor = 0
		}

		syncAnalyzeOffset(m)
	}
}

func syncAnalyzeOffset(m *tuictx.Model) {
	if m.Display.ViewportHeight == 0 {
		return
	}

	cursor := m.Analyze.Cursor
	offset := m.Analyze.Offset
	height := m.Display.ViewportHeight

	if cursor < offset {
		m.Analyze.Offset = cursor
	} else if cursor >= offset+height {
		m.Analyze.Offset = cursor - height + 1
	}

	if m.Analyze.Offset < 0 {
		m.Analyze.Offset = 0
	}
}
