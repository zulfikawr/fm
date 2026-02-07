package handlers

import (
	"strings"

	"github.com/zulfikawr/fm/internal/files/local"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/file"
	"github.com/zulfikawr/fm/internal/tui/handlers/integration"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// handleInputs handles routing and finalization of text/fuzzy inputs
func handleInputs(m *tuictx.Model, msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case messages.CompletionsMsg:
		if len(msg.Completions) > 0 {
			m.Inputs.ActiveInput.Suggestion = msg.Completions[0]
		} else {
			m.Inputs.ActiveInput.Suggestion = ""
		}
		return nil, true
	}

	if !m.UI.InputActive {
		return nil, false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Special case: navigation during search/filter
		isNavKey := false
		switch m.Inputs.Mode {
		case tuictx.InputFuzzySearch:
			switch msg.String() {
			case "up", "down", "tab", "alt+j", "alt+k", "alt+n", "alt+m":
				isNavKey = true
			}
		case tuictx.InputSearch:
			switch msg.String() {
			case "up", "down":
				isNavKey = true
			}
		}

		if isNavKey {
			if m.Inputs.Mode == tuictx.InputFuzzySearch {
				return integration.HandleSearch(m, msg), true
			}
			if m.Inputs.Mode == tuictx.InputSearch {
				if msg.String() == "up" {
					nav.MoveCursor(m, -1)
				} else {
					nav.MoveCursor(m, 1)
				}
				utils.UpdateSearchSuggestion(m)
				return nil, true
			}
		}

		if m.Inputs.Mode == tuictx.InputKeybinding {
			switch msg.String() {
			case "enter":
				if cmd := finalizeInput(m); cmd != nil {
					return cmd, true
				}
				return nil, true
			case "esc":
				m.StopInput(true)
				return nil, true
			case "backspace":
				m.Inputs.ActiveInput.SetValue("")
				return nil, true
			default:
				keyStr := msg.String()

				// Handle Shift properly for the keybinding recorder
				if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] >= 'A' && msg.Runes[0] <= 'Z' {
					// It's a shifted key, let's make it explicitly "shift+key"
					// for consistency in the configuration.
					keyStr = "shift+" + strings.ToLower(string(msg.Runes[0]))
				}

				if keyStr == " " {
					keyStr = "space"
				}

				currentValue := m.Inputs.ActiveInput.Value()
				if currentValue == "" {
					m.Inputs.ActiveInput.SetValue(keyStr)
				} else {
					keys := strings.Split(currentValue, ", ")
					exists := false
					for i, k := range keys {
						if k == keyStr {
							keys = append(keys[:i], keys[i+1:]...)
							exists = true
							break
						}
					}
					if !exists {
						keys = append(keys, keyStr)
					}
					m.Inputs.ActiveInput.SetValue(strings.Join(keys, ", "))
				}
				return nil, true
			}
		}

		if m.Inputs.Mode != tuictx.InputFuzzySearch {
			switch msg.String() {
			case "tab":
				if m.Inputs.ActiveInput.Suggestion != "" {
					m.Inputs.ActiveInput.SetValue(m.Inputs.ActiveInput.Suggestion)
					m.Inputs.ActiveInput.Suggestion = ""
					if m.Inputs.Mode == tuictx.InputSearch {
						nav.ApplyFilter(m)
					} else if m.Inputs.Mode == tuictx.InputGoto || (m.Inputs.Mode == tuictx.InputAuth && m.Inputs.AltMode) {
						fs := m.FS
						if m.Inputs.Mode == tuictx.InputAuth && m.Inputs.AltMode {
							fs = local.NewLocalFS()
						}
						return utils.FetchCompletions(m.Context, fs, m.Navigation.Path, m.Inputs.ActiveInput.Value()), true
					}
					return nil, true
				}

				if m.Inputs.Mode == tuictx.InputCreate {
					m.Inputs.AltMode = !m.Inputs.AltMode
					return nil, true
				}
			}
		}

		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		if m.Inputs.Mode == tuictx.InputFuzzySearch {
			// Trigger search on change
			if msg.String() != "enter" && msg.String() != "esc" {
				query := m.Inputs.ActiveInput.Value()
				if query != m.Search.Query {
					cmds = append(cmds, integration.TriggerSearch(m, query))
				}
			}
		}

		if m.Inputs.Mode == tuictx.InputSearch {
			cmds = append(cmds, nav.TriggerFilter(m))
			utils.UpdateSearchSuggestion(m)
		}

		if m.Inputs.Mode == tuictx.InputGoto || (m.Inputs.Mode == tuictx.InputAuth && m.Inputs.AltMode) {
			if msg.String() != "enter" && msg.String() != "esc" && msg.String() != "tab" {
				fs := m.FS
				if m.Inputs.Mode == tuictx.InputAuth && m.Inputs.AltMode {
					fs = local.NewLocalFS()
				}
				cmds = append(cmds, utils.FetchCompletions(m.Context, fs, m.Navigation.Path, m.Inputs.ActiveInput.Value()))
			}
		}

		// Handle Enter/Esc for inputs
		switch msg.String() {
		case "enter":
			if cmd := finalizeInput(m); cmd != nil {
				cmds = append(cmds, cmd)
			}
			return tea.Batch(cmds...), true
		case "esc":
			mode := m.Inputs.Mode
			m.StopInput(true)
			if mode == tuictx.InputSearch {
				m.Navigation.FilterQuery = ""
				nav.ApplyFilter(m)
			}
			if mode == tuictx.InputFuzzySearch {
				integration.StopSearch(m)
			}
			return nil, true
		}

		return tea.Batch(cmds...), true
	}

	return nil, false
}

func finalizeInput(m *tuictx.Model) tea.Cmd {
	val := m.Inputs.ActiveInput.Value()
	mode := m.Inputs.Mode

	switch mode {
	case tuictx.InputSearch:
		m.StopInput(false)
		return nil
	case tuictx.InputKeybinding:
		return app.FinalizeKeybinding(m)
	case tuictx.InputRename:
		m.StopInput(true)
		return file.PerformRename(m, val)
	case tuictx.InputConflictRename:
		m.StopInput(true)
		return file.PerformConflictRename(m, val)
	case tuictx.InputCreate:
		m.StopInput(true)
		return file.PerformCreate(m, val)
	case tuictx.InputZip:
		m.StopInput(true)
		return file.PerformZip(m, val)
	case tuictx.InputUnzip:
		m.StopInput(true)
		return file.PerformUnzip(m, val)
	case tuictx.InputGoto:
		m.StopInput(true)
		return nav.HandleGotoFinalize(m, val)
	case tuictx.InputAuth:
		m.StopInput(true)
		return integration.HandleAuthFinalize(m, val)
	case tuictx.InputFuzzySearch:
		if len(m.Search.Results) > 0 {
			res := m.Search.Results[m.Search.CursorFile]
			line := 1
			if m.Search.CursorMatch >= 0 && m.Search.CursorMatch < len(res.Matches) {
				line = res.Matches[m.Search.CursorMatch].Line
			}

			m.StopInput(true)
			integration.StopSearch(m)

			return OpenFileAtLineAction(m, res.Path, line)
		}
		m.StopInput(true)
		integration.StopSearch(m)
	}
	return nil
}
