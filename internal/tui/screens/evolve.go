package screens

import (
	"clipet/internal/game"
	"clipet/internal/tui/styles"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// EvolvePhase tracks the current display phase of the evolution screen.
type EvolvePhase int

const (
	EvolveChoosing  EvolvePhase = iota // player picks among candidates
	EvolveAnimating                    // playing the evolution animation
	EvolveDone                         // showing result
)

// EvolveModel is the evolution selection/animation screen.
type EvolveModel struct {
	pet        *game.Pet
	candidates []game.EvolveCandidate
	theme      styles.Theme

	phase      EvolvePhase
	choiceIdx  int
	animTick   int
	result     *game.EvolveCandidate
	oldStageID string
	width      int
	height     int
	done       bool
}

// NewEvolveModel creates a new evolution screen.
func NewEvolveModel(pet *game.Pet, candidates []game.EvolveCandidate, theme styles.Theme) EvolveModel {
	phase := EvolveChoosing
	if len(candidates) == 1 {
		phase = EvolveAnimating
	}
	return EvolveModel{
		pet:        pet,
		candidates: candidates,
		theme:      theme,
		phase:      phase,
		oldStageID: pet.StageID,
	}
}

// SetSize updates terminal dimensions.
func (e EvolveModel) SetSize(w, h int) EvolveModel {
	e.width = w
	e.height = h
	return e
}

// IsDone returns true when the evolution sequence is complete.
func (e EvolveModel) IsDone() bool {
	return e.done
}

// Result returns the chosen evolution candidate after completion.
func (e EvolveModel) Result() *game.EvolveCandidate {
	return e.result
}

// Tick advances the animation by one frame.
func (e EvolveModel) Tick() EvolveModel {
	if e.phase == EvolveAnimating {
		e.animTick++
		if e.animTick >= 6 {
			if e.result == nil && len(e.candidates) == 1 {
				e.result = &e.candidates[0]
			}
			if e.result != nil {
				game.DoEvolve(e.pet, *e.result)
				e.phase = EvolveDone
			}
		}
	}
	return e
}

// Update handles input for the evolution screen.
func (e EvolveModel) Update(msg tea.Msg) (EvolveModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch e.phase {
		case EvolveChoosing:
			switch msg.String() {
			case "up", "k":
				if e.choiceIdx > 0 {
					e.choiceIdx--
				}
			case "down", "j":
				if e.choiceIdx < len(e.candidates)-1 {
					e.choiceIdx++
				}
			case "enter":
				e.result = &e.candidates[e.choiceIdx]
				e.phase = EvolveAnimating
				e.animTick = 0
			}

		case EvolveDone:
			if msg.String() == "enter" || msg.String() == " " {
				e.done = true
			}
		}
	}
	return e, nil
}

// View renders the evolution screen.
func (e EvolveModel) View() string {
	switch e.phase {
	case EvolveChoosing:
		return e.viewChoosing()
	case EvolveAnimating:
		return e.viewAnimating()
	case EvolveDone:
		return e.viewDone()
	}
	return ""
}

func (e EvolveModel) viewChoosing() string {
	w := e.width
	if w < 40 {
		w = 40
	}

	title := e.theme.EvolveTitle.Width(w - 2).Render("✨ 进化！")

	desc := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Render(fmt.Sprintf("%s 可以进化了！请选择进化方向：", e.pet.Name))

	var choices []string
	for i, c := range e.candidates {
		label := fmt.Sprintf("%s (%s)", c.ToStage.Name, c.ToStage.Phase)
		if i == e.choiceIdx {
			choices = append(choices, e.theme.ActionCellSelected.Width(w-6).Render("▸ "+label))
		} else {
			choices = append(choices, e.theme.ActionCell.Width(w-6).Render("  "+label))
		}
	}

	help := e.theme.HelpBar.Render("↑↓ 选择  Enter 确认")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		desc,
		"",
		lipgloss.JoinVertical(lipgloss.Left, choices...),
		"",
		help,
	)
}

func (e EvolveModel) viewAnimating() string {
	frames := []string{
		"       ✦       ",
		"      ✦ ✦      ",
		"    ✦  ✦  ✦    ",
		"   ✦ ✦   ✦ ✦   ",
		"    ✦ ✦ ✦ ✦    ",
		"      ✦ ✦      ",
	}

	idx := e.animTick % len(frames)
	art := frames[idx]

	name := ""
	if e.result != nil {
		name = e.result.ToStage.Name
	} else if len(e.candidates) == 1 {
		name = e.candidates[0].ToStage.Name
	}

	w := e.width
	if w < 40 {
		w = 40
	}

	title := e.theme.EvolveTitle.Width(w - 2).Render("✨ 进化中...")

	sparkle := e.theme.EvolveArt.Width(w - 4).Render(art)

	info := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Width(w - 4).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("%s → %s", e.oldStageID, name))

	// Progress bar
	progress := e.animTick
	if progress > 6 {
		progress = 6
	}
	filled := strings.Repeat("█", progress)
	empty := strings.Repeat("░", 6-progress)
	bar := lipgloss.NewStyle().Foreground(styles.GoldColor()).Render(filled) +
		lipgloss.NewStyle().Foreground(styles.DimColor()).Render(empty)

	return lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		"",
		sparkle,
		"",
		info,
		"",
		lipgloss.NewStyle().Width(w-4).Align(lipgloss.Center).Render(bar),
		"",
	)
}

func (e EvolveModel) viewDone() string {
	if e.result == nil {
		return "进化完成"
	}

	w := e.width
	if w < 40 {
		w = 40
	}

	title := e.theme.EvolveTitle.Width(w - 2).Render("🎉 进化完成！")

	info := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Render(fmt.Sprintf(
			"%s 进化为：\n\n  %s（%s）",
			e.pet.Name,
			e.result.ToStage.Name,
			e.result.ToStage.Phase,
		))

	help := e.theme.HelpBar.Render("Enter 继续")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		info,
		"",
		help,
	)
}
