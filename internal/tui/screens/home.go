package screens

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"clipet/internal/tui/components"
	"clipet/internal/tui/styles"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// actionItem represents a single action in the 2x2 grid.
type actionItem struct {
	icon   string
	label  string
	action string
}

// 3x2 action grid (Pokemon-style).
var actionGrid = [3][2]actionItem{
	{{"🍖", "喂食", "feed"}, {"🎮", "玩耍", "play"}},
	{{"💤", "休息", "rest"}, {"💊", "治疗", "heal"}},
	{{"💬", "对话", "talk"}, {"📋", "信息", "info"}},
}

const actionRows = 3
const actionCols = 2

// HomeModel is the home screen model.
type HomeModel struct {
	pet      *game.Pet
	registry *plugin.Registry
	store    store.Store
	petView  *components.PetView
	theme    styles.Theme

	menuRow int
	menuCol int
	width   int
	height  int

	message   string // transient feedback message
	dialogue  string // last dialogue line
	msgIsInfo bool   // true if message is info-type
	msgIsWarn bool   // true if message is a warning (cooldown/prereq fail)
}

// NewHomeModel creates a new home screen model.
func NewHomeModel(
	pet *game.Pet,
	reg *plugin.Registry,
	st store.Store,
	pv *components.PetView,
	theme styles.Theme,
) HomeModel {
	return HomeModel{
		pet:      pet,
		registry: reg,
		store:    st,
		petView:  pv,
		theme:    theme,
	}
}

// SetSize updates the terminal dimensions.
func (h HomeModel) SetSize(w, ht int) HomeModel {
	h.width = w
	h.height = ht
	return h
}

// UpdatePet refreshes the pet reference.
func (h HomeModel) UpdatePet(pet *game.Pet) HomeModel {
	h.pet = pet
	h.petView.SetPet(pet)
	return h
}

// Update handles input for the home screen.
func (h HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if h.menuRow > 0 {
				h.menuRow--
			}
		case "down", "j":
			if h.menuRow < actionRows-1 {
				h.menuRow++
			}
		case "left", "h":
			if h.menuCol > 0 {
				h.menuCol--
			}
		case "right", "l":
			if h.menuCol < actionCols-1 {
				h.menuCol++
			}
		case "enter", " ":
			h = h.executeAction(actionGrid[h.menuRow][h.menuCol].action)
		case "f":
			h = h.executeAction("feed")
		case "p":
			h = h.executeAction("play")
		case "r":
			h = h.executeAction("rest")
		case "c":
			h = h.executeAction("heal")
		case "t":
			h = h.executeAction("talk")
		}
	}
	return h, nil
}

func (h HomeModel) executeAction(action string) HomeModel {
	switch action {
	case "feed":
		res := h.pet.Feed()
		if !res.OK {
			h.message = res.Message
			h.dialogue = ""
			h.msgIsInfo = false
			h.msgIsWarn = true
			return h
		}
		_ = h.store.Save(h.pet)
		ch := res.Changes["hunger"]
		h.message = fmt.Sprintf("喂食成功！饱腹度 %d → %d", ch[0], ch[1])
		h.dialogue = ""
		h.msgIsInfo = false
		h.msgIsWarn = false

	case "play":
		res := h.pet.Play()
		if !res.OK {
			h.message = res.Message
			h.dialogue = ""
			h.msgIsInfo = false
			h.msgIsWarn = true
			return h
		}
		_ = h.store.Save(h.pet)
		ch := res.Changes["happiness"]
		h.message = fmt.Sprintf("玩耍愉快！快乐度 %d → %d", ch[0], ch[1])
		h.dialogue = ""
		h.msgIsInfo = false
		h.msgIsWarn = false

	case "talk":
		res := h.pet.Talk()
		if !res.OK {
			h.message = res.Message
			h.dialogue = ""
			h.msgIsInfo = false
			h.msgIsWarn = true
			return h
		}
		line := h.registry.GetDialogue(h.pet.Species, h.pet.StageID, h.pet.MoodName())
		if line == "" {
			line = "......"
		}
		_ = h.store.Save(h.pet)
		h.dialogue = line
		h.message = ""
		h.msgIsInfo = false
		h.msgIsWarn = false

	case "rest":
		res := h.pet.Rest()
		if !res.OK {
			h.message = res.Message
			h.dialogue = ""
			h.msgIsInfo = false
			h.msgIsWarn = true
			return h
		}
		_ = h.store.Save(h.pet)
		chE := res.Changes["energy"]
		chH := res.Changes["health"]
		h.message = fmt.Sprintf("休息一下～精力 %d→%d  健康 %d→%d", chE[0], chE[1], chH[0], chH[1])
		h.dialogue = ""
		h.msgIsInfo = false
		h.msgIsWarn = false

	case "heal":
		res := h.pet.Heal()
		if !res.OK {
			h.message = res.Message
			h.dialogue = ""
			h.msgIsInfo = false
			h.msgIsWarn = true
			return h
		}
		_ = h.store.Save(h.pet)
		chH := res.Changes["health"]
		chE := res.Changes["energy"]
		h.message = fmt.Sprintf("治疗完成！健康 %d→%d  精力 %d→%d", chH[0], chH[1], chE[0], chE[1])
		h.dialogue = ""
		h.msgIsInfo = false
		h.msgIsWarn = false

	case "info":
		h.message = fmt.Sprintf(
			"互动 %d  喂食 %d  玩耍 %d  对话 %d  冒险 %d",
			h.pet.TotalInteractions,
			h.pet.FeedCount,
			h.pet.AccPlayful,
			h.pet.DialogueCount,
			h.pet.AdventuresCompleted,
		)
		h.dialogue = ""
		h.msgIsInfo = true
		h.msgIsWarn = false
	}
	return h
}

