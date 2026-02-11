package file

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/archive"
	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/logger"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func StartZip(m *tui_context.Model) tea.Cmd {
	targets := GetTargets(m)
	if len(targets) == 0 {
		return nil
	}

	m.StartInput(tui_context.InputZip)
	m.Inputs.ActiveInput.SetValue("archive.zip")
	m.Inputs.ActiveInput.SetCursor(len("archive"))
	return m.Inputs.ActiveInput.FocusCmd()
}

func PerformZipWithTargets(m *tui_context.Model, zipName string, targets []string) tea.Cmd {
	if zipName == "" {
		return nil
	}

	if len(targets) == 0 {
		return nil
	}

	dst := m.FS.Join(m.Navigation.Path, zipName)

	if m.Operations.ConflictPolicy == conflict.Ask {
		resolver := conflict.NewResolver()
		resolvedPath, renamed, err := resolver.Resolve(m.Context, m.FS, conflict.ResolveOptions{
			Src:    targets[0],
			Dst:    dst,
			Policy: m.Operations.ConflictPolicy,
		})
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				m.UI.Loading = false
				m.Operations.Conflict.Set(tui_context.ConflictParams{
					Source:       cerr.Source,
					Destination:  cerr.Destination,
					PendingItems: targets,
					IsMove:       false,
					OpType:       "zip",
				})
				m.Operations.ActionType = constants.ActionConflict
				m.UI.StartConfirming()
				return nil
			}
		}
		if resolvedPath == "" {
			return nil // Skip
		}
		if renamed {
			logger.Debugf("Operation destination renamed due to conflict: %s", resolvedPath)
		}
		dst = resolvedPath
	}

	m.ClearSelection()
	m.UI.Loading = true

	msg := fmt.Sprintf("Zipping %d items into %s", len(targets), zipName)
	return func() tea.Msg { return messages.PerformZipMsg{Targets: targets, Dst: dst, Message: msg} }
}

func PerformZip(m *tui_context.Model, zipName string) tea.Cmd {
	if zipName == "" {
		return nil
	}

	targets := GetTargets(m)
	if len(targets) == 0 {
		return nil
	}

	dst := m.FS.Join(m.Navigation.Path, zipName)

	if m.Operations.ConflictPolicy == conflict.Ask {
		resolver := conflict.NewResolver()
		resolvedPath, renamed, err := resolver.Resolve(m.Context, m.FS, conflict.ResolveOptions{
			Src:    targets[0],
			Dst:    dst,
			Policy: m.Operations.ConflictPolicy,
		})
		if err != nil {
			// ... (omitting fields for brevity in instruction, but keeping them in new_string)
			if cerr, ok := err.(*conflict.ConflictError); ok {
				m.UI.Loading = false
				m.Operations.Conflict.Set(tui_context.ConflictParams{
					Source:       cerr.Source,
					Destination:  cerr.Destination,
					PendingItems: targets,
					IsMove:       false,
					OpType:       "zip",
				})
				m.Operations.ActionType = constants.ActionConflict
				m.UI.StartConfirming()
				return nil
			}
		}
		if resolvedPath == "" {
			return nil // Skip
		}
		if renamed {
			logger.Debugf("Operation destination renamed due to conflict: %s", resolvedPath)
		}
		dst = resolvedPath
	}

	m.ClearSelection()
	m.UI.Loading = true

	msg := fmt.Sprintf("Zipping %d items into %s", len(targets), zipName)
	return func() tea.Msg { return messages.PerformZipMsg{Targets: targets, Dst: dst, Message: msg} }
}

