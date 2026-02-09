package app

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/logger"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"
	"github.com/zulfikawr/fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSettingsKeys handles settings-related messages
func HandleSettings(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.SettingsOpen {
			return handleSettingsKeys(m, msg)
		}
	}
	return nil
}

func handleSettingsKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	groups := buildSettingGroups(m)
	totalItems := 0
	for _, g := range groups {
		totalItems += len(g.Settings)
	}

	var reload bool
	var cmd tea.Cmd

	switch msg.String() {
	case "up", "k":
		if m.Settings.Cursor > 0 {
			m.Settings.Cursor--
		}
	case "down", "j":
		if m.Settings.Cursor < totalItems-1 {
			m.Settings.Cursor++
		}
	case "enter", "right", "l", " ":
		if m.UI.InputActive && m.Inputs.Mode == tui_context.InputKeybinding {
			return FinalizeKeybinding(m)
		}
		reload, cmd = ToggleSetting(m.Settings.Cursor, m)
	case "left", "h":
		reload, cmd = ToggleSettingPrev(m.Settings.Cursor, m)
	case "r":
		m.Operations.ActionType = constants.ActionResetSettings
		m.UI.StartConfirming()
	case "esc", "q":
		if m.UI.InputActive && m.Inputs.Mode == tui_context.InputKeybinding {
			m.StopInput(true)
			return nil
		}
		m.UI.ToggleSettings()
	}

	m.Settings.Offset = ScrollSettings(m)

	if reload {
		return tea.Batch(cmd, func() tea.Msg { return messages.ReloadMsg{} })
	}
	return cmd
}

func buildSettingGroups(m *tui_context.Model) []struct {
	Title    string
	Settings []struct{ Label string }
} {
	// Simple version just for counting items, actual rendering is in views/settings.go
	groups := []struct {
		Title    string
		Settings []struct{ Label string }
	}{
		{Title: "File Operations", Settings: make([]struct{ Label string }, 6)},
		{Title: "Display Options", Settings: make([]struct{ Label string }, 7)},
		{Title: "Search, Filtering & Inputs", Settings: make([]struct{ Label string }, 1)},
		{Title: "Appearance", Settings: make([]struct{ Label string }, 2)},
	}

	categories := []string{"general", "navigation", "file_ops", "selection", "search", "tabs"}
	titles := map[string]string{
		"general":    "Keybindings: General",
		"navigation": "Keybindings: Navigation",
		"file_ops":   "Keybindings: File Operations",
		"selection":  "Keybindings: Selection",
		"search":     "Keybindings: Search & Filter",
		"tabs":       "Keybindings: Tabs",
	}

	for _, cat := range categories {
		count := 0
		for _, kb := range m.Config.Keybindings {
			if kb.Category == cat {
				count++
			}
		}
		if count > 0 {
			groups = append(groups, struct {
				Title    string
				Settings []struct{ Label string }
			}{Title: titles[cat], Settings: make([]struct{ Label string }, count)})
		}
	}

	return groups
}

func GetSettingGroups(m *tui_context.Model) []struct {
	Title    string
	Settings []struct {
		Label    string
		Value    string
		Inactive bool
	}
} {
	// This should match the structure in internal/tui/components/views/settings.go
	// In a real refactor, we would unify these.
	return nil
}

// ScrollSettings recalculates the settings view offset
func ScrollSettings(m *tui_context.Model) int {
	cursor := m.Settings.Cursor
	offset := m.Settings.Offset
	height := m.Display.ViewportHeight

	// Map cursor index to actual display row index (including headers and spacing)
	groups := buildSettingGroups(m)
	rowIdx := 1 // Start with spacing above first group
	itemCount := 0

	for i, g := range groups {
		if i > 0 {
			rowIdx += 1 // Spacing between groups
		}
		rowIdx += 1 // Group title row

		groupSize := len(g.Settings)
		if cursor < itemCount+groupSize {
			// Found the group containing the cursor
			rowIdx += (cursor - itemCount)
			break
		}
		rowIdx += groupSize
		itemCount += groupSize
	}

	if rowIdx < offset {
		return rowIdx
	}

	// When moving back to the top, ensure we show the first header and its gap
	if cursor == 0 {
		return 0
	}

	if rowIdx >= offset+height {
		return rowIdx - height + 1
	}

	return offset
}

