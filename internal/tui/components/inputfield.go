package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// InputField is a reusable text input component
type InputField struct {
	Value       string
	Placeholder string
	CursorChar  string
	Width       int
	Styles      InputStyles
	Filter      func(rune) bool // Filter function for allowed characters
}

// InputStyles holds styles for input field
type InputStyles struct {
	Text       lipgloss.Style
	Placeholder lipgloss.Style
	Cursor     lipgloss.Style
}

// DefaultInputStyles returns default styling
func DefaultInputStyles() InputStyles {
	return InputStyles{
		Text:        lipgloss.NewStyle(),
		Placeholder: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Cursor:      lipgloss.NewStyle(),
	}
}

// NewInputField creates a new input field with defaults
func NewInputField() *InputField {
	return &InputField{
		Value:       "",
		Placeholder: "",
		CursorChar:  "█",
		Width:       0, // 0 = auto
		Styles:      DefaultInputStyles(),
		Filter:      nil, // nil = allow all
	}
}

// Init implements tea.Model
func (i *InputField) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (i *InputField) Update(msg tea.Msg) (*InputField, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return i.handleKeyPress(msg)
	}
	return i, nil
}

// View implements tea.Model
func (i *InputField) View() string {
	if i.Value == "" && i.Placeholder != "" {
		return i.Styles.Placeholder.Render(i.Placeholder) + i.Styles.Cursor.Render(i.CursorChar)
	}

	text := i.Value
	if i.Width > 0 && len(text) > i.Width {
		text = text[len(text)-i.Width:]
	}

	return i.Styles.Text.Render(text) + i.Styles.Cursor.Render(i.CursorChar)
}

// handleKeyPress processes keyboard input
func (i *InputField) handleKeyPress(msg tea.KeyPressMsg) (*InputField, tea.Cmd) {
	switch msg.String() {
	case "backspace":
		if len(i.Value) > 0 {
			// Convert to []rune to properly handle UTF-8
			runes := []rune(i.Value)
			i.Value = string(runes[:len(runes)-1])
		}
	default:
		// Handle character input
		if len(msg.String()) == 1 {
			ch := rune(msg.String()[0])
			// Apply filter if set
			if i.Filter == nil || i.Filter(ch) {
				i.Value += string(ch)
			}
		}
	}
	return i, nil
}

// SetValue sets the input value
func (i *InputField) SetValue(value string) *InputField {
	i.Value = value
	return i
}

// SetPlaceholder sets the placeholder text
func (i *InputField) SetPlaceholder(placeholder string) *InputField {
	i.Placeholder = placeholder
	return i
}

// SetWidth sets the maximum visible width (0 = unlimited)
func (i *InputField) SetWidth(width int) *InputField {
	i.Width = width
	return i
}

// SetCursorChar sets the cursor character
func (i *InputField) SetCursorChar(char string) *InputField {
	i.CursorChar = char
	return i
}

// SetFilter sets the character filter function
func (i *InputField) SetFilter(filter func(rune) bool) *InputField {
	i.Filter = filter
	return i
}

// SetStyles sets the input field styles
func (i *InputField) SetStyles(styles InputStyles) *InputField {
	i.Styles = styles
	return i
}

// Clear clears the input value
func (i *InputField) Clear() *InputField {
	i.Value = ""
	return i
}

// String returns the current value as string
func (i InputField) String() string {
	return i.Value
}

// NumericFilter returns a filter that only allows digits and specific extra chars
func NumericFilter(extraChars string) func(rune) bool {
	return func(r rune) bool {
		return (r >= '0' && r <= '9') || strings.ContainsRune(extraChars, r)
	}
}
