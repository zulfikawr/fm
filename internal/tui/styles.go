package tui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the application.
type Theme struct {
	Name       string
	Subtle     lipgloss.Color
	Dir        lipgloss.Color
	Exec       lipgloss.Color
	File       lipgloss.Color
	SelectedBg lipgloss.Color
	SelectedFg lipgloss.Color
	Bg         lipgloss.Color
	// Git Colors
	GitMod       lipgloss.Color
	GitStaged    lipgloss.Color
	GitUntracked lipgloss.Color
	GitConflict  lipgloss.Color
	GitGhost     lipgloss.Color
}

// Available Themes
var Themes = []Theme{
	{
		Name:         "Gruvbox",
		Subtle:       lipgloss.Color("243"),
		Dir:          lipgloss.Color("208"),
		Exec:         lipgloss.Color("142"),
		File:         lipgloss.Color("223"),
		SelectedBg:   lipgloss.Color("239"),
		SelectedFg:   lipgloss.Color("214"),
		Bg:           lipgloss.Color("235"),
		GitMod:       lipgloss.Color("214"),
		GitStaged:    lipgloss.Color("142"),
		GitUntracked: lipgloss.Color("243"),
		GitConflict:  lipgloss.Color("167"),
		GitGhost:     lipgloss.Color("237"),
	},
	{
		Name:         "Nord",
		Subtle:       lipgloss.Color("243"),
		Dir:          lipgloss.Color("81"),
		Exec:         lipgloss.Color("150"),
		File:         lipgloss.Color("254"),
		SelectedBg:   lipgloss.Color("67"),
		SelectedFg:   lipgloss.Color("255"),
		Bg:           lipgloss.Color("237"),
		GitMod:       lipgloss.Color("214"),
		GitStaged:    lipgloss.Color("150"),
		GitUntracked: lipgloss.Color("243"),
		GitConflict:  lipgloss.Color("167"),
		GitGhost:     lipgloss.Color("240"),
	},
	{
		Name:         "Dracula",
		Subtle:       lipgloss.Color("61"),
		Dir:          lipgloss.Color("212"),
		Exec:         lipgloss.Color("84"),
		File:         lipgloss.Color("255"),
		SelectedBg:   lipgloss.Color("62"),
		SelectedFg:   lipgloss.Color("255"),
		Bg:           lipgloss.Color("236"),
		GitMod:       lipgloss.Color("228"),
		GitStaged:    lipgloss.Color("84"),
		GitUntracked: lipgloss.Color("61"),
		GitConflict:  lipgloss.Color("203"),
		GitGhost:     lipgloss.Color("238"),
	},
	{
		Name:         "Monokai",
		Subtle:       lipgloss.Color("241"),
		Dir:          lipgloss.Color("197"),
		Exec:         lipgloss.Color("148"),
		File:         lipgloss.Color("231"),
		SelectedBg:   lipgloss.Color("238"),
		SelectedFg:   lipgloss.Color("148"),
		Bg:           lipgloss.Color("235"),
		GitMod:       lipgloss.Color("208"),
		GitStaged:    lipgloss.Color("148"),
		GitUntracked: lipgloss.Color("241"),
		GitConflict:  lipgloss.Color("197"),
		GitGhost:     lipgloss.Color("237"),
	},
	{
		Name:         "Solarized Dark",
		Subtle:       lipgloss.Color("241"),
		Dir:          lipgloss.Color("33"),
		Exec:         lipgloss.Color("64"),
		File:         lipgloss.Color("244"),
		SelectedBg:   lipgloss.Color("235"),
		SelectedFg:   lipgloss.Color("37"),
		Bg:           lipgloss.Color("234"),
		GitMod:       lipgloss.Color("166"),
		GitStaged:    lipgloss.Color("64"),
		GitUntracked: lipgloss.Color("241"),
		GitConflict:  lipgloss.Color("160"),
		GitGhost:     lipgloss.Color("236"),
	},
	{
		Name:         "Red",
		Subtle:       lipgloss.Color("238"),
		Dir:          lipgloss.Color("196"),
		Exec:         lipgloss.Color("124"),
		File:         lipgloss.Color("250"),
		SelectedBg:   lipgloss.Color("52"),
		SelectedFg:   lipgloss.Color("196"),
		Bg:           lipgloss.Color("233"),
		GitMod:       lipgloss.Color("208"),
		GitStaged:    lipgloss.Color("34"),
		GitUntracked: lipgloss.Color("240"),
		GitConflict:  lipgloss.Color("160"),
		GitGhost:     lipgloss.Color("235"),
	},
	{
		Name:         "Tokyo Night",
		Subtle:       lipgloss.Color("238"),
		Dir:          lipgloss.Color("117"),
		Exec:         lipgloss.Color("120"),
		File:         lipgloss.Color("253"),
		SelectedBg:   lipgloss.Color("236"),
		SelectedFg:   lipgloss.Color("117"),
		Bg:           lipgloss.Color("234"),
		GitMod:       lipgloss.Color("215"),
		GitStaged:    lipgloss.Color("120"),
		GitUntracked: lipgloss.Color("240"),
		GitConflict:  lipgloss.Color("161"),
		GitGhost:     lipgloss.Color("235"),
	},
	{
		Name:         "Rose Pine",
		Subtle:       lipgloss.Color("240"),
		Dir:          lipgloss.Color("38"),
		Exec:         lipgloss.Color("150"),
		File:         lipgloss.Color("254"),
		SelectedBg:   lipgloss.Color("236"),
		SelectedFg:   lipgloss.Color("150"),
		Bg:           lipgloss.Color("234"),
		GitMod:       lipgloss.Color("214"),
		GitStaged:    lipgloss.Color("150"),
		GitUntracked: lipgloss.Color("240"),
		GitConflict:  lipgloss.Color("167"),
		GitGhost:     lipgloss.Color("236"),
	},
	{
		Name:         "Catppuccin Mocha",
		Subtle:       lipgloss.Color("241"),
		Dir:          lipgloss.Color("111"),
		Exec:         lipgloss.Color("149"),
		File:         lipgloss.Color("253"),
		SelectedBg:   lipgloss.Color("236"),
		SelectedFg:   lipgloss.Color("111"),
		Bg:           lipgloss.Color("234"),
		GitMod:       lipgloss.Color("221"),
		GitStaged:    lipgloss.Color("149"),
		GitUntracked: lipgloss.Color("241"),
		GitConflict:  lipgloss.Color("203"),
		GitGhost:     lipgloss.Color("236"),
	},
}