func BuildFullSettingList(m *tui_context.Model) []SettingItem {
	cfg := m.Config
	var items []SettingItem

	// Group 1: File Operations (6)
	items = append(items, SettingItem{Label: "Show Hidden Files", Action: "toggle_hidden", HelpText: "Show/hide files starting with '.'"})
	items = append(items, SettingItem{Label: "Case-Sensitive Search", Action: "toggle_case", HelpText: "Search respects capitalization"})
	items = append(items, SettingItem{Label: "Confirm Operations", Action: "toggle_confirm", HelpText: "Ask before destructive actions"})
	items = append(items, SettingItem{Label: "Wrap Navigation", Action: "toggle_wrap", HelpText: "Cursor loops at list boundaries"})
	items = append(items, SettingItem{Label: "Preferred Editor", Action: "pick_editor", HelpText: "Choose default editor for opening files"})
	items = append(items, SettingItem{Label: "Use Trash (Move to Trash)", Action: "toggle_trash", HelpText: "Move deleted items to trash (off = permanent delete)"})

	// Group 2: Display Options (8)
	items = append(items, SettingItem{Label: "Show Column Headers", Action: "toggle_header", HelpText: "Show/hide list column headers"})
	items = append(items, SettingItem{Label: "Show RAM Usage", Action: "toggle_ram", HelpText: "Display RAM usage in footer"})
	items = append(items, SettingItem{Label: "Enable Git Status", Action: "toggle_git", HelpText: "Enable git status markers"})
	items = append(items, SettingItem{Label: "Show File Size", Action: "toggle_size", HelpText: "Show file size in list"})
	items = append(items, SettingItem{Label: "Size Format", Action: "pick_size_format", HelpText: "Change the file size display unit"})
	items = append(items, SettingItem{Label: "Show Date Modified", Action: "toggle_date", HelpText: "Show last modification time"})
	items = append(items, SettingItem{Label: "Date Format", Action: "pick_date_format", HelpText: "Change the date and time format"})
	items = append(items, SettingItem{Label: "Enable Mouse Support", Action: "toggle_mouse", HelpText: "Allow mouse interaction (clicks, scroll)"})

	// Group 3: Search, Filtering & Inputs (1)
	items = append(items, SettingItem{Label: "Enable Regex Search", Action: "toggle_regex", HelpText: "Use regular expressions for searching"})

	// Group 4: Appearance (2)
	items = append(items, SettingItem{Label: "Enable Nerd Font Icons", Action: "toggle_icons", HelpText: "Toggle Nerd Font icons (requires download)"})
	items = append(items, SettingItem{Label: "Theme", Action: "pick_theme", HelpText: "Change the application color scheme"})

	// Group 5+: Keybindings
	categories := []string{"general", "navigation", "file_ops", "selection", "search", "tabs"}
	for _, cat := range categories {
		for _, kb := range cfg.Keybindings {
			if kb.Category == cat {
				items = append(items, SettingItem{
					Label:        kb.HumanLabel(),
					IsKeybinding: true,
					Action:       kb.Action,
					Keys:         kb.Keys,
				})
			}
		}
	}

	return items
}

// SettingItem represents a single setting in the list
type SettingItem struct {
	Label        string
	Action       string
	HelpText     string
	IsKeybinding bool
	Keys         []string
}

func ToggleSetting(idx int, m *tui_context.Model) (bool, tea.Cmd) {
	items := BuildFullSettingList(m)
	if idx < 0 || idx >= len(items) {
		return false, nil
	}

	item := items[idx]
	cfg := m.Config
	reload := false
	var cmd tea.Cmd

	if item.IsKeybinding {
		m.StartInput(tui_context.InputKeybinding)
		m.Inputs.ActiveInput.SetValue(strings.Join(item.Keys, ", "))

		// Find the human label
		label := item.Label

		m.Inputs.ActiveInput.SetPrompt(fmt.Sprintf("Bind %s: ", label))
		return false, m.Inputs.ActiveInput.FocusCmd()
	}

	switch item.Action {
	case "toggle_hidden":
		cfg.ShowHidden = !cfg.ShowHidden
		reload = true
	case "toggle_case":
		cfg.CaseSensitive = !cfg.CaseSensitive
		reload = true
	case "toggle_confirm":
		cfg.ConfirmOperations = !cfg.ConfirmOperations
	case "toggle_wrap":
		cfg.WrapNavigation = !cfg.WrapNavigation
	case "pick_editor":
		cfg.EditorIndex = (cfg.EditorIndex + 1) % len(constants.Editors)
	case "toggle_trash":
		cfg.UseTrash = !cfg.UseTrash
	case "toggle_header":
		cfg.ShowHeader = !cfg.ShowHeader
		m.SyncViewportHeight()
	case "toggle_git":
		cfg.EnableGit = !cfg.EnableGit
		reload = true
	case "toggle_size":
		cfg.ShowSize = !cfg.ShowSize
		m.SyncViewportHeight()
		reload = true
	case "pick_size_format":
		if cfg.ShowSize {
			cfg.SizeFormatIndex = (cfg.SizeFormatIndex + 1) % len(format.SizeFormats)
			reload = true
		}
	case "toggle_date":
		cfg.ShowDateModified = !cfg.ShowDateModified
		m.SyncViewportHeight()
		reload = true
	case "pick_date_format":
		if cfg.ShowDateModified {
			cfg.DateFormatIndex = (cfg.DateFormatIndex + 1) % len(format.DateFormats)
			reload = true
		}
	case "toggle_mouse":
		cfg.EnableMouse = !cfg.EnableMouse
	case "toggle_regex":
		cfg.EnableRegexSearch = !cfg.EnableRegexSearch
	case "toggle_ram":
		cfg.ShowRAMUsage = !cfg.ShowRAMUsage
	case "toggle_icons":
		if !cfg.EnableIcons {
			if !theme.HasIconsDownloaded() {
				m.UI.Loading = true
				cmd = func() tea.Msg {
					err := theme.DownloadIcons()
					return messages.IconsDownloadedMsg{Err: err}
				}
				return true, cmd
			}
			m.UI.TestingIcons = true
			m.Operations.ActionType = constants.ActionTestIcons
			m.UI.StartConfirming()
		} else {
			cfg.EnableIcons = false
			reload = true
		}
	case "pick_theme":
		cfg.ThemeIndex = (cfg.ThemeIndex + 1) % len(theme.Themes)
		m.Display.Styles = theme.GetStylesheet(cfg.ThemeIndex)
		m.Display.LoadingSpinner.Style = m.Display.LoadingSpinner.Style.Foreground(theme.Themes[cfg.ThemeIndex].Dir)
	}

	if err := cfg.Save(); err != nil {
		logger.Errorf("Failed to save config: %v", err)
	}
	m.Config = cfg
	return reload, cmd
}