// ----- View rendering -----

func (h HomeModel) View() string {
	if h.width == 0 {
		return "正在加载..."
	}

	// Calculate panel widths — use fixed right panel, left gets remainder
	totalInner := h.width - 2
	if totalInner < 50 {
		totalInner = 50
	}
	// Right panel: label(6) + bar(10) + num(4) + padding/border(6) = ~26
	const rightPanelW = 30
	leftW := totalInner - rightPanelW
	if leftW < 20 {
		leftW = 20
	}
	rightW := totalInner - leftW

	// 1) Title bar
	title := h.theme.TitleBar.Width(totalInner).Render("🐾 Clipet")

	// 2) Main area: pet art (left) | status panel (right)
	petArt := h.renderPetPanel(leftW)
	statusPanel := h.renderStatusPanel(rightW)
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, petArt, statusPanel)

	// 3) Dialogue / message area
	msgArea := h.renderMessageArea(totalInner)

	// 4) Action menu (2x2 grid)
	actionMenu := h.renderActionGrid(totalInner)

	// 5) Help bar
	help := h.theme.HelpBar.Width(totalInner).Render("←→↑↓ 选择  Enter 确认  f喂食 p玩耍 r休息 c治疗 t对话  q 退出")

	// Compose
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		mainArea,
		msgArea,
		actionMenu,
		help,
	)
}

