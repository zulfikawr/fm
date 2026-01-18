package ui

import (
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/tui/theme"

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
func (in Input) Value() string {
	return string(in.value)
}

// SetValue sets the text value and moves the cursor to the end.
func (in *Input) SetValue(s string) {
	in.value = []rune(s)
	in.cursor = len(in.value)
}

// Reset clears the value and resets state.
func (in *Input) Reset() {
	in.value = []rune{}
	in.cursor = 0
	in.EchoMode = EchoNormal
	in.Placeholder = ""
}

// SetCursor sets the cursor position.
func (in *Input) SetCursor(pos int) {
	if pos < 0 {
		in.cursor = 0
	} else if pos > len(in.value) {
		in.cursor = len(in.value)
	} else {
		in.cursor = pos
	}
}

// Focus sets focus on the input and starts blinking.
func (in *Input) Focus() {
	in.focused = true
	in.showCursor = true
	in.blinkID++
}

// FocusCmd sets focus and returns a command to start the blink loop.
func (in *Input) FocusCmd() tea.Cmd {
	in.Focus()
	return in.Blink()
}

// Blur removes focus from the input.
func (in *Input) Blur() {
	in.focused = false
	in.showCursor = false
	in.blinkID++
}

// Focused returns whether the input is focused.
func (in Input) Focused() bool {
	return in.focused
}

// SetPrompt sets the prompt string.
func (in *Input) SetPrompt(p string) {
	in.Prompt = p
}

// Blink returns a command that toggles the cursor visibility.
func (in Input) Blink() tea.Cmd {
	id := in.blinkID
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return BlinkMsg{ID: id}
	})
}

// SetStyles updates the styles of the input.
func (in *Input) SetStyles(styles theme.Stylesheet) {
	in.styles = styles
	bg := styles.Footer.GetBackground()
	primary := styles.KeyCol.GetForeground()
	if primary == nil {
		primary = lipgloss.Color("15")
	}
	in.TextStyle = styles.Footer.UnsetPadding().UnsetWidth().Background(bg)
	in.PromptStyle = styles.Footer.UnsetPadding().UnsetWidth().Background(bg)
	in.PlaceholderStyle = styles.DimCol.Background(bg)
	in.Cursor.Style = lipgloss.NewStyle().Background(primary).Foreground(bg)
}

// FixBackground ensures the background matches the theme.
func (in *Input) FixBackground() {
	bg := in.styles.Footer.GetBackground()
	primary := in.styles.KeyCol.GetForeground()
	if primary == nil {
		primary = lipgloss.Color("15")
	}
	in.TextStyle = in.TextStyle.Background(bg)
	in.PromptStyle = in.PromptStyle.Background(bg)
	in.PlaceholderStyle = in.PlaceholderStyle.Background(bg)
	in.Cursor.Style = in.Cursor.Style.Background(primary).Foreground(bg)
}

func (in Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	switch msg := msg.(type) {
	case BlinkMsg:
		if in.focused && msg.ID == in.blinkID {
			in.showCursor = !in.showCursor
			return in, in.Blink()
		}
		return in, nil
	}

	if !in.focused {
		return in, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		in.showCursor = true // Reset blink on keypress
		switch msg.Type {
		case tea.KeyBackspace:
			if in.cursor > 0 {
				in.value = append(in.value[:in.cursor-1], in.value[in.cursor:]...)
				in.cursor--
			}
		case tea.KeyDelete:
			if in.cursor < len(in.value) {
				in.value = append(in.value[:in.cursor], in.value[in.cursor+1:]...)
			}
		case tea.KeyLeft:
			if in.cursor > 0 {
				in.cursor--
			}
		case tea.KeyRight:
			if in.cursor < len(in.value) {
				in.cursor++
			}
		case tea.KeyHome:
			in.cursor = 0
		case tea.KeyEnd:
			in.cursor = len(in.value)
		case tea.KeyRunes:
			if in.CharLimit <= 0 || len(in.value) < in.CharLimit {
				runes := msg.Runes
				newVal := make([]rune, len(in.value)+len(runes))
				copy(newVal, in.value[:in.cursor])
				copy(newVal[in.cursor:], runes)
				copy(newVal[in.cursor+len(runes):], in.value[in.cursor:])
				in.value = newVal
				in.cursor += len(runes)
			}
		case tea.KeySpace:
			if in.CharLimit <= 0 || len(in.value) < in.CharLimit {
				newVal := make([]rune, len(in.value)+1)
				copy(newVal, in.value[:in.cursor])
				newVal[in.cursor] = ' '
				copy(newVal[in.cursor+1:], in.value[in.cursor:])
				in.value = newVal
				in.cursor++
			}
		}
	}

	return in, nil
}

// View renders the input component.
func (in Input) View() string {
	var b strings.Builder

	// Render prompt
	if in.Prompt != "" {
		b.WriteString(in.PromptStyle.Render(in.Prompt))
	}

	// Render value with cursor
	val := in.value
	switch in.EchoMode {
	case EchoPassword:
		masked := make([]rune, len(in.value))
		for idx := range masked {
			masked[idx] = in.EchoCharacter
		}
		val = masked
	case EchoNone:
		val = []rune{}
	}

	// Simple scrolling/clipping if Width is set
	displayOffset := 0
	if in.Width > 0 && in.cursor >= in.Width {
		displayOffset = in.cursor - in.Width + 1
	}

	for idx := displayOffset; ; idx++ {
		// Stop if we reach the Width limit or the end of the value
		if in.Width > 0 && idx-displayOffset >= in.Width {
			break
		}

		if idx == in.cursor && in.focused {
			char := " "
			if idx < len(val) {
				char = string(val[idx])
			}
			if in.showCursor {
				b.WriteString(in.Cursor.Style.Render(char))
			} else {
				b.WriteString(in.TextStyle.Render(char))
			}
			if idx >= len(val) {
				break
			}
			continue
		}

		if idx < len(val) {
			b.WriteString(in.TextStyle.Render(string(val[idx])))
		} else {
			if idx == displayOffset && len(val) == 0 && in.Placeholder != "" {
				b.WriteString(in.PlaceholderStyle.Render(in.Placeholder))
				break
			}
			// Fill remaining Width with spaces if focused (for trailing cursor)
			if !in.focused || idx > in.cursor {
				break
			}
		}

		if idx >= len(val) && (!in.focused || idx >= in.cursor) {
			break
		}
	}

	return b.String()
}
