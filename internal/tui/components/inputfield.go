package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

// InputField is a wrapper around bubbles/textinput for simplified usage
type InputField struct {
	ti     textinput.Model
	Filter func(rune) bool // Custom filter function
}

// NewInputField creates a new input field with defaults
func NewInputField() *InputField {
	ti := textinput.New()
	ti.Focus()

	return &InputField{
		ti: ti,
	}
}

// Init implements tea.Model
func (i *InputField) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (i *InputField) Update(msg tea.Msg) (*InputField, tea.Cmd) {
	var cmd tea.Cmd

	// Handle character input with custom filter
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if len(keyMsg.String()) == 1 {
			ch := rune(keyMsg.String()[0])
			if i.Filter != nil && !i.Filter(ch) {
				// Reject this character by not passing it to textinput
				return i, nil
			}
		}
	}

	i.ti, cmd = i.ti.Update(msg)
	return i, cmd
}

// View implements tea.Model
func (i *InputField) View() string {
	return i.ti.View()
}

// SetValue sets the input value
func (i *InputField) SetValue(value string) *InputField {
	i.ti.SetValue(value)
	return i
}

// SetPlaceholder sets the placeholder text
func (i *InputField) SetPlaceholder(placeholder string) *InputField {
	i.ti.Placeholder = placeholder
	return i
}

// SetWidth sets the maximum visible width (0 = unlimited)
func (i *InputField) SetWidth(width int) *InputField {
	i.ti.SetWidth(width)
	return i
}

// SetCursorChar sets the cursor character (note: bubbles uses built-in cursor)
func (i *InputField) SetCursorChar(char string) *InputField {
	// bubbles/textinput uses a built-in cursor, this is kept for API compatibility
	return i
}

// SetFilter sets the character filter function
func (i *InputField) SetFilter(filter func(rune) bool) *InputField {
	i.Filter = filter
	return i
}

// SetStyles sets the input field styles (note: limited support)
func (i *InputField) SetStyles(styles InputStyles) *InputField {
	// bubbles/textinput has its own styling system
	// For simplicity, we keep default styles
	// Full customization would require converting InputStyles to textinput.Styles
	return i
}

// Clear clears the input value
func (i *InputField) Clear() *InputField {
	i.ti.SetValue("")
	return i
}

// String returns the current value as string
func (i InputField) String() string {
	return i.ti.Value()
}

// Value returns the current value
func (i *InputField) Value() string {
	return i.ti.Value()
}

// NumericFilter returns a filter that only allows digits and specific extra chars
func NumericFilter(extraChars string) func(rune) bool {
	return func(r rune) bool {
		return (r >= '0' && r <= '9') || containsRune(extraChars, r)
	}
}

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}

// InputStyles is kept for API compatibility (not used with bubbles/textinput)
type InputStyles struct {
	Text        lipgloss.Style
	Placeholder lipgloss.Style
	Cursor      lipgloss.Style
}
