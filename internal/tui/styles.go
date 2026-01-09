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
		Name:         "Default",
		Subtle:       lipgloss.Color("241"),
		Dir:          lipgloss.Color("39"),
		Exec:         lipgloss.Color("42"),
		File:         lipgloss.Color("252"),
		SelectedBg:   lipgloss.Color("57"),
		SelectedFg:   lipgloss.Color("229"),
		Bg:           lipgloss.Color("235"),
		GitMod:       lipgloss.Color("214"), // Orange
		GitStaged:    lipgloss.Color("42"),  // Green
		GitUntracked: lipgloss.Color("241"), // Grey
		GitConflict:  lipgloss.Color("196"), // Red
		GitGhost:     lipgloss.Color("239"), // Dark Grey
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
}

// Stylesheet holds the computed styles for a theme.
type Stylesheet struct {
	Header           lipgloss.Style
	Footer           lipgloss.Style
	Item             lipgloss.Style
	SelectedItem     lipgloss.Style
	DirCol           lipgloss.Style
	ExecCol          lipgloss.Style
	FileCol          lipgloss.Style
	Modal            lipgloss.Style
	ModalTitle       lipgloss.Style
	ModalSelected    lipgloss.Style
	ModalUnselected  lipgloss.Style
	DimmedBackground lipgloss.Style
	// Git Styles
	GitMod       lipgloss.Style
	GitStaged    lipgloss.Style
	GitUntracked lipgloss.Style
	GitConflict  lipgloss.Style
	GitGhost     lipgloss.Style
}

// NewStylesheet computes styles based on the provided theme.
func NewStylesheet(t Theme) Stylesheet {
	return Stylesheet{
		Header: lipgloss.NewStyle().
			Foreground(t.File).
			Background(t.Subtle).
			Padding(0, 1).
			Bold(true),

		Footer: lipgloss.NewStyle().
			Foreground(t.File).
			Background(t.Bg),

		Item: lipgloss.NewStyle().PaddingLeft(2),

		SelectedItem: lipgloss.NewStyle().
			Foreground(t.SelectedFg).
			Background(t.SelectedBg).
			PaddingLeft(2).
			Bold(true),

		DirCol:  lipgloss.NewStyle().Foreground(t.Dir).Bold(true),
		ExecCol: lipgloss.NewStyle().Foreground(t.Exec),
		FileCol: lipgloss.NewStyle().Foreground(t.File),

		GitMod:       lipgloss.NewStyle().Foreground(t.GitMod),
		GitStaged:    lipgloss.NewStyle().Foreground(t.GitStaged),
		GitUntracked: lipgloss.NewStyle().Foreground(t.GitUntracked),
		GitConflict:  lipgloss.NewStyle().Foreground(t.GitConflict).Bold(true),
		GitGhost:     lipgloss.NewStyle().Foreground(t.GitGhost).Strikethrough(true),

		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Subtle).
			Padding(1, 2).
			Background(t.Bg),

		ModalTitle: lipgloss.NewStyle().
			Foreground(t.Dir).
			Bold(true).
			MarginBottom(1),

		ModalSelected: lipgloss.NewStyle().
			Foreground(t.SelectedFg).
			Background(t.SelectedBg).
			Padding(0, 1),

		ModalUnselected: lipgloss.NewStyle().
			Foreground(t.File).
			Padding(0, 1),

		DimmedBackground: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}
