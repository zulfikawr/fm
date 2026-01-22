package file

import (
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func OpenFile(m *tui_context.Model, selected core.Item) tea.Cmd {
	if !m.FS.IsLocal() {
		return utils.SetErrMsg(m, "Opening remote files not supported yet")
	}

	execCmd, isTerminal, err := ops.GetOpenCmd(ops.OpenOptions{
		FS:        m.FS,
		Path:      selected.Path,
		EditorIdx: m.Config.EditorIndex,
	})
	if err != nil {
		return utils.LogError(m, err, "Open")
	}

	if isTerminal {
		return tea.ExecProcess(execCmd, func(err error) tea.Msg {
			if err != nil {
				return messages.ErrorMsg{Err: err}
			}
			return nil
		})
	} else {
		if err := execCmd.Start(); err != nil {
			return utils.LogError(m, err, "Open")
		}
		return nil
	}
}

func OpenFileAtLine(m *tui_context.Model, path string, line int) tea.Cmd {
	execCmd, isTerminal, err := ops.GetOpenAtLineCmd(ops.OpenOptions{
		FS:        m.FS,
		Path:      path,
		EditorIdx: m.Config.EditorIndex,
		Line:      line,
	})
	if err != nil {
		return utils.SetErrMsg(m, "Error: "+err.Error())
	}

	if isTerminal {
		return tea.ExecProcess(execCmd, func(err error) tea.Msg {
			if err != nil {
				return messages.ErrorMsg{Err: err}
			}
			return nil
		})
	} else {
		if err := execCmd.Start(); err != nil {
			return utils.SetErrMsg(m, "Error: "+err.Error())
		}
		return nil
	}
}