func StartUnzip(m *tui_context.Model) tea.Cmd {
	targets := GetTargets(m)
	var zipPath string
	if len(targets) > 0 {
		tempItem := core.Item{Name: targets[0], IsDir: false, State: core.ItemState{}}
		if tempItem.IsArchive() {
			zipPath = targets[0]
		}
	} else if len(m.Navigation.FilteredItems) > 0 {
		selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
		if !selected.State.IsUp && selected.IsArchive() {
			zipPath = selected.Path
		}
	}

	if zipPath == "" {
		return func() tea.Msg {
			return messages.ErrorMsg{Err: fmt.Errorf("please select a supported archive file to unzip")}
		}
	}

	m.StartInput(tui_context.InputUnzip)

	extractPath := archive.GetDefaultExtractionPath(m.FS, zipPath)
	folderName := m.FS.Base(extractPath)
	m.Inputs.ActiveInput.SetValue(folderName)
	return m.Inputs.ActiveInput.FocusCmd()
}

func PerformUnzipWithTargets(m *tui_context.Model, destName string, targets []string) tea.Cmd {
	if destName == "" {
		return nil
	}

	var zipPath string
	if len(targets) > 0 {
		zipPath = targets[0]
	} else {
		selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
		zipPath = selected.Path
	}

	dst := m.FS.Join(m.Navigation.Path, destName)

	if m.Operations.ConflictPolicy == conflict.Ask {
		resolver := conflict.NewResolver()
		resolvedPath, renamed, err := resolver.Resolve(m.Context, m.FS, conflict.ResolveOptions{
			Src:    zipPath,
			Dst:    dst,
			Policy: m.Operations.ConflictPolicy,
		})
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				m.UI.Loading = false
				m.Operations.Conflict.Set(tui_context.ConflictParams{
					Source:       cerr.Source,
					Destination:  cerr.Destination,
					PendingItems: []string{zipPath},
					IsMove:       false,
				})
				m.UI.StartConfirming()
				return nil
			}
		}
		if resolvedPath == "" {
			return nil // Skip
		}
		if renamed {
			logger.Debugf("Operation destination renamed due to conflict: %s", resolvedPath)
		}
		dst = resolvedPath
	}

	m.ClearSelection()
	m.UI.Loading = true

	msg := fmt.Sprintf("Extracting %s to %s", m.FS.Base(zipPath), destName)
	return func() tea.Msg { return messages.PerformUnzipMsg{ZipPath: zipPath, Dst: dst, Message: msg} }
}

func PerformUnzip(m *tui_context.Model, destName string) tea.Cmd {
	if destName == "" {
		return nil
	}

	targets := GetTargets(m)
	var zipPath string
	if len(targets) > 0 {
		zipPath = targets[0]
	} else {
		selected := m.Navigation.FilteredItems[m.Navigation.Cursor]
		zipPath = selected.Path
	}

	dst := m.FS.Join(m.Navigation.Path, destName)

	if m.Operations.ConflictPolicy == conflict.Ask {
		resolver := conflict.NewResolver()
		resolvedPath, renamed, err := resolver.Resolve(m.Context, m.FS, conflict.ResolveOptions{
			Src:    zipPath,
			Dst:    dst,
			Policy: m.Operations.ConflictPolicy,
		})
		if err != nil {
			// ... (omitting fields for brevity in instruction, but keeping them in new_string)
			if cerr, ok := err.(*conflict.ConflictError); ok {
				m.UI.Loading = false
				m.Operations.Conflict.Set(tui_context.ConflictParams{
					Source:       cerr.Source,
					Destination:  cerr.Destination,
					PendingItems: []string{zipPath},
					IsMove:       false,
					OpType:       "unzip",
				})
				m.Operations.ActionType = constants.ActionConflict
				m.UI.StartConfirming()
				return nil
			}
		}
		if resolvedPath == "" {
			return nil // Skip
		}
		if renamed {
			logger.Debugf("Operation destination renamed due to conflict: %s", resolvedPath)
		}
		dst = resolvedPath
	}

	m.ClearSelection()
	m.UI.Loading = true

	msg := fmt.Sprintf("Extracting %s into %s", m.FS.Base(zipPath), destName)
	return func() tea.Msg { return messages.PerformUnzipMsg{ZipPath: zipPath, Dst: dst, Message: msg} }
}
