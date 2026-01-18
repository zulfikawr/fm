package handlers

import (
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/ops"
	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleUpdate is the main message dispatcher (Reducer)
func HandleUpdate(m *context.Model, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Synchronize viewport height to ensure correct scrolling calculations
	m.SyncViewportHeight()

	// 0. Update UI components (like spinners) - Always run this first
	var spinnerCmd tea.Cmd
	m.Display.LoadingSpinner, spinnerCmd = m.Display.LoadingSpinner.Update(msg)
	if spinnerCmd != nil {
		cmds = append(cmds, spinnerCmd)
	}

	// 1. Priority/Global Handlers
	switch msg := msg.(type) {
	case ui.BlinkMsg:
		var cmd tea.Cmd
		m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
		return cmd

	case ui.TickMsg:
		return tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.Display.Width = msg.Width
		m.Display.Height = msg.Height
		m.SyncViewportHeight()
		return tea.Batch(cmds...)

	case WatchEventMsg:
		if m.Watcher.DebounceTimer != nil {
			m.Watcher.DebounceTimer.Stop()
		}
		m.Watcher.DebounceTimer = time.NewTimer(150 * time.Millisecond)
		return func() tea.Msg {
			<-m.Watcher.DebounceTimer.C
			return DebounceWatchMsg{}
		}

	case DebounceWatchMsg:
		m.Watcher.IsListening = false
		return Reload(m, false)

	case WatcherErrorMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds,
			SetErrMsg(m, "Watcher error: restarting..."),
			RestartWatcherAction(m),
		)
		return tea.Batch(cmds...)

	case WatcherClosedMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds,
			SetMsg(m, "Watcher closed: restarting..."),
			RestartWatcherAction(m),
		)
		return tea.Batch(cmds...)

	case RemotePollMsg:
		m.Watcher.IsListening = false
		cmds = append(cmds, Reload(m, true))
		return tea.Batch(cmds...)

	case DebounceFilterMsg:
		if msg.Generation == m.Navigation.FilterGen {
			ApplyFilter(m)
		}
		return nil

	case ClearMsg:
		m.Operations.Progress.Hide()
		m.Message.Pop()
		return tea.Batch(cmds...)

	case ErrorMsg:
		// Update original log entry with failure message
		for i := range m.Logs.Entries {
			if m.Logs.Entries[i].ID == msg.LogID {
				msgText := m.Logs.Entries[i].Message
				if strings.HasPrefix(msgText, "Pasting ") {
					msgText = "Failed to paste " + msgText[8:]
				} else if strings.HasPrefix(msgText, "Moving ") {
					msgText = "Failed to move " + msgText[7:]
				} else if strings.HasPrefix(msgText, "Deleting ") {
					msgText = "Failed to delete " + msgText[9:]
				} else {
					msgText = "Failed: " + msgText
				}
				LogUpdate(m, msg.LogID, context.StatusError, context.LogError, msgText, msg.Err.Error())
				break
			}
		}

		m.UI.Loading = false
		m.Operations.ProcessingItems = make(map[string]bool)

		cmds = append(cmds,
			LogError(m, msg.Err, "Operation failed"),
			Reload(m, false),
			tea.Tick(constants.ProgressDisplayDuration, func(time.Time) tea.Msg {
				return ClearMsg{}
			}),
		)
		return tea.Batch(cmds...)

	case tea.KeyMsg:
		// 1. Text Input Handling (Highest Priority)
		if m.UI.InputActive {
			// Special case: fuzzy search navigation
			isFuzzyNavKey := false
			if m.Inputs.Mode == context.InputFuzzySearch {
				switch msg.String() {
				case "up", "down", "tab", "alt+j", "alt+k", "alt+n", "alt+m":
					isFuzzyNavKey = true
				default:
					var cmd tea.Cmd
					m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
					if cmd != nil {
						cmds = append(cmds, cmd)
					}

					// Trigger search on change
					if msg.String() != "enter" && msg.String() != "esc" {
						query := m.Inputs.ActiveInput.Value()
						if query != m.Search.Query {
							cmds = append(cmds, TriggerSearch(m, query))
						}
					}
				}
			} else {
				switch msg.String() {
				case "tab":
					if m.Inputs.Mode == context.InputGoto {
						m.Inputs.AltMode = !m.Inputs.AltMode
						return nil
					}
					if m.Inputs.Mode == context.InputAuth {
						m.Inputs.AltMode = !m.Inputs.AltMode
						return nil
					}
				}

				var cmd tea.Cmd
				m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}

				if m.Inputs.Mode == context.InputSearch {
					if m.Navigation.FilterTimer != nil {
						m.Navigation.FilterTimer.Stop()
					}
					m.Navigation.FilterGen++
					gen := m.Navigation.FilterGen
					m.Navigation.FilterTimer = time.NewTimer(50 * time.Millisecond)
					cmds = append(cmds, func() tea.Msg {
						<-m.Navigation.FilterTimer.C
						return DebounceFilterMsg{Generation: gen}
					})
				}
			}

			// Handle Enter/Esc for inputs
			switch msg.String() {
			case "enter":
				if cmd := finalizeInput(m); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return tea.Batch(cmds...)
			case "esc":
				mode := m.Inputs.Mode
				m.StopInput(true)
				if mode == context.InputSearch {
					m.Navigation.FilterQuery = ""
					ApplyFilter(m)
				}
				if mode == context.InputFuzzySearch {
					StopSearch(m)
				}
				return tea.Batch(cmds...)
			}

			if !isFuzzyNavKey {
				return tea.Batch(cmds...)
			}
		}

		// Handle global keys (Quit, etc)
		switch msg.String() {
		case "ctrl+c":
			if m.Message.Text == "Press [Ctrl+C] again to close" {
				if m.FS.IsLocal() && m.Watcher.Watcher != nil {
					m.Watcher.Watcher.Close()
				}
				return tea.Quit
			}
			return SetMsg(m, "Press [Ctrl+C] again to close")
		case "alt+l":
			m.UI.ToggleLogs()
			return tea.Batch(cmds...)
		case "alt+c":
			m.UI.ToggleClipboard()
			return tea.Batch(cmds...)
		case ".":
			m.UI.ToggleSettings()
			return tea.Batch(cmds...)
		case "esc":
			// 1. High Priority: Cancel active confirmation or prompt
			if m.UI.Confirming {
				m.UI.StopConfirming()
				m.Operations.ActionType = constants.ActionNone
				return tea.Batch(cmds...)
			}
			if m.UI.HostConfirm {
				m.UI.HostConfirm = false
				m.Remote.HostConfirmReq = nil
				return tea.Batch(cmds...)
			}

			// 2. Handle closing modals next
			if m.UI.SettingsOpen {
				m.UI.ToggleSettings()
				return tea.Batch(cmds...)
			}
			if m.UI.LogOpen {
				m.UI.ToggleLogs()
				return tea.Batch(cmds...)
			}
			if m.UI.ClipboardOpen {
				m.UI.ToggleClipboard()
				return tea.Batch(cmds...)
			}

			// 3. Global esc handling for clearing selection or filter
			if !m.UI.InputActive && !m.UI.Confirming &&
				!m.UI.RemoteAuth && !m.UI.HostConfirm {
				if m.Navigation.FilterQuery != "" {
					m.Navigation.FilterQuery = ""
					ApplyFilter(m)
					return tea.Batch(cmds...)
				}
				m.ClearSelection()
				return tea.Batch(cmds...)
			}
		}
	}

	// 2. Delegate to domain handlers based on message type or UI state
	if cmd := HandleNavigation(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := HandleFileOps(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := HandleGit(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := HandleRemote(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := HandleSearch(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := HandleSettings(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := HandleLogs(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := HandleClipboard(m, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

var openFileAtLineAction = openFileAtLine

func finalizeInput(m *context.Model) tea.Cmd {
	val := m.Inputs.ActiveInput.Value()
	mode := m.Inputs.Mode

	switch mode {
	case context.InputSearch:
		m.StopInput(false)
		return nil
	case context.InputRename:
		m.StopInput(true)
		return PerformRename(m, val)
	case context.InputZip:
		m.StopInput(true)
		return PerformZip(m, val)
	case context.InputUnzip:
		m.StopInput(true)
		return PerformUnzip(m, val)
	case context.InputGoto:
		// Implement HandleGoto logic from tui_old
		m.StopInput(true)
		return handleGotoFinalize(m, val)
	case context.InputAuth:
		m.StopInput(true)
		return handleAuthFinalize(m, val)
	case context.InputFuzzySearch:
		if len(m.Search.Results) > 0 {
			res := m.Search.Results[m.Search.CursorFile]
			line := 1
			if m.Search.CursorMatch >= 0 && m.Search.CursorMatch < len(res.Matches) {
				line = res.Matches[m.Search.CursorMatch].Line
			}

			m.StopInput(true)
			StopSearch(m)

			return openFileAtLineAction(m, res.Path, line)
		}
		m.StopInput(true)
		StopSearch(m)
	}
	return nil
}

func handleGotoFinalize(m *context.Model, input string) tea.Cmd {
	// If we are currently on a remote filesystem
	if !m.FS.IsLocal() {
		if m.Inputs.AltMode { // AltMode true means Local mode when on Remote FS
			// Switch back to local
			return SwitchToLocal(m, input)
		}

		isPath := strings.HasPrefix(input, "/") || strings.HasPrefix(input, ".") || strings.HasPrefix(input, "~") || input == ""
		isConnection := strings.Contains(input, "@")

		if isPath && !isConnection {
			return NavigateToPath(m, input)
		}

		return handleRemoteGoto(m, input)
	}

	// Currently on local filesystem
	isRemote := m.Inputs.AltMode
	if !isRemote {
		// Auto-detect remote connection string
		isRemote = strings.Contains(input, "@") || (!strings.HasPrefix(input, "/") && !strings.HasPrefix(input, "./") && !strings.HasPrefix(input, "../") && !strings.HasPrefix(input, "~") && strings.Contains(input, "."))
	}

	if isRemote {
		return handleRemoteGoto(m, input)
	}

	return NavigateToPath(m, input)
}

func handleRemoteGoto(m *context.Model, input string) tea.Cmd {
	host := input
	user := ""
	keyPath := ""

	// 1. Resolve alias from ~/.ssh/config
	sshConfigs, _ := ssh.ParseSSHConfig()
	if cfg, ok := sshConfigs[input]; ok {
		host = cfg.HostName
		if host == "" {
			host = input
		}
		user = cfg.User
		keyPath = cfg.IdentityFile
	} else if strings.Contains(input, "@") {
		// 2. Parse user@host
		parts := strings.SplitN(input, "@", 2)
		user = parts[0]
		host = parts[1]
	}

	m.Remote.Host = host
	m.Remote.User = user
	m.UI.Loading = true
	m.UI.RemoteAuth = false
	m.Inputs.AltMode = false // Default to password mode for auth prompt

	return tea.Batch(
		connectRemote(host, user, "", keyPath, m.Remote.HostConfirmChan),
		listenForHostConfirmation(m.Remote.HostConfirmChan),
	)
}

func handleAuthFinalize(m *context.Model, input string) tea.Cmd {
	m.UI.Loading = true
	password := ""
	keyPath := ""

	if m.Inputs.AltMode {
		keyPath = input
	} else {
		password = input
	}

	return tea.Batch(
		connectRemote(m.Remote.Host, m.Remote.User, password, keyPath, m.Remote.HostConfirmChan),
		listenForHostConfirmation(m.Remote.HostConfirmChan),
	)
}

func openFileAtLine(m *context.Model, path string, line int) tea.Cmd {
	execCmd, isTerminal, err := ops.GetOpenAtLineCmd(m.FS, path, m.Config.EditorIndex, line)
	if err != nil {
		return SetErrMsg(m, "Error: "+err.Error())
	}

	if isTerminal {
		return tea.ExecProcess(execCmd, func(err error) tea.Msg {
			if err != nil {
				return ErrorMsg{Err: err}
			}
			return nil
		})
	} else {
		if err := execCmd.Start(); err != nil {
			return SetErrMsg(m, "Error: "+err.Error())
		}
		return nil
	}
}
