package ui

import (
	"strings"
	"time"

	"fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BlinkMsg is sent to toggle cursor visibility for blinking
type BlinkMsg struct {
	ID int
}

// Cursor holds styling for the input cursor
type Cursor struct {
	Style lipgloss.Style
}

type EchoMode int

const (
	EchoNormal EchoMode = iota
	EchoPassword
	EchoNone
)

// Input is a custom text input component implemented from scratch.
type Input struct {
	value       []rune
	cursor      int
	Width       int
	Prompt      string
	focused     bool
	CharLimit   int
	Placeholder string
	showCursor  bool
	blinkID     int

	// Echo configuration
	EchoMode      EchoMode
	EchoCharacter rune

	// Styles
	TextStyle        lipgloss.Style
	PromptStyle      lipgloss.Style
	PlaceholderStyle lipgloss.Style
	Cursor           Cursor

	styles theme.Stylesheet
}

// NewInput creates a new custom text input.
func NewInput(styles theme.Stylesheet) Input {
	bg := styles.Footer.GetBackground()
	primary := styles.KeyCol.GetForeground()
	if primary == nil {
		primary = lipgloss.Color("15") // Fallback to white
	}

	return Input{
		Width:            30,
		CharLimit:        256,
		styles:           styles,
		showCursor:       true,
		EchoMode:         EchoNormal,
		EchoCharacter:    '*',
		Cursor:           Cursor{Style: lipgloss.NewStyle().Background(primary).Foreground(bg)},
		TextStyle:        styles.Footer.UnsetPadding().UnsetWidth().Background(bg),
		PromptStyle:      styles.Footer.UnsetPadding().UnsetWidth().Background(bg),
		PlaceholderStyle: styles.DimCol.Background(bg),
	}
}

// Value returns the current text value.
func (i Input) Value() string {
	return string(i.value)
}

// SetValue sets the text value and moves the cursor to the end.
func (i *Input) SetValue(s string) {
	i.value = []rune(s)
	i.cursor = len(i.value)
}

// Reset clears the value and resets state.
func (i *Input) Reset() {
	i.value = []rune{}
	i.cursor = 0
	i.EchoMode = EchoNormal
	i.Placeholder = ""
}

// SetCursor sets the cursor position.
func (i *Input) SetCursor(pos int) {
	if pos < 0 {
		i.cursor = 0
	} else if pos > len(i.value) {
		i.cursor = len(i.value)
	} else {
		i.cursor = pos
	}
}

// Focus sets focus on the input and starts blinking.
func (i *Input) Focus() {
	i.focused = true
	i.showCursor = true
	i.blinkID++
}

// FocusCmd sets focus and returns a command to start the blink loop.
func (i *Input) FocusCmd() tea.Cmd {
	i.Focus()
	return i.Blink()
}

// Blur removes focus from the input.
func (i *Input) Blur() {
	i.focused = false
	i.showCursor = false
	i.blinkID++
}

// Focused returns whether the input is focused.
func (i Input) Focused() bool {
	return i.focused
}

// SetPrompt sets the prompt string.
func (i *Input) SetPrompt(p string) {
	i.Prompt = p
}

// Blink returns a command that toggles the cursor visibility.
func (i Input) Blink() tea.Cmd {
	id := i.blinkID
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return BlinkMsg{ID: id}
	})
}

// SetStyles updates the styles of the input.
func (i *Input) SetStyles(styles theme.Stylesheet) {
	i.styles = styles
	bg := styles.Footer.GetBackground()
	primary := styles.KeyCol.GetForeground()
	if primary == nil {
		primary = lipgloss.Color("15")
	}
	i.TextStyle = styles.Footer.UnsetPadding().UnsetWidth().Background(bg)
	i.PromptStyle = styles.Footer.UnsetPadding().UnsetWidth().Background(bg)
	i.PlaceholderStyle = styles.DimCol.Background(bg)
	i.Cursor.Style = lipgloss.NewStyle().Background(primary).Foreground(bg)
}

