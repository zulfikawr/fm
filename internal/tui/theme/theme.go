package theme

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

// GetStylesheet returns the stylesheet for the given theme index.
func GetStylesheet(index int) Stylesheet {
	if index < 0 || index >= len(Themes) {
		index = 0
	}
	return NewStylesheet(Themes[index])
}
