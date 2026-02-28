package screens

import (
	"clipet/internal/game"
	"clipet/internal/game/games"
	"clipet/internal/i18n"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"clipet/internal/tui/components"
	"clipet/internal/tui/keys"
	"clipet/internal/tui/styles"
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// actionItem represents a single action.
type actionItem struct {
	icon   string
	label  string
	action string
}

// menuCategory groups related actions under a tab.
type menuCategory struct {
	icon    string
	label   string
	actions []actionItem
}

// categories defines the two-level menu structure (untranslated keys).
var categories = []menuCategory{
	{"🐾", "care", []actionItem{
		{"🍖", "feed", "feed"},
		{"💤", "rest", "rest"},
		{"💊", "heal", "heal"},
	}},
	{"🎮", "interact", []actionItem{
		{"🎮", "play", "play"},
		{"💬", "talk", "talk"},
		{"🗺️", "adventure", "adventure"},
	}},
	{"🎯", "games", []actionItem{
		{"⚡", "game_reaction", "game_reaction"},
		{"🎲", "game_guess", "game_guess"},
	}},
	{"📋", "view", []actionItem{
		{"📋", "info", "info"},
		{"✨", "extra_attrs", "extra_attrs"},
	}},
}

// getTranslatedCategories returns categories with translated labels.
func (h HomeModel) getTranslatedCategories() []menuCategory {
	result := make([]menuCategory, len(categories))
	for i, cat := range categories {
		translatedCat := menuCategory{
			icon:  cat.icon,
			label: h.i18n.T("ui.home.categories." + cat.label),
		}
		translatedCat.actions = make([]actionItem, len(cat.actions))
		for j, action := range cat.actions {
			translatedCat.actions[j] = actionItem{
				icon:   action.icon,
				label:  h.i18n.T("ui.home.actions." + action.action),
				action: action.action,
			}
		}
		result[i] = translatedCat
	}
	return result
}

// HomeModel is the home screen model.
type HomeModel struct {
	pet      *game.Pet
	registry *plugin.Registry
	store    store.Store
	i18n     *i18n.Manager
	petView  *components.PetView
	theme    styles.Theme
	bubble   components.DialogueBubble
	gameMgr  *games.GameManager
	keyMap   keys.HomeKeyMap
	help     help.Model

	catIdx    int  // selected category tab
	actIdx    int  // selected action within category
	inSubmenu bool // true when navigating sub-actions
	width     int
	height    int

	message    string // transient feedback message
	msgIsInfo  bool   // true if message is info-type
	msgIsWarn  bool   // true if message is a warning
	lastTalkAt time.Time

	successMsg     string // success message with animation
	successAnimFrame int   // animation frame counter

	activeGame games.MiniGame // non-nil when a game is in progress

	pendingAdventure *plugin.Adventure // set when user triggers adventure
}

// NewHomeModel creates a new home screen model.
func NewHomeModel(
	pet *game.Pet,
	reg *plugin.Registry,
	st store.Store,
	pv *components.PetView,
	theme styles.Theme,
	i18nMgr *i18n.Manager,
) HomeModel {
	return HomeModel{
		pet:        pet,
		registry:   reg,
		store:      st,
		i18n:       i18nMgr,
		petView:    pv,
		bubble:     components.NewDialogueBubble(),
		gameMgr:    games.NewGameManager(),
		theme:      theme,
		lastTalkAt: time.Now(),
		keyMap:     keys.NewHomeKeyMap(i18nMgr),
		help:       help.New(),
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

// TickAutoDialogue handles periodic auto-dialogue (called from app tick).
func (h HomeModel) TickAutoDialogue() HomeModel {
	if h.activeGame != nil {
		return h
	}
	if time.Since(h.lastTalkAt) < 1*time.Minute {
		return h
	}
	if rand.Float32() >= 0.3 {
		// 失败，30秒后重试
		h.lastTalkAt = h.lastTalkAt.Add(30 * time.Second)
		return h
	}
	line := h.registry.GetDialogue(h.pet.Species, h.pet.StageID, h.pet.MoodName())
	if line != "" && line != "......" {
		h.bubble.UpdateText(line)
	}
	h.lastTalkAt = time.Now()
	return h
}

// TickSuccessAnimation advances the success animation frame.
func (h HomeModel) TickSuccessAnimation() HomeModel {
	if h.successMsg != "" {
		h.successAnimFrame++
		// Clear animation after 4 frames (about 2 seconds at 2fps)
		if h.successAnimFrame >= 4 {
			h.successMsg = ""
			h.successAnimFrame = 0
		}
	}
	return h
}

// TickGame advances the active mini-game by one tick.
func (h HomeModel) TickGame() HomeModel {
	if h.activeGame != nil {
		h.activeGame.Tick()
	}
	return h
}

// IsPlayingGame returns true if a mini-game is in progress.
func (h HomeModel) IsPlayingGame() bool {
	return h.activeGame != nil
}

// PendingAdventure returns the adventure to start, if any.
func (h HomeModel) PendingAdventure() *plugin.Adventure {
	return h.pendingAdventure
}

// ClearPendingAdventure clears the pending adventure request.
func (h HomeModel) ClearPendingAdventure() HomeModel {
	h.pendingAdventure = nil
	return h
}

// getCurrentActions returns the current category's actions, including dynamically added skills.
func (h HomeModel) getCurrentActions() []actionItem {
	translatedCats := h.getTranslatedCategories()
	cat := translatedCats[h.catIdx]
	actions := cat.actions

	// Dynamically add skill actions to the "interact" category (index 1)
	if h.catIdx == 1 && h.pet.CapabilitiesRegistry() != nil {
		skills := h.pet.CapabilitiesRegistry().GetActiveTraits(h.pet.Species)
		for _, skill := range skills {
			// Use localized skill name from plugin registry
			skillName := h.registry.GetTraitName(h.pet.Species, skill.ID)
			actions = append(actions, actionItem{
				icon:   "✨",
				label:  skillName,
				action: "skill:" + skill.ID,
			})
		}
	}

	return actions
}

// Update handles input for the home screen.
func (h HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	// If a game is active, delegate all input to the game
	if h.activeGame != nil {
		return h.updateGame(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Global shortcut keys always work
		switch {
		case key.Matches(msg, h.keyMap.Actions.Feed):
			return h.executeAction("feed"), nil
		case key.Matches(msg, h.keyMap.Actions.Play):
			return h.executeAction("play"), nil
		case key.Matches(msg, h.keyMap.Actions.Rest):
			return h.executeAction("rest"), nil
		case key.Matches(msg, h.keyMap.Actions.Heal):
			return h.executeAction("heal"), nil
		case key.Matches(msg, h.keyMap.Actions.Talk):
			return h.executeAction("talk"), nil
		}

		// "g" key for games (special case)
		if msg.String() == "g" {
			if h.inSubmenu && h.catIdx == 2 { // Games category
				translatedCats := h.getTranslatedCategories()
				act := translatedCats[h.catIdx].actions[h.actIdx]
				return h.executeAction(act.action), nil
			}
			return h, nil
		}

		if !h.inSubmenu {
			// Level 0: category tabs
			switch {
			case key.Matches(msg, h.keyMap.Navigation.Left):
				if h.catIdx > 0 {
					h.catIdx--
				}
			case key.Matches(msg, h.keyMap.Navigation.Right):
				if h.catIdx < len(categories)-1 {
					h.catIdx++
				}
			case key.Matches(msg, h.keyMap.Navigation.Down), key.Matches(msg, h.keyMap.Navigation.Enter):
				h.inSubmenu = true
				h.actIdx = 0
			}
		} else {
			// Level 1: sub-actions
			actions := h.getCurrentActions()
			switch {
			case key.Matches(msg, h.keyMap.Navigation.Left):
				if h.actIdx > 0 {
					h.actIdx--
				}
			case key.Matches(msg, h.keyMap.Navigation.Right):
				if h.actIdx < len(actions)-1 {
					h.actIdx++
				}
			case key.Matches(msg, h.keyMap.Navigation.Back):
				h.inSubmenu = false
			case key.Matches(msg, h.keyMap.Navigation.Enter):
				return h.executeAction(actions[h.actIdx].action), nil
			}
		}
	}
	return h, nil
}

// updateGame handles input while a game is active.
func (h HomeModel) updateGame(msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Esc cancels an unfinished game
		if key.Matches(msg, h.keyMap.Navigation.Back) && !h.activeGame.IsFinished() {
			h.activeGame = nil
			h.message = h.i18n.T("ui.home.game_cancelled")
			h.msgIsWarn = false
			h.msgIsInfo = false
			return h, nil
		}

		h.activeGame.HandleKey(msg.String())

		// If game is finished and confirmed, process result
		if h.activeGame.IsConfirmed() {
			h = h.processGameResult()
			h.activeGame = nil
		}
	}
	return h, nil
}

// ----- Action dispatch -----

// failMsg sets a warning message.
func (h HomeModel) failMsg(msg string) HomeModel {
	h.message = msg
	h.msgIsInfo = false
	h.msgIsWarn = true
	return h
}

// localizeGameError converts a game.ActionResult with ErrorType to localized message.
// Falls back to res.Message if ErrorType is empty or unknown.
func (h HomeModel) localizeGameError(res game.ActionResult) string {
	if res.ErrorType == "" {
		// No ErrorType, use original message (backward compatibility)
		return res.Message
	}

	// Map ErrorType to i18n key
	var i18nKey string
	switch res.ErrorType {
	case game.ErrEnergyLow:
		i18nKey = "game.errors.energy_low"
	case game.ErrHealthLow:
		i18nKey = "game.errors.health_low"
	case game.ErrCooldown:
		i18nKey = "game.errors.cooldown"
	case game.ErrDead:
		i18nKey = "game.errors.dead"
	case game.ErrInvalidAction:
		i18nKey = "game.errors.invalid_action"
	case game.ErrFullHunger:
		i18nKey = "game.errors.full_hunger"
	case game.ErrFullEnergy:
		i18nKey = "game.errors.full_energy"
	case game.ErrSkillSystem:
		i18nKey = "game.errors.skill_system"
	case game.ErrSkillUnknown:
		i18nKey = "game.errors.skill_unknown"
	case game.ErrSkillNotActive:
		i18nKey = "game.errors.skill_not_active"
	default:
		// Unknown ErrorType, fallback to Message
		return res.Message
	}

	return h.i18n.T(i18nKey)
}

// localizeAdventureError converts a game.AdventureCheckResult with ErrorType to localized message.
func (h HomeModel) localizeAdventureError(check game.AdventureCheckResult) string {
	if check.ErrorType == "" {
		return check.Message
	}

	var i18nKey string
	switch check.ErrorType {
	case game.ErrEnergyLow:
		i18nKey = "game.errors.energy_low"
	case game.ErrDead:
		i18nKey = "game.errors.dead"
	default:
		return check.Message
	}

	return h.i18n.T(i18nKey)
}

// okMsg sets a success message with animation and saves pet state.
func (h HomeModel) okMsg(msg string) HomeModel {
	h.successMsg = msg
	h.successAnimFrame = 0
	h.message = "" // Clear normal message
	h.msgIsInfo = false
	h.msgIsWarn = false
	if err := h.store.Save(h.pet); err != nil {
		h.successMsg = msg + " " + h.i18n.T("ui.home.save_failed")
	}
	return h
}

// applyActionResult applies animation and shows message from ActionResult.
func (h HomeModel) applyActionResult(res game.ActionResult, msg string) HomeModel {
	// Apply animation if specified
	if res.Animation != "" && res.AnimationDuration > 0 {
		h.pet.CurrentAnimation = res.Animation
		h.pet.AnimationEndTime = time.Now().Add(res.AnimationDuration)
	}
	return h.okMsg(msg)
}

// infoMsg sets an informational message.
func (h HomeModel) infoMsg(msg string) HomeModel {
	h.message = msg
	h.msgIsInfo = true
	h.msgIsWarn = false
	return h
}

func (h HomeModel) executeAction(action string) HomeModel {
	switch action {
	case "feed":
		res := h.pet.Feed()
		if !res.OK {
			return h.failMsg(h.localizeGameError(res))
		}
		ch := res.Changes["hunger"]
		return h.applyActionResult(res, h.i18n.T("ui.home.feed_success", "oldHunger", ch[0], "newHunger", ch[1]))

	case "play":
		res := h.pet.Play()
		if !res.OK {
			return h.failMsg(h.localizeGameError(res))
		}
		ch := res.Changes["happiness"]
		return h.applyActionResult(res, h.i18n.T("ui.home.play_success", "oldHappiness", ch[0], "newHappiness", ch[1]))

	case "talk":
		res := h.pet.Talk()
		if !res.OK {
			return h.failMsg(h.localizeGameError(res))
		}
		line := h.registry.GetDialogue(h.pet.Species, h.pet.StageID, h.pet.MoodName())
		if line == "" {
			line = "......"
		}
		h.bubble.UpdateText(line)
		h.lastTalkAt = time.Now()
		return h.okMsg(h.i18n.T("ui.home.talk_success"))

	case "rest":
		res := h.pet.Rest()
		if !res.OK {
			return h.failMsg(h.localizeGameError(res))
		}
		chE := res.Changes["energy"]
		chH := res.Changes["health"]
		return h.applyActionResult(res, h.i18n.T("ui.home.rest_success", "oldEnergy", chE[0], "newEnergy", chE[1], "oldHealth", chH[0], "newHealth", chH[1]))

	case "heal":
		res := h.pet.Heal()
		if !res.OK {
			return h.failMsg(h.localizeGameError(res))
		}
		chH := res.Changes["health"]
		chE := res.Changes["energy"]
		return h.okMsg(h.i18n.T("ui.home.heal_success", "oldHealth", chH[0], "newHealth", chH[1], "oldEnergy", chE[0], "newEnergy", chE[1]))

	case "info":
		return h.infoMsg(h.i18n.T("ui.home.stats_interactions",
			"interact", h.pet.TotalInteractions,
			"feed", h.pet.FeedCount,
			"play", h.pet.AccPlayful,
			"talk", h.pet.DialogueCount,
			"adventure", h.pet.AdventuresCompleted,
		))

	case "extra_attrs":
		if len(h.pet.CustomAttributes) == 0 {
			return h.infoMsg(h.i18n.T("ui.home.no_extra_attrs"))
		}
		var lines []string
		lines = append(lines, h.i18n.T("ui.home.extra_attrs_title"))
		for attrName, value := range h.pet.CustomAttributes {
			lines = append(lines, fmt.Sprintf("  %s: %d", attrName, value))
		}
		return h.infoMsg(strings.Join(lines, "\n"))

	case "game_reaction":
		return h.startGame(games.GameReactionSpeed)

	case "game_guess":
		return h.startGame(games.GameGuessNumber)

	case "adventure":
		check := game.CanAdventure(h.pet)
		if !check.OK {
			return h.failMsg(h.localizeAdventureError(check))
		}
		if time.Since(h.pet.LastAdventureAt) < game.CooldownAdventure {
			remain := game.CooldownAdventure - time.Since(h.pet.LastAdventureAt)
			return h.failMsg(h.i18n.T("ui.home.adventure_cooldown", "minutes", int(remain.Minutes())+1))
		}
		adv := game.PickAdventure(h.pet, h.registry)
		if adv == nil {
			return h.failMsg(h.i18n.T("ui.home.no_adventures"))
		}
		h.pendingAdventure = adv
		return h

	default:
		// Handle skill actions (format: "skill:skill_id")
		if strings.HasPrefix(action, "skill:") {
			skillID := strings.TrimPrefix(action, "skill:")
			res := h.pet.UseSkill(skillID)
			if !res.OK {
				return h.failMsg(h.localizeGameError(res))
			}
			chH := res.Changes["health"]
			chE := res.Changes["energy"]
			// Format: "{message}  Health X→Y  Energy A→B"
			detailMsg := fmt.Sprintf("%s  %s %d→%d  %s %d→%d",
				res.Message,
				h.i18n.T("game.stats.health"), chH[0], chH[1],
				h.i18n.T("game.stats.energy"), chE[0], chE[1])
			return h.applyActionResult(res, detailMsg)
		}
		return h
	}
}

// startGame initiates a mini-game session.
func (h HomeModel) startGame(gt games.GameType) HomeModel {
	config, ok := h.gameMgr.GetConfig(gt)
	if !ok {
		return h.failMsg(h.i18n.T("ui.home.game_unavailable"))
	}
	if h.pet.Energy < config.MinEnergy {
		return h.failMsg(h.i18n.T("game.errors.not_enough_energy"))
	}

	// Deduct energy upfront
	h.pet.Energy = game.Clamp(h.pet.Energy-config.EnergyCost, 0, 100)

	g := h.gameMgr.NewGame(gt)
	g.Start()
	h.activeGame = g
	h.message = ""
	return h
}

// processGameResult applies the game outcome to pet attributes.
func (h HomeModel) processGameResult() HomeModel {
	result := h.activeGame.GetResult()
	config := h.activeGame.GetConfig()

	if result.Won {
		h.pet.Happiness = game.Clamp(h.pet.Happiness+config.WinHappiness, 0, 100)
		h.pet.GamesWon++
		h.message = h.i18n.T("ui.home.game_won", "message", result.Message, "happiness", config.WinHappiness)
	} else {
		h.pet.Happiness = game.Clamp(h.pet.Happiness+config.LoseHappiness, 0, 100)
		h.message = h.i18n.T("ui.home.game_lost", "message", result.Message, "happiness", config.LoseHappiness)
	}
	h.pet.TotalInteractions++
	h.msgIsWarn = false
	h.msgIsInfo = false

	if err := h.store.Save(h.pet); err != nil {
		h.message += " " + h.i18n.T("ui.home.save_failed")
	}
	return h
}

// ----- View rendering -----

func (h HomeModel) View() string {
	if h.width == 0 {
		return h.i18n.T("ui.common.loading")
	}

	// If playing a game, render game overlay
	if h.activeGame != nil {
		return h.renderGameView()
	}

	totalInner := h.width - 2
	if totalInner < 50 {
		totalInner = 50
	}

	// Right panel shows status (name, stats, etc.)
	// 宠物显示占宽度的一半左右
	rightPanelW := totalInner / 2
	if rightPanelW < 30 {
		rightPanelW = 30
	}
	leftW := totalInner - rightPanelW
	if leftW < 20 {
		leftW = 20
	}
	rightW := totalInner - leftW

	// 1) Title bar - show ending message if pet has died
	var title string
	if h.hasEnding() {
		// Special title bar for ending with red/gold background
		title = h.theme.TitleBar.
			Width(totalInner).
			Background(lipgloss.Color("#AA3355")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render("💔 " + h.getLocalizedEndingMessage())
	} else {
		title = h.theme.TitleBar.Width(totalInner).Render("🐾 Clipet")
	}

	// 2) Main area: pet art (left) | status panel (right)
	petArt := h.renderPetPanel(leftW)
	statusPanel := h.renderStatusPanel(rightW)
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, petArt, statusPanel)

	// 3) Message area
	msgArea := h.renderMessageArea(totalInner)

	// 4) Action menu (category tabs + sub-actions)
	// Hide menu if pet has died
	var actionMenu string
	if h.hasEnding() {
		actionMenu = ""
	} else {
		actionMenu = h.renderActionMenu(totalInner)
	}

	// 5) Help bar
	var helpText string
	if h.hasEnding() {
		helpText = "q " + h.i18n.T("ui.common.quit")
	} else if h.inSubmenu {
		helpText = h.help.View(h.keyMap)
	} else {
		helpText = h.help.View(h.keyMap)
	}
	helpBar := h.theme.HelpBar.Width(totalInner).Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		mainArea,
		msgArea,
		actionMenu,
		helpBar,
	)
}

// renderGameView renders a split-screen view with pet on left and game on right.
func (h HomeModel) renderGameView() string {
	totalInner := h.width - 2
	if totalInner < 50 {
		totalInner = 50
	}

	// Split into left (pet) and right (game) panels
	// 宠物显示占宽度的一半左右
	leftPanelW := totalInner / 2
	if leftPanelW < 20 {
		leftPanelW = 20
	}
	rightPanelW := totalInner - leftPanelW
	if rightPanelW < 30 {
		rightPanelW = 30
	}

	// Left panel: Pet view (same as normal home view)
	leftPanel := h.renderPetPanel(leftPanelW)

	// Right panel: Game content
	var title string
	if h.hasEnding() {
		// Show ending message instead of game title
		title = h.theme.TitleBar.
			Width(rightPanelW).
			Background(lipgloss.Color("#AA3355")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render("💔 " + h.getLocalizedEndingMessage())
	} else {
		title = h.theme.TitleBar.Width(rightPanelW).Render("🎮 " + h.activeGame.GetConfig().Name)
	}

	gameContent := h.activeGame.View()
	gameBox := h.theme.GamePanel.
		Width(rightPanelW - 4).
		Render(gameContent)

	var helpText string
	if h.hasEnding() {
		helpText = "q " + h.i18n.T("ui.common.quit")
	} else if h.activeGame.IsFinished() {
		helpText = "Enter " + h.i18n.T("ui.common.continue") + "  Esc " + h.i18n.T("ui.common.quit")
	} else {
		helpText = h.i18n.T("ui.home.help_game")
	}
	helpBar := h.theme.HelpBar.Width(rightPanelW).Render(helpText)

	rightPanel := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		gameBox,
		"",
		helpBar,
	)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

// renderPetPanel renders the left panel with centered ASCII art + dialogue bubble.
func (h HomeModel) renderPetPanel(width int) string {
	// art is already a normalized rectangular block (all lines same width)
	art := h.petView.Render()

	const minHeight = 12
	lines := strings.Split(art, "\n")
	for len(lines) < minHeight {
		lines = append(lines, "")
	}
	art = strings.Join(lines, "\n")

	innerW := width - 6
	if innerW < 20 {
		innerW = 20
	}

	// Add dialogue bubble above pet if active
	bubbleText := h.bubble.Render()
	if bubbleText != "" {
		art = bubbleText + "\n\n" + art
	}

	return h.theme.PetPanel.
		Width(innerW).
		Height(minHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Render(art)
}

// renderStatusPanel renders the right panel with pet info and stats.
func (h HomeModel) renderStatusPanel(width int) string {
	p := h.pet

	name := h.theme.StatusName.Render(p.Name)

	stageName := p.StageID
	if stage := h.registry.GetStage(p.Species, p.StageID); stage != nil {
		stageName = stage.Name
	}
	stageLine := h.theme.StatusLabel.Render(h.i18n.T("game.stats.stage")) + " " +
		h.theme.StatusValue.Render(fmt.Sprintf("%s (%s)", stageName, p.Stage))

	moodStr, moodStyle := h.moodDisplay()
	moodLine := h.theme.StatusLabel.Render(h.i18n.T("game.stats.mood")) + " " + moodStyle.Render(moodStr)

	ageLine := h.theme.StatusLabel.Render(h.i18n.T("game.stats.age")) + " " +
		h.theme.StatusValue.Render(h.i18n.T("game.pet.age_hours", "hours", fmt.Sprintf("%.1f", p.AgeHours())))

	const contentW = 20
	sep := lipgloss.NewStyle().
		Foreground(styles.DimColor()).
		Render(strings.Repeat("-", contentW))

	bars := []string{
		h.statBar("🍖", h.i18n.T("game.stats.hunger"), p.Hunger),
		h.statBar("😺", h.i18n.T("game.stats.happiness"), p.Happiness),
		h.statBar("💊", h.i18n.T("game.stats.health"), p.Health),
		h.statBar("💤", h.i18n.T("game.stats.energy"), p.Energy),
	}
	statsBlock := strings.Join(bars, "\n")

	// Add more statistics
	stats := fmt.Sprintf("🗣 %s %d  🗺 %s %d",
		h.i18n.T("game.stats.dialogue"), p.DialogueCount,
		h.i18n.T("game.stats.adventure"), p.AdventuresCompleted)

	content := lipgloss.JoinVertical(lipgloss.Left,
		name,
		stageLine,
		moodLine,
		ageLine,
		sep,
		statsBlock,
		sep,
		stats,
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
		return h.i18n.T("game.mood.happy"), h.theme.MoodHappy
	case "normal":
		return h.i18n.T("game.mood.normal"), h.theme.MoodNormal
	case "unhappy":
		return h.i18n.T("game.mood.unhappy"), h.theme.MoodSad
	case "sad":
		return h.i18n.T("game.mood.sad"), h.theme.MoodSad
	case "miserable":
		return h.i18n.T("game.mood.miserable"), h.theme.MoodMiserable
	default:
		return h.i18n.T("game.mood.unknown"), h.theme.MoodNormal
	}
}

// isPetNearEndOfLife checks if pet is actually near end of life
func (h HomeModel) isPetNearEndOfLife() bool {
	// Get species lifecycle config
	species := h.registry.GetSpecies(h.pet.Species)
	if species == nil {
		return false
	}

	maxAge := species.Lifecycle.MaxAgeHours
	if maxAge <= 0 {
		return false // Eternal or invalid
	}

	// Calculate age percentage
	agePercent := h.pet.AgeHours() / maxAge
	return agePercent >= species.Lifecycle.WarningThreshold
}

// getActionCooldown returns the cooldown remaining for an action (empty string if no cooldown).
func (h HomeModel) getActionCooldown(action string) string {
	p := h.pet
	var cooldown time.Duration

	switch action {
	case "feed":
		cooldown = game.CalculateDynamicCooldown(p.Registry(), p.Species, "feed", p.Hunger)
		return cooldownLeft(p.LastFedAt, cooldown)
	case "play":
		cooldown = game.CalculateDynamicCooldown(p.Registry(), p.Species, "play", p.Happiness)
		return cooldownLeft(p.LastPlayedAt, cooldown)
	case "rest":
		// Low energy = urgent (short cooldown)
		cooldown = game.CalculateDynamicCooldown(p.Registry(), p.Species, "rest", p.Energy)
		return cooldownLeft(p.LastRestedAt, cooldown)
	case "heal":
		// Low health = urgent (short cooldown)
		cooldown = game.CalculateDynamicCooldown(p.Registry(), p.Species, "heal", p.Health)
		return cooldownLeft(p.LastHealedAt, cooldown)
	case "talk":
		cooldown = game.CalculateDynamicCooldown(p.Registry(), p.Species, "talk", p.Happiness)
		return cooldownLeft(p.LastTalkedAt, cooldown)
	case "adventure":
		// Adventure uses fixed cooldown (not dynamic)
		return cooldownLeft(p.LastAdventureAt, game.CooldownAdventure)
	default:
		// Handle skill actions (format: "skill:skill_id")
		if strings.HasPrefix(action, "skill:") {
			skillID := strings.TrimPrefix(action, "skill:")
			return h.getSkillCooldown(skillID)
		}
		return ""
	}
}

// cooldownLeft returns remaining cooldown time as a string.
func cooldownLeft(last time.Time, cd time.Duration) string {
	remaining := cd - time.Since(last)
	if remaining <= 0 {
		return ""
	}
	if remaining < time.Minute {
		return fmt.Sprintf("%ds", int(remaining.Seconds()))
	}
	return fmt.Sprintf("%dm", int(remaining.Minutes()))
}

// getSkillCooldown returns the cooldown remaining for a skill (empty string if no cooldown).
func (h HomeModel) getSkillCooldown(skillID string) string {
	capReg := h.pet.CapabilitiesRegistry()
	if capReg == nil {
		return ""
	}

	trait, exists := capReg.GetTrait(h.pet.Species, skillID)
	if !exists || trait.Type != "active" || trait.ActiveEffect == nil {
		return ""
	}

	// Cooldown is already parsed as time.Duration
	cooldown := trait.ActiveEffect.Cooldown
	if cooldown == 0 {
		cooldown = 30 * time.Minute // fallback
	}

	return cooldownLeft(h.pet.LastSkillUsedAt, cooldown)
}

func (h HomeModel) statBar(icon, label string, value int) string {
	const barLen = 10
	filled := value / 10
	if filled > barLen {
		filled = barLen
	}
	empty := barLen - filled

	lab := h.theme.StatLabel.Render(icon + " " + label)
	fStr := h.theme.StatFilled.Render(strings.Repeat(" ", filled))
	eStr := h.theme.StatEmpty.Render(strings.Repeat(" ", empty))

	return fmt.Sprintf("%s%s%s %3d", lab, fStr, eStr, value)
}

// getLocalizedEndingMessage returns the localized ending message.
// Priority: 1) Plugin locale, 2) Core i18n, 3) Plugin-provided message
func (h HomeModel) getLocalizedEndingMessage() string {
	if h.pet.EndingType == "" {
		return ""
	}

	// 1. Try plugin locale first (if plugin provided custom ending)
	if msg := h.registry.GetEndingMessage(h.pet.Species, h.pet.EndingType); msg != "" {
		return msg
	}

	// 2. Try core i18n (for default endings: peaceful_rest, blissful_passing, heroic_tale)
	if msg := h.i18n.T("game.endings." + h.pet.EndingType); msg != "" {
		// Check if i18n returned the key itself (meaning not found)
		if !strings.HasPrefix(msg, "game.endings.") {
			return msg
		}
	}

	// 3. Fallback to plugin-provided message (already localized by plugin)
	if h.pet.EndingMessage != "" {
		return h.pet.EndingMessage
	}

	// 4. Ultimate fallback
	return h.i18n.T("game.endings.peaceful_rest")
}

// hasEnding returns true if the pet has an ending (either type or message).
func (h HomeModel) hasEnding() bool {
	return !h.pet.Alive && (h.pet.EndingType != "" || h.pet.EndingMessage != "")
}

// renderMessageArea renders the action feedback area.
func (h HomeModel) renderMessageArea(width int) string {
	innerW := width - 6
	if innerW < 10 {
		innerW = 10
	}

	// Don't show messages if pet has died (ending is shown in title bar)
	if h.hasEnding() {
		return ""
	}

	// Render success animation if active
	if h.successMsg != "" {
		// Create blinking effect
		check := "✓"
		if h.successAnimFrame%2 == 0 {
			check = "✓" // Bright
		} else {
			check = "✓" // Still bright, but we'll change color
		}

		style := h.theme.MessageBox.Width(innerW).
			BorderForeground(lipgloss.Color("#04B575")).
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

		return style.Render(check + " " + h.successMsg)
	}

	// Show regular messages (warnings, info, normal)
	if h.message != "" {
		if h.msgIsWarn {
			return h.theme.MessageBox.Width(innerW).
				BorderForeground(lipgloss.Color("#AA5555")).
				Foreground(lipgloss.Color("#FF8888")).
				Render("⚠ " + h.message)
		}
		if h.msgIsInfo {
			return h.theme.MessageBox.Width(innerW).
				BorderForeground(lipgloss.Color("#555570")).
				Foreground(lipgloss.Color("#EAEAEA")).
				Render("📋 " + h.message)
		}
		return h.theme.MessageBox.Width(innerW).Render("✨ " + h.message)
	}

	// Show lifecycle warning only when no other messages
	// Check dynamically if pet is actually near end of life
	if h.pet.LifecycleWarningShown && h.isPetNearEndOfLife() {
		return h.theme.MessageBox.Width(innerW).
			BorderForeground(lipgloss.Color("#AA5555")).
			Foreground(lipgloss.Color("#FF8888")).
			Render(h.i18n.T("ui.home.lifecycle_warning"))
	}

	// Empty placeholder
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.DimColor()).
		Foreground(styles.DimColor()).
		Width(innerW).
		Padding(0, 2).
		Render(h.i18n.T("ui.home.waiting"))
}

// renderActionMenu renders the two-level category tabs + sub-action menu.
func (h HomeModel) renderActionMenu(totalWidth int) string {
	translatedCats := h.getTranslatedCategories()
	tabW := (totalWidth - 4) / len(translatedCats)
	if tabW < 8 {
		tabW = 8
	}
	var tabs []string
	for i, cat := range translatedCats {
		label := cat.icon + " " + cat.label
		if i == h.catIdx && !h.inSubmenu {
			tabs = append(tabs, h.theme.CategoryTabActive.Width(tabW).Render("▸ "+label))
		} else if i == h.catIdx && h.inSubmenu {
			tabs = append(tabs, h.theme.CategoryTabOpen.Width(tabW).Render("▾ "+label))
		} else {
			tabs = append(tabs, h.theme.CategoryTab.Width(tabW).Render("  "+label))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Center, tabs...)

	actions := h.getCurrentActions()

	actW := (totalWidth - 4) / len(actions)
	if actW < 8 {
		actW = 8
	}
	var acts []string
	for i, act := range actions {
		// Check if action is on cooldown
		cooldown := h.getActionCooldown(act.action)
		label := act.icon + " " + act.label

		if cooldown != "" {
			// Show grayed out with countdown
			label = act.icon + " " + act.label + " (" + cooldown + ")"
			if h.inSubmenu && i == h.actIdx {
				acts = append(acts, h.theme.ActionCellSelected.Width(actW).
					Foreground(lipgloss.Color("#555570")).
					Render("▸ "+label))
			} else {
				acts = append(acts, h.theme.ActionCell.Width(actW).
					Foreground(lipgloss.Color("#555570")).
					Render("  "+label))
			}
		} else {
			// Normal display
			if h.inSubmenu && i == h.actIdx {
				acts = append(acts, h.theme.ActionCellSelected.Width(actW).Render("▸ "+label))
			} else {
				acts = append(acts, h.theme.ActionCell.Width(actW).Render("  "+label))
			}
		}
	}
	actRow := lipgloss.JoinHorizontal(lipgloss.Center, acts...)

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, actRow)
}
