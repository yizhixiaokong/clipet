package screens

import (
	"clipet/internal/game"
	"clipet/internal/i18n"
	"clipet/internal/tui/keys"
	"clipet/internal/tui/styles"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// OfflineSettlementModel is the offline settlement report screen.
type OfflineSettlementModel struct {
	results []game.DecayRoundResult
	theme   styles.Theme
	i18n    *i18n.Manager
	keyMap keys.OfflineSettlementKeyMap
	help   help.Model

	scrollOffset int // Current scroll position
	maxVisible   int // Max visible lines (calculated from height)
	width        int
	height       int
	done         bool
}

// Style helpers
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	dangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAEAEA"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555570"))
)

// NewOfflineSettlementModel creates a new offline settlement screen.
func NewOfflineSettlementModel(results []game.DecayRoundResult, theme styles.Theme, i18nMgr *i18n.Manager) OfflineSettlementModel {
	return OfflineSettlementModel{
		results: results,
		theme:   theme,
		i18n:    i18nMgr,
		keyMap:  keys.NewOfflineSettlementKeyMap(i18nMgr),
		help:    help.New(),
	}
}

// SetSize updates terminal dimensions.
func (m OfflineSettlementModel) SetSize(w, h int) OfflineSettlementModel {
	m.width = w
	m.height = h
	// Reserve space for header (3 lines) and footer (3 lines)
	m.maxVisible = h - 6
	if m.maxVisible < 5 {
		m.maxVisible = 5
	}
	// Adjust scroll offset if needed
	m.scrollOffset = clamp(m.scrollOffset, 0, m.maxScroll())
	return m
}

// IsDone returns true when the user dismisses the screen.
func (m OfflineSettlementModel) IsDone() bool {
	return m.done
}

// Update handles key input.
func (m OfflineSettlementModel) Update(msg tea.Msg) (OfflineSettlementModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			m.done = true
		case key.Matches(msg, m.keyMap.Up):
			m.scrollOffset = clamp(m.scrollOffset-1, 0, m.maxScroll())
		case key.Matches(msg, m.keyMap.Down):
			m.scrollOffset = clamp(m.scrollOffset+1, 0, m.maxScroll())
		case key.Matches(msg, m.keyMap.Top):
			m.scrollOffset = 0
		case key.Matches(msg, m.keyMap.Bottom):
			m.scrollOffset = m.maxScroll()
		}
	}
	return m, nil
}

// View renders the offline settlement report.
func (m OfflineSettlementModel) View() string {
	if len(m.results) == 0 {
		return m.i18n.T("ui.offline_settlement.no_data")
	}

	var b strings.Builder

	// Header
	title := titleStyle.Render(m.i18n.T("ui.offline_settlement.title"))
	b.WriteString(title + "\n\n")

	// Summary line
	criticalCount := 0
	for _, r := range m.results {
		if r.CriticalState {
			criticalCount++
		}
	}

	summaryText := m.i18n.T("ui.offline_settlement.summary", "rounds", len(m.results), "critical", criticalCount)
	if criticalCount > 0 {
		summaryText = warningStyle.Render(summaryText)
	}
	b.WriteString("  " + summaryText + "\n\n")

	// Build all content lines
	var lines []string

	for i, r := range m.results {
		// Round header
		roundHeader := m.i18n.T("ui.offline_settlement.round_header", "round", r.Round, "duration", fmt.Sprintf("%.1f", r.Duration.Hours()))
		if r.CriticalState {
			roundHeader = dangerStyle.Render(roundHeader)
		} else {
			roundHeader = textStyle.Render(roundHeader)
		}
		lines = append(lines, "  "+roundHeader)

		// Effects
		for _, effect := range r.Effects {
			effectLine := "    • " + effect
			if strings.Contains(effect, "⚠️") || strings.Contains(effect, "🚨") || strings.Contains(effect, "💔") {
				effectLine = dangerStyle.Render(effectLine)
			} else if strings.Contains(effect, "🛡️") {
				effectLine = successStyle.Render(effectLine)
			} else {
				effectLine = textStyle.Render(effectLine)
			}
			lines = append(lines, effectLine)
		}

		// Attribute changes
		beforeStr := fmt.Sprintf("%2d,%2d,%2d,%2d",
			r.StartAttrs[0], r.StartAttrs[1], r.StartAttrs[2], r.StartAttrs[3])
		afterStr := fmt.Sprintf("%2d,%2d,%2d,%2d",
			r.EndAttrs[0], r.EndAttrs[1], r.EndAttrs[2], r.EndAttrs[3])
		attrLine := m.i18n.T("ui.offline_settlement.attr_line", "before", beforeStr, "after", afterStr)
		lines = append(lines, "    "+mutedStyle.Render(attrLine))

		// Add blank line between rounds (except last)
		if i < len(m.results)-1 {
			lines = append(lines, "")
		}
	}

	// Apply scrolling
	totalLines := len(lines)
	startIdx := m.scrollOffset
	endIdx := startIdx + m.maxVisible
	if endIdx > totalLines {
		endIdx = totalLines
	}

	visibleLines := lines[startIdx:endIdx]
	for _, line := range visibleLines {
		b.WriteString(line + "\n")
	}

	// Footer with navigation hints
	b.WriteString("\n")
	footer := mutedStyle.Render(m.i18n.T("ui.offline_settlement.footer"))

	// Scroll indicator
	if totalLines > m.maxVisible {
		scrollInfo := fmt.Sprintf(" [%d/%d]", startIdx+1, totalLines)
		footer += mutedStyle.Render(scrollInfo)
	}

	// Right-align footer
	padding := m.width - lipgloss.Width(footer)
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding/2))
	}
	b.WriteString(footer + "\n")

	// Warning if critical states detected
	if criticalCount > 0 {
		b.WriteString("\n")
		warning := warningStyle.Render(m.i18n.T("ui.offline_settlement.critical_warning", "count", criticalCount))
		b.WriteString(warning + "\n")
	}

	return b.String()
}

// maxScroll returns the maximum scroll offset.
func (m OfflineSettlementModel) maxScroll() int {
	// Count total lines (approximate)
	totalLines := 0
	for i, r := range m.results {
		totalLines += 1 // Round header
		totalLines += len(r.Effects)
		totalLines += 1 // Attribute line
		if i < len(m.results)-1 {
			totalLines += 1 // Blank line
		}
	}

	maxScroll := totalLines - m.maxVisible
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
