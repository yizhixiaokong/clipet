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
	i18n       *i18n.Manager
	keyMap     keys.EvolveKeyMap
	help       help.Model

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
func NewEvolveModel(pet *game.Pet, candidates []game.EvolveCandidate, theme styles.Theme, i18nMgr *i18n.Manager) EvolveModel {
	phase := EvolveChoosing
	if len(candidates) == 1 {
		phase = EvolveAnimating
	}
	return EvolveModel{
		pet:        pet,
		candidates: candidates,
		theme:      theme,
		i18n:       i18nMgr,
		keyMap:     keys.NewEvolveKeyMap(i18nMgr),
		help:       help.New(),
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
			switch {
			case key.Matches(msg, e.keyMap.Navigation.Up):
				if e.choiceIdx > 0 {
					e.choiceIdx--
				}
			case key.Matches(msg, e.keyMap.Navigation.Down):
				if e.choiceIdx < len(e.candidates)-1 {
					e.choiceIdx++
				}
			case key.Matches(msg, e.keyMap.Navigation.Enter):
				e.result = &e.candidates[e.choiceIdx]
				e.phase = EvolveAnimating
				e.animTick = 0
			case key.Matches(msg, e.keyMap.Navigation.Back):
				e.done = true // 取消进化，返回主屏幕
			}

		case EvolveDone:
			if key.Matches(msg, e.keyMap.Navigation.Enter) {
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

	title := e.theme.EvolveTitle.Width(w - 2).Render("✨ " + e.i18n.T("game.evolution.evolving"))

	desc := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Render(e.i18n.T("ui.evolve.can_evolve", "name", e.pet.Name))

	var choices []string
	for i, c := range e.candidates {
		label := fmt.Sprintf("%s (%s)", c.ToStage.Name, c.ToStage.Phase)
		if i == e.choiceIdx {
			choices = append(choices, e.theme.ActionCellSelected.Width(w-6).Render("▸ "+label))
		} else {
			choices = append(choices, e.theme.ActionCell.Width(w-6).Render("  "+label))
		}
	}

	helpBar := e.theme.HelpBar.Render(e.help.View(e.keyMap))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		desc,
		"",
		lipgloss.JoinVertical(lipgloss.Left, choices...),
		"",
		helpBar,
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

	title := e.theme.EvolveTitle.Width(w - 2).Render("✨ " + e.i18n.T("game.evolution.evolving"))

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
		return e.i18n.T("game.evolution.evolution_complete", "oldStage", "", "newStage", "")
	}

	w := e.width
	if w < 40 {
		w = 40
	}

	title := e.theme.EvolveTitle.Width(w - 2).Render("🎉 " + e.i18n.T("game.evolution.evolution_complete", "oldStage", "", "newStage", ""))

	info := lipgloss.NewStyle().
		Foreground(styles.TextColor()).
		Render(e.i18n.T("ui.evolve.evolved_to",
			"name", e.pet.Name,
			"stage", e.result.ToStage.Name,
			"phase", e.result.ToStage.Phase))

	help := e.theme.HelpBar.Render("Enter " + e.i18n.T("ui.common.continue"))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		info,
		"",
		help,
	)
}
