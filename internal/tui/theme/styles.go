package theme

import "github.com/charmbracelet/lipgloss"

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
	GitStyles      map[string]lipgloss.Style
	DimCol         lipgloss.Style
	KeyCol         lipgloss.Style
	SettingsHeader lipgloss.Style
	ProgressBar    lipgloss.Style
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
		GitStyles: map[string]lipgloss.Style{
			"M": lipgloss.NewStyle().Foreground(t.GitMod),
			"A": lipgloss.NewStyle().Foreground(t.GitStaged),
			"?": lipgloss.NewStyle().Foreground(t.GitUntracked),
			"U": lipgloss.NewStyle().Foreground(t.GitConflict).Bold(true),
			"D": lipgloss.NewStyle().Foreground(t.GitConflict).Bold(true),
			"!": lipgloss.NewStyle().Foreground(t.Subtle),
		},
		DimCol: lipgloss.NewStyle().Foreground(t.Subtle),
		KeyCol: lipgloss.NewStyle().Foreground(t.Dir),
		SettingsHeader: lipgloss.NewStyle().
			Foreground(t.Dir).
			Bold(true).
			Padding(0, 0, 0, 1),
		ProgressBar: lipgloss.NewStyle().Foreground(t.Dir),
	}
}
