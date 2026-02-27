package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// ProgressBar wraps bubbles/progress with convenience methods for simple use cases
type ProgressBar struct {
	width      int
	value      int
	max        int
	filledChar string
	emptyChar  string
	filledStyle lipgloss.Style
	emptyStyle  lipgloss.Style
}

// NewProgressBar creates a new progress bar with default styling
func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		width:       20,
		filledChar:  "█",
		emptyChar:   "░",
		filledStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")),
		emptyStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A4A")),
	}
}

// Render renders the progress bar as a string
func (pb *ProgressBar) Render() string {
	if pb.max <= 0 {
		return ""
	}

	// Calculate percentage
	ratio := float64(pb.value) / float64(pb.max)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	// Custom rendering using our characters
	filledWidth := int(ratio * float64(pb.width))
	emptyWidth := pb.width - filledWidth

	var bar strings.Builder
	for i := 0; i < filledWidth; i++ {
		bar.WriteString(pb.filledChar)
	}
	for i := 0; i < emptyWidth; i++ {
		bar.WriteString(pb.emptyChar)
	}

	return pb.filledStyle.Render(strings.Repeat(pb.filledChar, filledWidth)) +
		pb.emptyStyle.Render(strings.Repeat(pb.emptyChar, emptyWidth))
}

// SetValue sets the current value
func (pb *ProgressBar) SetValue(value int) *ProgressBar {
	pb.value = value
	return pb
}

// SetMax sets the maximum value
func (pb *ProgressBar) SetMax(max int) *ProgressBar {
	pb.max = max
	return pb
}

// SetWidth sets the bar width
func (pb *ProgressBar) SetWidth(width int) *ProgressBar {
	pb.width = width
	return pb
}

// SetFilledStyle sets the filled portion style
func (pb *ProgressBar) SetFilledStyle(style lipgloss.Style) *ProgressBar {
	pb.filledStyle = style
	return pb
}

// SetEmptyStyle sets the empty portion style
func (pb *ProgressBar) SetEmptyStyle(style lipgloss.Style) *ProgressBar {
	pb.emptyStyle = style
	return pb
}