// FixBackground ensures the background matches the theme.
func (i *Input) FixBackground() {
	bg := i.styles.Footer.GetBackground()
	primary := i.styles.KeyCol.GetForeground()
	if primary == nil {
		primary = lipgloss.Color("15")
	}
	i.TextStyle = i.TextStyle.Background(bg)
	i.PromptStyle = i.PromptStyle.Background(bg)
	i.PlaceholderStyle = i.PlaceholderStyle.Background(bg)
	i.Cursor.Style = i.Cursor.Style.Background(primary).Foreground(bg)
}

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	switch msg := msg.(type) {
	case BlinkMsg:
		if i.focused && msg.ID == i.blinkID {
			i.showCursor = !i.showCursor
			return i, i.Blink()
		}
		return i, nil
	}

	if !i.focused {
		return i, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		i.showCursor = true // Reset blink on keypress
		switch msg.Type {
		case tea.KeyBackspace:
			if i.cursor > 0 {
				i.value = append(i.value[:i.cursor-1], i.value[i.cursor:]...)
				i.cursor--
			}
		case tea.KeyDelete:
			if i.cursor < len(i.value) {
				i.value = append(i.value[:i.cursor], i.value[i.cursor+1:]...)
			}
		case tea.KeyLeft:
			if i.cursor > 0 {
				i.cursor--
			}
		case tea.KeyRight:
			if i.cursor < len(i.value) {
				i.cursor++
			}
		case tea.KeyHome:
			i.cursor = 0
		case tea.KeyEnd:
			i.cursor = len(i.value)
		case tea.KeyRunes:
			if i.CharLimit <= 0 || len(i.value) < i.CharLimit {
				runes := msg.Runes
				newVal := make([]rune, len(i.value)+len(runes))
				copy(newVal, i.value[:i.cursor])
				copy(newVal[i.cursor:], runes)
				copy(newVal[i.cursor+len(runes):], i.value[i.cursor:])
				i.value = newVal
				i.cursor += len(runes)
			}
		case tea.KeySpace:
			if i.CharLimit <= 0 || len(i.value) < i.CharLimit {
				newVal := make([]rune, len(i.value)+1)
				copy(newVal, i.value[:i.cursor])
				newVal[i.cursor] = ' '
				copy(newVal[i.cursor+1:], i.value[i.cursor:])
				i.value = newVal
				i.cursor++
			}
		}
	}

	return i, nil
}

// View renders the input component.
func (i Input) View() string {
	var b strings.Builder

	// Render prompt
	if i.Prompt != "" {
		b.WriteString(i.PromptStyle.Render(i.Prompt))
	}

	// Render value with cursor
	val := i.value
	switch i.EchoMode {
	case EchoPassword:
		masked := make([]rune, len(i.value))
		for idx := range masked {
			masked[idx] = i.EchoCharacter
		}
		val = masked
	case EchoNone:
		val = []rune{}
	}

	// Simple scrolling/clipping if Width is set
	displayOffset := 0
	if i.Width > 0 && i.cursor >= i.Width {
		displayOffset = i.cursor - i.Width + 1
	}

	for idx := displayOffset; ; idx++ {
		// Stop if we reach the Width limit or the end of the value
		if i.Width > 0 && idx-displayOffset >= i.Width {
			break
		}

		if idx == i.cursor && i.focused {
			char := " "
			if idx < len(val) {
				char = string(val[idx])
			}
			if i.showCursor {
				b.WriteString(i.Cursor.Style.Render(char))
			} else {
				b.WriteString(i.TextStyle.Render(char))
			}
			if idx >= len(val) {
				break
			}
			continue
		}

		if idx < len(val) {
			b.WriteString(i.TextStyle.Render(string(val[idx])))
		} else {
			if idx == displayOffset && len(val) == 0 && i.Placeholder != "" {
				b.WriteString(i.PlaceholderStyle.Render(i.Placeholder))
				break
			}
			// Fill remaining Width with spaces if focused (for trailing cursor)
			if !i.focused || idx > i.cursor {
				break
			}
		}

		if idx >= len(val) && (!i.focused || idx >= i.cursor) {
			break
		}
	}

	return b.String()
}
