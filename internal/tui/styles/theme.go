// Package styles provides Lipgloss style definitions for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// Theme holds all style definitions for the TUI.
type Theme struct {
	Title            lipgloss.Style
	StatusBar        lipgloss.Style
	PetBox           lipgloss.Style
	StatLabel        lipgloss.Style
	StatFilled       lipgloss.Style
	StatEmpty        lipgloss.Style
	Dialogue         lipgloss.Style
	Help             lipgloss.Style
	MoodHappy        lipgloss.Style
	MoodNormal       lipgloss.Style
	MoodSad          lipgloss.Style
	MoodMiserable    lipgloss.Style
	MenuItem         lipgloss.Style
	MenuItemSelected lipgloss.Style
}

// DefaultTheme returns the default color theme.
func DefaultTheme() Theme {
	return Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Background(lipgloss.Color("#2A2A2A")).
			Padding(0, 1),

		PetBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2),

		StatLabel: lipgloss.NewStyle().
			Width(6).
			Foreground(lipgloss.Color("#FAFAFA")),

		StatFilled: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),

		StatEmpty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C3C3C")),

		Dialogue: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFD700")).
			Padding(0, 1).
			Foreground(lipgloss.Color("#FFD700")),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),

		MoodHappy: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),

		MoodNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")),

		MoodSad: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6347")),

		MoodMiserable: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true),

		MenuItem: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("#A0A0A0")),

		MenuItemSelected: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),
	}
}
