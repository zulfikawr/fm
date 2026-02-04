package cli

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RunConfig handles the config subcommand
func RunConfig(args *Args) error {
	if args.ConfigReset {
		return resetConfig()
	}

	if args.ConfigInit {
		return initConfig()
	}

	return showConfig()
}

func showConfig() error {
	cfg := config.Load()
	t := theme.Themes[cfg.ThemeIndex]
	styles := theme.NewStylesheet(t)

	fmt.Println(styles.DirCol.Render("Current Configuration"))
	fmt.Println(styles.MutedCol.Render(fmt.Sprintf("Path: %s", config.GetConfigPath())))
	fmt.Println()

	renderSetting := func(label string, value interface{}) {
		labelStyle := lipgloss.NewStyle().Foreground(t.Secondary).Bold(true)
		valueStyle := lipgloss.NewStyle().Foreground(t.File)
		fmt.Printf("  %s %v\n", labelStyle.Render(fmt.Sprintf("%-25s", label+":")), valueStyle.Render(fmt.Sprint(value)))
	}

	renderBool := func(label string, val bool) {
		labelStyle := lipgloss.NewStyle().Foreground(t.Secondary).Bold(true)
		var valStr string
		if val {
			valStr = lipgloss.NewStyle().Foreground(t.Success).Render("Enabled")
		} else {
			valStr = lipgloss.NewStyle().Foreground(t.Error).Render("Disabled")
		}
		fmt.Printf("  %s %s\n", labelStyle.Render(fmt.Sprintf("%-25s", label+":")), valStr)
	}

	fmt.Println(styles.SettingsHeader.Render("Appearance"))
	renderSetting("Theme", t.Name)
	renderBool("Nerd Font Icons", cfg.EnableIcons)

	fmt.Println()
	fmt.Println(styles.SettingsHeader.Render("Display Options"))
	renderBool("Show Column Headers", cfg.ShowHeader)
	renderBool("Enable Git Status", cfg.EnableGit)
	renderBool("Show File Size", cfg.ShowSize)
	renderSetting("Size Format", format.SizeFormats[cfg.SizeFormatIndex])
	renderBool("Show Date Modified", cfg.ShowDateModified)
	renderSetting("Date Format", format.DateFormats[cfg.DateFormatIndex].Name)
	renderBool("Enable Mouse Support", cfg.EnableMouse)

	fmt.Println()
	fmt.Println(styles.SettingsHeader.Render("File Operations"))
	renderBool("Show Hidden Files", cfg.ShowHidden)
	renderBool("Case-Sensitive Search", cfg.CaseSensitive)
	renderBool("Confirm Operations", cfg.ConfirmOperations)
	renderBool("Wrap Navigation", cfg.WrapNavigation)
	renderSetting("Preferred Editor", constants.Editors[cfg.EditorIndex])
	renderBool("Use Trash", cfg.UseTrash)

	fmt.Println()
	fmt.Println(styles.SettingsHeader.Render("Search & Filtering"))
	renderBool("Enable Regex Search", cfg.EnableRegexSearch)

	fmt.Println()
	return nil
}

func resetConfig() error {
	cfg := config.DefaultConfig()
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ Configuration reset to defaults."))
	return nil
}

// ConfigInitModel handles the interactive config initialization
type ConfigInitModel struct {
	cursor   int
	step     int
	config   config.Config
	quitting bool
}

func initConfig() error {
	m := ConfigInitModel{
		config: config.Load(),
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("config init failed: %w", err)
	}

	return nil
}

func (m ConfigInitModel) Init() tea.Cmd {
	return nil
}

func (m ConfigInitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			max := 1
			switch m.step {
			case 0: // Theme
				max = len(theme.Themes) - 1
			case 3: // Editor
				max = len(constants.Editors) - 1
			}
			if m.cursor < max {
				m.cursor++
			}
		case "enter", " ":
			return m.nextStep()
		}
	}
	return m, nil
}

func (m ConfigInitModel) nextStep() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0: // Theme
		m.config.ThemeIndex = m.cursor
		m.step++
		m.cursor = 0
	case 1: // Icons
		m.config.EnableIcons = m.cursor == 0
		m.step++
		m.cursor = 0
	case 2: // Mouse
		m.config.EnableMouse = m.cursor == 0
		m.step++
		m.cursor = 0
	case 3: // Editor
		m.config.EditorIndex = m.cursor
		m.step++
	case 4: // Save
		if err := m.config.Save(); err != nil {
			return m, func() tea.Msg { return err }
		}
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m ConfigInitModel) View() string {
	if m.quitting {
		if m.step == 4 {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("\n✓ Configuration saved successfully!\n")
		}
		return "\nConfiguration cancelled.\n"
	}

	t := theme.Themes[m.config.ThemeIndex]
	styles := theme.NewStylesheet(t)

	var s string
	s += styles.DirCol.Render("fm configuration wizard") + "\n\n"

	switch m.step {
	case 0:
		s += "Select a theme:\n\n"
		for i, themeItem := range theme.Themes {
			cursor := " "
			name := themeItem.Name
			if m.cursor == i {
				cursor = styles.DirCol.Render(">")
				name = styles.DirCol.Render(name)
			}
			s += fmt.Sprintf("%s %s\n", cursor, name)
		}
	case 1:
		s += "Enable Nerd Font icons? (Requires a Nerd Font installed)\n\n"
		options := []string{"Yes", "No"}
		for i, opt := range options {
			cursor := " "
			label := opt
			if m.cursor == i {
				cursor = styles.DirCol.Render(">")
				label = styles.DirCol.Render(label)
			}
			s += fmt.Sprintf("%s %s\n", cursor, label)
		}
	case 2:
		s += "Enable mouse support?\n\n"
		options := []string{"Yes", "No"}
		for i, opt := range options {
			cursor := " "
			label := opt
			if m.cursor == i {
				cursor = styles.DirCol.Render(">")
				label = styles.DirCol.Render(label)
			}
			s += fmt.Sprintf("%s %s\n", cursor, label)
		}
	case 3:
		s += "Select your preferred editor:\n\n"
		for i, ed := range constants.Editors {
			cursor := " "
			label := ed
			if m.cursor == i {
				cursor = styles.DirCol.Render(">")
				label = styles.DirCol.Render(label)
			}
			s += fmt.Sprintf("%s %s\n", cursor, label)
		}
	case 4:
		s += "Everything looks good! Press Enter to save.\n"
	}

	s += "\n" + styles.MutedCol.Render("[enter] confirm • [q/esc] cancel")
	return s
}
