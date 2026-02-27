// Package styles provides Lipgloss style definitions for the TUI.
package styles

import (
	"charm.land/lipgloss/v2"
)

// Theme holds all style definitions for the TUI.
type Theme struct {
	// Layout
	TitleBar lipgloss.Style
	HelpBar  lipgloss.Style

	// Pet art panel (left)
	PetPanel lipgloss.Style

	// Status panel (right)
	StatusPanel  lipgloss.Style
	StatusName   lipgloss.Style
	StatusLabel  lipgloss.Style
	StatusValue  lipgloss.Style
	StatFilled   lipgloss.Style
	StatEmpty    lipgloss.Style
	StatLabel    lipgloss.Style
	SectionTitle lipgloss.Style

	// Dialogue / message area
	DialogueBox lipgloss.Style
	MessageBox  lipgloss.Style

	// Action menu (2x2 grid)
	ActionCell         lipgloss.Style
	ActionCellSelected lipgloss.Style

	// Category tabs (two-level menu)
	CategoryTab       lipgloss.Style
	CategoryTabActive lipgloss.Style
	CategoryTabOpen   lipgloss.Style

	// Game overlay
	GamePanel lipgloss.Style

	// Evolve screen
	EvolveTitle lipgloss.Style
	EvolveArt   lipgloss.Style

	// Mood colors
	MoodHappy     lipgloss.Style
	MoodNormal    lipgloss.Style
	MoodSad       lipgloss.Style
	MoodMiserable lipgloss.Style
}

// Color palette.
var (
	colorPrimary    = lipgloss.Color("#7D56F4")
	colorBg         = lipgloss.Color("#1A1A2E")
	colorPanelBg    = lipgloss.Color("#16213E")
	colorAccent     = lipgloss.Color("#E94560")
	colorGold       = lipgloss.Color("#FFD700")
	colorText       = lipgloss.Color("#EAEAEA")
	colorDim        = lipgloss.Color("#555570")
	colorGreen      = lipgloss.Color("#04B575")
	colorBarEmpty   = lipgloss.Color("#2A2A4A")
	colorSelected   = lipgloss.Color("#7D56F4")
	colorSelectedFg = lipgloss.Color("#FFFFFF")
)

// DefaultTheme returns the default color theme.
func DefaultTheme() Theme {
	border := lipgloss.RoundedBorder()

	return Theme{
		TitleBar: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSelectedFg).
			Background(colorPrimary).
			Padding(0, 2),

		HelpBar: lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1),

		PetPanel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorPrimary).
			Padding(1, 2),

		StatusPanel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorPrimary).
			Padding(1, 2),

		StatusName: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText),

		StatusLabel: lipgloss.NewStyle().
			Foreground(colorDim).
			Width(6),

		StatusValue: lipgloss.NewStyle().
			Foreground(colorText),

		StatLabel: lipgloss.NewStyle().
			Foreground(colorText).
			Width(6),

		StatFilled: lipgloss.NewStyle().
			Background(colorGreen),

		StatEmpty: lipgloss.NewStyle().
			Background(colorBarEmpty),

		SectionTitle: lipgloss.NewStyle().
			Foreground(colorGold).
			Bold(true),

		DialogueBox: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorGold).
			Foreground(colorGold).
			Padding(0, 2),

		MessageBox: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorGreen).
			Foreground(colorGreen).
			Padding(0, 2),

		ActionCell: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorDim).
			Foreground(colorDim).
			Padding(0, 1).
			Align(lipgloss.Center),

		ActionCellSelected: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorSelected).
			Foreground(colorSelectedFg).
			Background(colorSelected).
			Bold(true).
			Padding(0, 1).
			Align(lipgloss.Center),

		CategoryTab: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorDim).
			Foreground(colorDim).
			Padding(0, 1).
			Align(lipgloss.Center),

		CategoryTabActive: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorGold).
			Foreground(colorGold).
			Bold(true).
			Padding(0, 1).
			Align(lipgloss.Center),

		CategoryTabOpen: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorGold).
			Foreground(colorSelectedFg).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 1).
			Align(lipgloss.Center),

		GamePanel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(colorGold).
			Foreground(colorText).
			Padding(1, 3),

		EvolveTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSelectedFg).
			Background(colorAccent).
			Padding(0, 2),

		EvolveArt: lipgloss.NewStyle().
			Foreground(colorGold).
			Bold(true).
			Align(lipgloss.Center),

		MoodHappy: lipgloss.NewStyle().
			Foreground(colorGreen),

		MoodNormal: lipgloss.NewStyle().
			Foreground(colorText),

		MoodSad: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6347")),

		MoodMiserable: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true),
	}
}