// renderPetPanel renders the left panel with centered ASCII art.
func (h HomeModel) renderPetPanel(width int) string {
	art := h.petView.Render()

	// Minimum height to keep layout stable
	const minHeight = 10
	lines := strings.Split(art, "\n")
	for len(lines) < minHeight {
		lines = append(lines, "")
	}

	// Find max line width for centering
	maxW := 0
	for _, l := range lines {
		if len(l) > maxW {
			maxW = len(l)
		}
	}

	// Center art within panel
	centered := strings.Join(lines, "\n")

	innerW := width - 6 // border + padding
	if innerW < maxW {
		innerW = maxW
	}

	return h.theme.PetPanel.
		Width(innerW).
		Height(minHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Render(centered)
}

// renderStatusPanel renders the right panel with pet info and stats.
func (h HomeModel) renderStatusPanel(width int) string {
	p := h.pet

	// Pet name
	name := h.theme.StatusName.Render(p.Name)

	// Stage info
	stageName := p.StageID
	if stage := h.registry.GetStage(p.Species, p.StageID); stage != nil {
		stageName = stage.Name
	}
	stageLine := h.theme.StatusLabel.Render("阶段") + " " +
		h.theme.StatusValue.Render(fmt.Sprintf("%s (%s)", stageName, p.Stage))

	// Mood
	moodStr, moodStyle := h.moodDisplay()
	moodLine := h.theme.StatusLabel.Render("心情") + " " + moodStyle.Render(moodStr)

	// Age
	ageLine := h.theme.StatusLabel.Render("年龄") + " " +
		h.theme.StatusValue.Render(fmt.Sprintf("%.1f 小时", p.AgeHours()))

	// Content width: label(6) + bar(10) + space+num(4) = 20
	const contentW = 20
	sep := lipgloss.NewStyle().
		Foreground(styles.DimColor()).
		Render(strings.Repeat("─", contentW))

	// Stats bars
	bars := []string{
		h.statBar("饱腹", p.Hunger),
		h.statBar("快乐", p.Happiness),
		h.statBar("健康", p.Health),
		h.statBar("精力", p.Energy),
	}
	statsBlock := strings.Join(bars, "\n")

	// Summary
	summary := lipgloss.NewStyle().Foreground(styles.DimColor()).Render(
		fmt.Sprintf("互动 %d", p.TotalInteractions))

	content := lipgloss.JoinVertical(lipgloss.Left,
		name,
		stageLine,
		moodLine,
		ageLine,
		sep,
		statsBlock,
		sep,
		summary,
	)

	const minHeight = 10
	innerW := width - 6
	if innerW < contentW {
		innerW = contentW
	}
	return h.theme.StatusPanel.
		Width(innerW).
		Height(minHeight).
		Render(content)
}

func (h HomeModel) moodDisplay() (string, lipgloss.Style) {
	mood := h.pet.MoodName()
	switch mood {
	case "happy":
		return "😊 开心", h.theme.MoodHappy
	case "normal":
		return "😐 普通", h.theme.MoodNormal
	case "unhappy":
		return "😕 不太好", h.theme.MoodSad
	case "sad":
		return "😢 难过", h.theme.MoodSad
	case "miserable":
		return "😭 非常差", h.theme.MoodMiserable
	default:
		return "❓ 未知", h.theme.MoodNormal
	}
}

func (h HomeModel) statBar(label string, value int) string {
	const barLen = 10
	filled := value / 10
	if filled > barLen {
		filled = barLen
	}
	empty := barLen - filled

	lab := h.theme.StatLabel.Render(label)
	fStr := h.theme.StatFilled.Render(strings.Repeat("━", filled))
	eStr := h.theme.StatEmpty.Render(strings.Repeat("─", empty))

	return fmt.Sprintf("%s%s%s %3d", lab, fStr, eStr, value)
}

// renderMessageArea renders the dialogue or action feedback.
func (h HomeModel) renderMessageArea(width int) string {
	innerW := width - 6
	if innerW < 10 {
		innerW = 10
	}

	if h.dialogue != "" {
		return h.theme.DialogueBox.Width(innerW).Render("💬 " + h.dialogue)
	}
	if h.message != "" {
		if h.msgIsWarn {
			return h.theme.MessageBox.Width(innerW).
				Copy().BorderForeground(lipgloss.Color("#AA5555")).
				Foreground(lipgloss.Color("#FF8888")).
				Render("⚠ " + h.message)
		}
		if h.msgIsInfo {
			return h.theme.MessageBox.Width(innerW).
				Copy().BorderForeground(lipgloss.Color("#555570")).
				Foreground(lipgloss.Color("#EAEAEA")).
				Render("📋 " + h.message)
		}
		return h.theme.MessageBox.Width(innerW).Render("✨ " + h.message)
	}

	// Empty placeholder to keep layout stable
	return h.theme.DialogueBox.Width(innerW).
		Copy().BorderForeground(styles.DimColor()).
		Foreground(styles.DimColor()).
		Render("  等待指令...")
}

// renderActionGrid renders the Pokemon-style 2x2 action menu.
func (h HomeModel) renderActionGrid(totalWidth int) string {
	cellInner := (totalWidth - 8) / 2 // account for borders
	if cellInner < 8 {
		cellInner = 8
	}

	renderCell := func(row, col int) string {
		item := actionGrid[row][col]
		label := item.icon + " " + item.label
		selected := h.menuRow == row && h.menuCol == col

		if selected {
			return h.theme.ActionCellSelected.Width(cellInner).Render("▸ " + label)
		}
		return h.theme.ActionCell.Width(cellInner).Render("  " + label)
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Center,
		renderCell(0, 0), renderCell(0, 1))
	midRow := lipgloss.JoinHorizontal(lipgloss.Center,
		renderCell(1, 0), renderCell(1, 1))
	botRow := lipgloss.JoinHorizontal(lipgloss.Center,
		renderCell(2, 0), renderCell(2, 1))

	return lipgloss.JoinVertical(lipgloss.Left, topRow, midRow, botRow)
}
