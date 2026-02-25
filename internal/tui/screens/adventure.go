package screens

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/tui/styles"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// AdventurePhase tracks the current phase of an adventure.
type AdventurePhase int

const (
	AdventureIntro    AdventurePhase = iota // show description
	AdventureChoosing                       // player picks a choice
	AdventureResolving                      // brief animation
	AdventureResult                         // show outcome + effects
)

// AdventureModel is the adventure event screen.
type AdventureModel struct {
	pet       *game.Pet
	adventure plugin.Adventure
	theme     styles.Theme

	phase     AdventurePhase
	choiceIdx int
	outcome   *plugin.AdventureOutcome
	changes   map[string][2]int
	animTick  int
	width     int
	height    int
	done      bool
}

// NewAdventureModel creates a new adventure screen for the given adventure.
func NewAdventureModel(pet *game.Pet, adv plugin.Adventure, theme styles.Theme) AdventureModel {
	return AdventureModel{
		pet:       pet,
		adventure: adv,
		theme:     theme,
		phase:     AdventureIntro,
	}
}

// SetSize updates terminal dimensions.
func (a AdventureModel) SetSize(w, h int) AdventureModel {
	a.width = w
	a.height = h
	return a
}

// IsDone returns true when the adventure is complete.
func (a AdventureModel) IsDone() bool {
	return a.done
}

// Tick advances the resolving animation.
func (a AdventureModel) Tick() AdventureModel {
	if a.phase == AdventureResolving {
		a.animTick++
		if a.animTick >= 4 {
			// Resolve and apply
			choice := a.adventure.Choices[a.choiceIdx]
			outcome := game.ResolveOutcome(choice)
			a.outcome = &outcome
			a.changes = game.ApplyAdventureOutcome(a.pet, outcome)
			a.pet.LastAdventureAt = time.Now()
			a.phase = AdventureResult
		}
	}
	return a
}

// Update handles key input for the adventure screen.
func (a AdventureModel) Update(msg tea.Msg) (AdventureModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch a.phase {
		case AdventureIntro:
			switch msg.String() {
			case "enter", " ":
				a.phase = AdventureChoosing
			case "escape":
				a.done = true
			}

		case AdventureChoosing:
			switch msg.String() {
			case "up", "k":
				if a.choiceIdx > 0 {
					a.choiceIdx--
				}
			case "down", "j":
				if a.choiceIdx < len(a.adventure.Choices)-1 {
					a.choiceIdx++
				}
			case "enter":
				a.phase = AdventureResolving
				a.animTick = 0
			case "escape":
				a.phase = AdventureIntro
			}

		case AdventureResult:
			if msg.String() == "enter" || msg.String() == " " {
				a.done = true
			}
		}
	}
	return a, nil
}

// View renders the adventure screen.
func (a AdventureModel) View() string {
	w := a.width
	if w < 50 {
		w = 50
	}

	switch a.phase {
	case AdventureIntro:
		return a.viewIntro(w)
	case AdventureChoosing:
		return a.viewChoosing(w)
	case AdventureResolving:
		return a.viewResolving(w)
	case AdventureResult:
		return a.viewResult(w)
	}
	return ""
}

func (a AdventureModel) viewIntro(w int) string {
	title := a.theme.EvolveTitle.
		Background(lipgloss.Color("#7D56F4")).
		Width(w - 2).
		Render("🗺️ " + a.adventure.Name)

	desc := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Width(w - 8).
		Render(a.adventure.Description)

	hint := lipgloss.NewStyle().
		Foreground(styles.DimColor()).
		Italic(true).
		Render("精力消耗: -10")

	help := a.theme.HelpBar.Render("Enter 继续  Esc 放弃")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		desc,
		"",
		hint,
		"",
		help,
	)
}

func (a AdventureModel) viewChoosing(w int) string {
	title := a.theme.EvolveTitle.
		Background(lipgloss.Color("#7D56F4")).
		Width(w - 2).
		Render("🗺️ " + a.adventure.Name)

	prompt := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Bold(true).
		Render("你要怎么做？")

	var choices []string
	for i, c := range a.adventure.Choices {
		label := c.Text
		if i == a.choiceIdx {
			choices = append(choices, a.theme.ActionCellSelected.Width(w-6).Render("▸ "+label))
		} else {
			choices = append(choices, a.theme.ActionCell.Width(w-6).Render("  "+label))
		}
	}

	help := a.theme.HelpBar.Render("↑↓ 选择  Enter 确认  Esc 返回")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		prompt,
		"",
		lipgloss.JoinVertical(lipgloss.Left, choices...),
		"",
		help,
	)
}

func (a AdventureModel) viewResolving(w int) string {
	title := a.theme.EvolveTitle.
		Background(lipgloss.Color("#7D56F4")).
		Width(w - 2).
		Render("🗺️ " + a.adventure.Name)

	frames := []string{"⏳", "⌛", "⏳", "⌛"}
	icon := frames[a.animTick%len(frames)]

	anim := lipgloss.NewStyle().
		Foreground(styles.GoldColor()).
		Bold(true).
		Width(w - 4).
		Align(lipgloss.Center).
		Render(icon + " 冒险中...")

	progress := a.animTick
	if progress > 4 {
		progress = 4
	}
	filled := strings.Repeat("█", progress)
	empty := strings.Repeat("░", 4-progress)
	bar := lipgloss.NewStyle().Foreground(styles.GoldColor()).Render(filled) +
		lipgloss.NewStyle().Foreground(styles.DimColor()).Render(empty)

	return lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		"",
		anim,
		"",
		lipgloss.NewStyle().Width(w-4).Align(lipgloss.Center).Render(bar),
		"",
	)
}

func (a AdventureModel) viewResult(w int) string {
	if a.outcome == nil {
		return "冒险结束"
	}

	title := a.theme.EvolveTitle.
		Background(lipgloss.Color("#7D56F4")).
		Width(w - 2).
		Render("🗺️ 冒险结果")

	outcomeText := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Width(w - 8).
		Render(a.outcome.Text)

	// Build effect display
	var effectLines []string
	attrNames := map[string]string{
		"hunger":    "饱腹",
		"happiness": "快乐",
		"health":    "健康",
		"energy":    "精力",
	}
	for attr, vals := range a.changes {
		name := attrNames[attr]
		if name == "" {
			name = attr
		}
		delta := vals[1] - vals[0]
		sign := "+"
		color := lipgloss.Color("#04B575") // green
		if delta < 0 {
			sign = ""
			color = lipgloss.Color("#E94560") // red
		}
		effectLines = append(effectLines,
			lipgloss.NewStyle().Foreground(color).Render(
				fmt.Sprintf("  %s %d → %d (%s%d)", name, vals[0], vals[1], sign, delta)))
	}

	effectBlock := ""
	if len(effectLines) > 0 {
		effectBlock = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.TextColor()).
			Render("属性变化：") + "\n" + strings.Join(effectLines, "\n")
	} else {
		effectBlock = lipgloss.NewStyle().
			Foreground(styles.DimColor()).
			Render("没有属性变化")
	}

	help := a.theme.HelpBar.Render("Enter 返回")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		outcomeText,
		"",
		effectBlock,
		"",
		help,
	)
}