func ToggleSettingPrev(idx int, m *tui_context.Model) (bool, tea.Cmd) {
	items := BuildFullSettingList(m)
	if idx < 0 || idx >= len(items) {
		return false, nil
	}

	item := items[idx]
	if item.IsKeybinding {
		return ToggleSetting(idx, m)
	}

	cfg := m.Config
	var reload bool
	var cmd tea.Cmd

	switch item.Action {
	case "pick_editor":
		cfg.EditorIndex = (cfg.EditorIndex - 1 + len(constants.Editors)) % len(constants.Editors)
	case "pick_size_format":
		if cfg.ShowSize {
			cfg.SizeFormatIndex = (cfg.SizeFormatIndex - 1 + len(format.SizeFormats)) % len(format.SizeFormats)
			reload = true
		}
	case "pick_date_format":
		if cfg.ShowDateModified {
			cfg.DateFormatIndex = (cfg.DateFormatIndex - 1 + len(format.DateFormats)) % len(format.DateFormats)
			reload = true
		}
	case "pick_theme":
		cfg.ThemeIndex = (cfg.ThemeIndex - 1 + len(theme.Themes)) % len(theme.Themes)
		m.Display.Styles = theme.GetStylesheet(cfg.ThemeIndex)
		m.Display.LoadingSpinner.Style = m.Display.LoadingSpinner.Style.Foreground(theme.Themes[cfg.ThemeIndex].Dir)
	default:
		return ToggleSetting(idx, m)
	}

	if err := cfg.Save(); err != nil {
		logger.Errorf("Failed to save config: %v", err)
	}
	m.Config = cfg
	return reload, cmd
}

// FinalizeKeybinding saves the new keybinding for an action
func FinalizeKeybinding(m *tui_context.Model) tea.Cmd {
	val := m.Inputs.ActiveInput.Value()
	m.StopInput(true)

	items := BuildFullSettingList(m)
	idx := m.Settings.Cursor
	if idx < 0 || idx >= len(items) || !items[idx].IsKeybinding {
		return nil
	}

	targetAction := items[idx].Action

	// Simple comma-separated keys parsing
	keys := strings.Split(val, ", ")
	for i := range keys {
		keys[i] = strings.TrimSpace(keys[i])
	}

	// Update the specific keybinding in the config
	newKeybinds := make([]config.Keybinding, len(m.Config.Keybindings))
	copy(newKeybinds, m.Config.Keybindings)

	found := false
	for i := range newKeybinds {
		if newKeybinds[i].Action == targetAction {
			newKeybinds[i].Keys = keys
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	// Validate before saving
	if err := config.ValidateKeybindings(newKeybinds); err != nil {
		return utils.SetErrMsg(m, err.Error())
	}

	m.Config.Keybindings = newKeybinds
	if err := m.Config.Save(); err != nil {
		return utils.SetErrMsg(m, "Failed to save keybindings: "+err.Error())
	}

	return utils.SetMsg(m, "Keybinding updated")
}

// ConfirmSettingsReset resets all settings to defaults
func ConfirmSettingsReset(m *tui_context.Model) tea.Cmd {
	newCfg := config.DefaultConfig()
	if err := newCfg.Save(); err != nil {
		logger.Errorf("Failed to save config: %v", err)
	}

	m.Config = newCfg
	m.Display.Styles = theme.GetStylesheet(m.Config.ThemeIndex)
	m.UI.StopConfirming()
	m.Operations.ActionType = ""

	return tea.Batch(utils.SetMsg(m, "Settings reset to defaults"), func() tea.Msg { return messages.ReloadMsg{} })
}