// Stylesheet holds the computed styles for a theme.
type Stylesheet struct {
	Header               lipgloss.Style
	Footer               lipgloss.Style
	ListHeader           lipgloss.Style
	Separator            lipgloss.Style
	Item                 lipgloss.Style
	SelectedItem         lipgloss.Style
	SettingsItem         lipgloss.Style
	SettingsSelectedItem lipgloss.Style
	DirCol               lipgloss.Style
	ExecCol              lipgloss.Style
	FileCol              lipgloss.Style
	// Git Styles
	GitMod         lipgloss.Style
	GitStaged      lipgloss.Style
	GitUntracked   lipgloss.Style
	GitConflict    lipgloss.Style
	GitGhost       lipgloss.Style
	GitIgnored     lipgloss.Style
	DimCol         lipgloss.Style
	KeyCol         lipgloss.Style
	SettingsHeader lipgloss.Style
}

// NewStylesheet computes styles based on the provided theme.
func NewStylesheet(t Theme) Stylesheet {
	return Stylesheet{
		Header: lipgloss.NewStyle().
			Foreground(t.Dir).
			Background(t.Bg).
			Padding(0, 1).
			Bold(true),

		Footer: lipgloss.NewStyle().
			Foreground(t.Subtle).
			Background(t.Bg),

		ListHeader: lipgloss.NewStyle().
			Foreground(t.Subtle).
			Bold(true),

		Separator: lipgloss.NewStyle().
			Foreground(t.Subtle),

		Item: lipgloss.NewStyle(),

		SelectedItem: lipgloss.NewStyle().
			Foreground(t.SelectedFg).
			Background(t.SelectedBg).
			Bold(true),

		SettingsItem: lipgloss.NewStyle().PaddingLeft(2),

		SettingsSelectedItem: lipgloss.NewStyle().
			Foreground(t.SelectedFg).
			Background(t.SelectedBg).
			Bold(true).
			PaddingLeft(2),

		DirCol:  lipgloss.NewStyle().Foreground(t.Dir).Bold(true),
		ExecCol: lipgloss.NewStyle().Foreground(t.Exec),
		FileCol: lipgloss.NewStyle().Foreground(t.File),

		GitMod:       lipgloss.NewStyle().Foreground(t.GitMod),
		GitStaged:    lipgloss.NewStyle().Foreground(t.GitStaged),
		GitUntracked: lipgloss.NewStyle().Foreground(t.GitUntracked),
		GitConflict:  lipgloss.NewStyle().Foreground(t.GitConflict).Bold(true),
		GitGhost:     lipgloss.NewStyle().Foreground(t.GitGhost).Strikethrough(true),
		GitIgnored:   lipgloss.NewStyle().Foreground(t.Subtle),
		DimCol:       lipgloss.NewStyle().Foreground(t.Subtle),
		KeyCol:       lipgloss.NewStyle().Foreground(t.Dir),
		SettingsHeader: lipgloss.NewStyle().
			Foreground(t.Dir).
			Bold(true).
			Padding(0, 0, 0, 1),
	}
}
