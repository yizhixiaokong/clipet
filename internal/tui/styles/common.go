package styles

import "charm.land/lipgloss/v2"

// DevCommandStyles contains common styles for clipet-dev commands
var DevCommandStyles = struct {
	// Panel is the base panel style
	Panel lipgloss.Style

	// Title is for panel headers
	Title lipgloss.Style

	// Info is for secondary information
	Info lipgloss.Style

	// Error is for error messages
	Error lipgloss.Style

	// Success is for success messages
	Success lipgloss.Style

	// Warning is for warning messages
	Warning lipgloss.Style

	// Selected is for selected items
	Selected lipgloss.Style

	// InputLabel is for input field labels
	InputLabel lipgloss.Style

	// BarFilled is for progress bar filled portion
	BarFilled lipgloss.Style

	// BarEmpty is for progress bar empty portion
	BarEmpty lipgloss.Style

	// Item is for list items
	Item lipgloss.Style

	// SelItem is for selected list items (cursor highlight)
	SelItem lipgloss.Style
}{
	Panel: lipgloss.NewStyle().
		Padding(0, 1),

	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1),

	Info: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555570")),

	Error: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6347")).
		Bold(true),

	Success: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true),

	Warning: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true),

	Selected: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true),

	InputLabel: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true),

	BarFilled: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")),

	BarEmpty: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2A2A4A")),

	Item: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EAEAEA")),

	SelItem: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true),
}

// MakeTitleStyle creates a title style with custom background color
func MakeTitleStyle(bgColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(bgColor)).
		Padding(0, 1)
}
