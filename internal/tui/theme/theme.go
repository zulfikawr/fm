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
	// Semantic Colors
	Primary   lipgloss.Color // Main accent (usually same as Dir)
	Secondary lipgloss.Color // Secondary accent (usually same as Exec)
	Accent    lipgloss.Color // Tertiary accent for highlights
	Muted     lipgloss.Color // Muted text (lighter than Subtle)
	Highlight lipgloss.Color // Bright highlights for badges/counts
	Info      lipgloss.Color // Informational elements
	Success   lipgloss.Color // Success states
	Warning   lipgloss.Color // Warning states
	Error     lipgloss.Color // Error states
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
		Primary:      lipgloss.Color("208"), // Orange
		Secondary:    lipgloss.Color("142"), // Green
		Accent:       lipgloss.Color("109"), // Blue-gray
		Muted:        lipgloss.Color("245"), // Lighter gray
		Highlight:    lipgloss.Color("214"), // Bright orange
		Info:         lipgloss.Color("66"),  // Blue-teal
		Success:      lipgloss.Color("142"), // Green
		Warning:      lipgloss.Color("214"), // Orange
		Error:        lipgloss.Color("167"), // Red
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
		Primary:      lipgloss.Color("81"),  // Cyan
		Secondary:    lipgloss.Color("150"), // Light green
		Accent:       lipgloss.Color("109"), // Gray-blue
		Muted:        lipgloss.Color("246"), // Lighter gray
		Highlight:    lipgloss.Color("117"), // Bright cyan
		Info:         lipgloss.Color("67"),  // Darker blue
		Success:      lipgloss.Color("150"), // Green
		Warning:      lipgloss.Color("214"), // Warm orange
		Error:        lipgloss.Color("167"), // Red
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
		Primary:      lipgloss.Color("212"), // Purple
		Secondary:    lipgloss.Color("84"),  // Cyan
		Accent:       lipgloss.Color("141"), // Pink-purple
		Muted:        lipgloss.Color("104"), // Lighter purple-gray
		Highlight:    lipgloss.Color("219"), // Bright pink
		Info:         lipgloss.Color("117"), // Blue
		Success:      lipgloss.Color("84"),  // Cyan
		Warning:      lipgloss.Color("228"), // Yellow
		Error:        lipgloss.Color("203"), // Red
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
		Primary:      lipgloss.Color("197"), // Magenta
		Secondary:    lipgloss.Color("148"), // Green
		Accent:       lipgloss.Color("81"),  // Cyan
		Muted:        lipgloss.Color("244"), // Lighter gray
		Highlight:    lipgloss.Color("213"), // Bright magenta
		Info:         lipgloss.Color("117"), // Blue
		Success:      lipgloss.Color("148"), // Green
		Warning:      lipgloss.Color("208"), // Orange
		Error:        lipgloss.Color("197"), // Red
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
		Primary:      lipgloss.Color("33"),  // Blue
		Secondary:    lipgloss.Color("64"),  // Green
		Accent:       lipgloss.Color("37"),  // Cyan
		Muted:        lipgloss.Color("244"), // Base0
		Highlight:    lipgloss.Color("136"), // Yellow
		Info:         lipgloss.Color("61"),  // Violet
		Success:      lipgloss.Color("64"),  // Green
		Warning:      lipgloss.Color("166"), // Orange
		Error:        lipgloss.Color("160"), // Red
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
		Primary:      lipgloss.Color("196"), // Bright red
		Secondary:    lipgloss.Color("124"), // Dark red
		Accent:       lipgloss.Color("203"), // Medium red-pink
		Muted:        lipgloss.Color("243"), // Lighter gray
		Highlight:    lipgloss.Color("209"), // Bright red-orange
		Info:         lipgloss.Color("167"), // Red-pink
		Success:      lipgloss.Color("34"),  // Green
		Warning:      lipgloss.Color("208"), // Orange
		Error:        lipgloss.Color("160"), // Dark red
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
		Primary:      lipgloss.Color("117"), // Sky blue
		Secondary:    lipgloss.Color("120"), // Teal
		Accent:       lipgloss.Color("147"), // Purple
		Muted:        lipgloss.Color("242"), // Lighter gray
		Highlight:    lipgloss.Color("159"), // Bright cyan
		Info:         lipgloss.Color("111"), // Blue
		Success:      lipgloss.Color("120"), // Teal
		Warning:      lipgloss.Color("215"), // Orange
		Error:        lipgloss.Color("161"), // Red
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
		Primary:      lipgloss.Color("38"),  // Foam (teal)
		Secondary:    lipgloss.Color("150"), // Pine (green)
		Accent:       lipgloss.Color("183"), // Rose (pink)
		Muted:        lipgloss.Color("245"), // Muted
		Highlight:    lipgloss.Color("219"), // Iris (bright pink)
		Info:         lipgloss.Color("110"), // Iris (purple-blue)
		Success:      lipgloss.Color("150"), // Pine
		Warning:      lipgloss.Color("214"), // Gold
		Error:        lipgloss.Color("167"), // Love (red)
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
		Primary:      lipgloss.Color("111"), // Lavender
		Secondary:    lipgloss.Color("149"), // Mauve
		Accent:       lipgloss.Color("147"), // Sapphire
		Muted:        lipgloss.Color("245"), // Overlay1
		Highlight:    lipgloss.Color("183"), // Pink
		Info:         lipgloss.Color("117"), // Sky
		Success:      lipgloss.Color("149"), // Green
		Warning:      lipgloss.Color("221"), // Yellow
		Error:        lipgloss.Color("203"), // Red
	},
}

// GetStylesheet returns the stylesheet for the given theme index.
func GetStylesheet(index int) Stylesheet {
	if index < 0 || index >= len(Themes) {
		index = 0
	}
	return NewStylesheet(Themes[index])
}
