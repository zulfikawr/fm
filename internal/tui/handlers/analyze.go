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
				return nil
			}
		}
		return nil
	}

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonLeft:
				// Only handle click if not a wheel event (handled by global)
				if msg.Action == tea.MouseActionPress {
					// Header is removed, so row starts at msg.Y
					row := msg.Y
					if row >= 0 && m.Analyze.ActiveNode != nil {
						idx := row + m.Analyze.Offset
						items := getAnalyzeItems(m, m.Analyze.ActiveNode)
						if idx >= 0 && idx < len(items) {
							m.Analyze.Cursor = idx
						}
					}
				}
			}
		}
		return nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.UI.AnalyzeOpen = false
			return nil

		case "up", "k":
			if m.Analyze.Cursor > 0 {
				m.Analyze.Cursor--
			}
			if m.Analyze.Cursor < m.Analyze.Offset {
				m.Analyze.Offset = m.Analyze.Cursor
			}
			return nil

		case "down", "j":
			items := getAnalyzeItems(m, m.Analyze.ActiveNode)
			max := len(items)
			if m.Analyze.Cursor < max-1 {
				m.Analyze.Cursor++
			}
			if m.Analyze.Cursor >= m.Analyze.Offset+m.Display.ViewportHeight {
				m.Analyze.Offset = m.Analyze.Cursor - m.Display.ViewportHeight + 1
			}
			return nil

		case "enter", "right", "l":
			items := getAnalyzeItems(m, m.Analyze.ActiveNode)
			if m.Analyze.Cursor < len(items) {
				selected := items[m.Analyze.Cursor]
				if selected.IsDirectory {
					if selected.Name == "↑ .." {
						// Going up via "↑ .."
						if m.Analyze.ActiveNode.Parent != nil {
							m.Analyze.ActiveNode = m.Analyze.ActiveNode.Parent
							m.Analyze.Cursor = 0
							m.Analyze.Offset = 0
							return nil
						}
						// If at root of analysis, re-analyze parent
						return StartAnalysisAtPath(m, m.FS.Dir(m.Analyze.ActiveNode.Path))
					}
					// Drilling down
					m.Analyze.ActiveNode = selected
					m.Analyze.Cursor = 0
					m.Analyze.Offset = 0
					return nil
				}
			}
			return nil

		case "backspace", "left", "h":
			if m.Analyze.ActiveNode != nil {
				if m.Analyze.ActiveNode.Parent != nil {
					m.Analyze.ActiveNode = m.Analyze.ActiveNode.Parent
					m.Analyze.Cursor = 0
					m.Analyze.Offset = 0
					return nil
				}
				// At root of current analysis, go up and re-analyze
				parentPath := m.FS.Dir(m.Analyze.ActiveNode.Path)
				// If we are already at filesystem root, exit
				if parentPath == m.Analyze.ActiveNode.Path {
					m.UI.AnalyzeOpen = false
					return nil
				}
				return StartAnalysisAtPath(m, parentPath)
			}
			m.UI.AnalyzeOpen = false
			return nil

		case "d":
			items := getAnalyzeItems(m, m.Analyze.ActiveNode)
			if m.Analyze.Cursor < len(items) {
				selected := items[m.Analyze.Cursor]
				if selected.Name == "↑ .." {
					return nil
				}
				if m.Config.ConfirmOperations {
					m.UI.StartConfirming()
					m.Operations.ActionType = constants.ActionDelete
					return nil
				}
				return PerformDeleteFromAnalyze(m)
			}
			return nil

		default:
			// Suppress global keys by returning nil and not bubbling up
			return nil
		}

	case messages.AnalyzeFinishedMsg:
		m.UI.Loading = false
		if msg.Err != nil {
			return nil
		}
		m.Analyze.Result = msg.Result
		m.Analyze.ActiveNode = msg.Result
		m.Analyze.Cursor = 0
		m.Analyze.Offset = 0
		return nil
	}

	return nil
}

func getAnalyzeItems(m *tuictx.Model, node *core.AnalysisResult) []*core.AnalysisResult {
	if node == nil {
		return nil
	}
	var items []*core.AnalysisResult
	
	// Check if we should show "↑ .." (not at filesystem root)
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
	// Update navigation path so if we exit, we are in the new dir
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
			OpCtx: ops.OpContext{Context: context.Background(), FS: m.FS},
			Paths: []string{path},
		})
		if err != nil {
			return messages.StatusMsg{Message: "Failed to delete: " + err.Error(), IsError: true}
		}
		return messages.StartAnalyzeMsg{}
	}
}
