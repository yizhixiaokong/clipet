package components

import (
	"clipet/internal/tui/styles"
	"strings"

	"charm.land/lipgloss/v2"
)

// DialogueBubble renders a speech bubble for pet dialogues.
type DialogueBubble struct {
	text  string
	style lipgloss.Style
}

// NewDialogueBubble creates a new dialogue bubble component.
func NewDialogueBubble() DialogueBubble {
	return DialogueBubble{
		style: styles.DefaultTheme().DialogueBox,
	}
}

// SetStyle overrides the default bubble style.
func (d *DialogueBubble) SetStyle(style lipgloss.Style) {
	d.style = style
}

// UpdateText sets the dialogue text for display.
func (d *DialogueBubble) UpdateText(text string) {
	d.text = text
}

// Render returns the formatted bubble with optional alignment.
func (d DialogueBubble) Render() string {
	if d.text == "" {
		return ""
	}

	// Wrap text to reasonable width for bubbles
	wrapped := d.wordWrap(d.text, 30)
	return d.style.Render(wrapped)
}

// RenderAligned positions the bubble relative to anchor points.
// position: "above" | "below" | "left" | "right"
//        offset pixels from anchor
func (d DialogueBubble) RenderAligned(position string, offset int) string {
	content := d.Render()
	if content == "" {
		return ""
	}

	// Position the bubble relative to the content
	var positioned string
	switch position {
	case "above":
		// Add some vertical spacing for floating above
		// Since we're in terminal, we'll just render the bubble normally
		// and render() will handle alignment via parent container
		positioned = content
	case "below":
		// Same as above - alignment handled by parent
		positioned = content
	case "left":
		// In terminal layout, horizontal positioning is done by parent container
		positioned = content
	case "right":
		// Same as left
		positioned = content
	default:
		positioned = content
	}

	return positioned
}

// wordWrap wraps text to max width, preserving lines.
func (d DialogueBubble) wordWrap(text string, max int) string {
	if len(text) <= max {
		return text
	}

	var result strings.Builder
	line := ""
	words := strings.Fields(text)

	for _, word := range words {
		if len(line)+len(word)+1 > max {
			if line != "" {
				result.WriteString(line)
				result.WriteString("\n")
			}
			line = word
		} else {
			if line == "" {
				line = word
			} else {
				line = line + " " + word
			}
		}
	}
	if line != "" {
		result.WriteString(line)
	}

	return result.String()
}