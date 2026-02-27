// Package dev provides TUI models for clipet-dev commands
package dev

import (
	"clipet/internal/plugin"
	"clipet/internal/tui/components"
	"clipet/internal/tui/styles"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// PreviewKeyMap defines keybindings for preview command
type PreviewKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	SpeedUp   key.Binding
	SlowDown  key.Binding
	Quit      key.Binding
}

// DefaultPreviewKeyMap returns default keybindings for preview command
var DefaultPreviewKeyMap = PreviewKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "上移"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "下移"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "折叠"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "展开"),
	),
	SpeedUp: key.NewBinding(
		key.WithKeys("+", "="),
		key.WithHelp("+", "加速"),
	),
	SlowDown: key.NewBinding(
		key.WithKeys("-", "_"),
		key.WithHelp("-", "减速"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q/Esc", "退出"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k PreviewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k PreviewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.SpeedUp, k.SlowDown, k.Quit},
	}
}

// PreviewModel is the Bubble Tea model for interactive preview.
type PreviewModel struct {
	Pack     *plugin.SpeciesPack
	Tree     components.TreeList
	Fps      int
	FrameIdx int
	Width    int
	Height   int
	Quitting bool
	KeyMap   PreviewKeyMap
	Help     help.Model
}

// PreviewTickMsg drives animation.
type PreviewTickMsg time.Time

// NewPreviewModel creates a new preview TUI model
func NewPreviewModel(pack *plugin.SpeciesPack, fps int, initStage, initAnim string) *PreviewModel {
	h := help.New()
	h.ShowAll = false

	roots := buildPreviewTree(pack)
	if len(roots) == 0 {
		return &PreviewModel{Pack: pack, Fps: fps, KeyMap: DefaultPreviewKeyMap, Help: h}
	}

	// Create TreeList component
	tree := components.NewTreeList(roots)
	tree.ShowConnectors = false // Preview doesn't need connectors
	tree.ExpandToLevel(2)       // Expand phase and stage to show all animations
	tree.SetSize(40, 20)        // Initial size (will be updated)

	m := &PreviewModel{
		Pack:   pack,
		Tree:   tree,
		Fps:    fps,
		KeyMap: DefaultPreviewKeyMap,
		Help:   h,
	}

	// Set initial cursor position
	cursorSet := false
	if initStage != "" {
		// Try to find matching animation node by traversing the tree
		if initAnim != "" {
			// Try exact match: find animation with matching stageID and animState
			for _, root := range roots {
				for _, stage := range root.Children {
					if stage.ID == initStage {
						for _, anim := range stage.Children {
							if data, ok := anim.Data.(map[string]any); ok {
								if data["stageID"] == initStage && data["animState"] == initAnim {
									m.Tree.SetCursor(anim.ID)
									cursorSet = true
									break
								}
							}
						}
					}
					if cursorSet {
						break
					}
				}
				if cursorSet {
					break
				}
			}
		}

		if !cursorSet {
			// Fallback: try stageID.idle
			for _, root := range roots {
				for _, stage := range root.Children {
					if stage.ID == initStage {
						for _, anim := range stage.Children {
							if data, ok := anim.Data.(map[string]any); ok {
								if data["stageID"] == initStage && data["animState"] == "idle" {
									m.Tree.SetCursor(anim.ID)
									cursorSet = true
									break
								}
							}
						}
					}
					if cursorSet {
						break
					}
				}
				if cursorSet {
					break
				}
			}
		}

		if !cursorSet {
			// Final fallback: first animation of this stage
			for _, root := range roots {
				for _, stage := range root.Children {
					if stage.ID == initStage && len(stage.Children) > 0 {
						m.Tree.SetCursor(stage.Children[0].ID)
						cursorSet = true
						break
					}
				}
				if cursorSet {
					break
				}
			}
		}
	}
	// If still not set, NewTreeList already positioned cursor at first selectable

	return m
}

// Init implements tea.Model
func (m *PreviewModel) Init() tea.Cmd {
	return doPreviewTick(m.Fps)
}

// Update implements tea.Model
func (m *PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update tree size (right panel)
		rightW := m.Width * 45 / 100
		if rightW < 20 {
			rightW = 20
		}
		m.Tree.SetSize(rightW-2, m.Height-8)
		return m, nil

	case PreviewTickMsg:
		m.FrameIdx++
		return m, doPreviewTick(m.Fps)

	case tea.KeyPressMsg:
		// Handle global keys
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			m.Quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.SpeedUp):
			if m.Fps < 30 {
				m.Fps++
			}
			return m, nil
		case key.Matches(msg, m.KeyMap.SlowDown):
			if m.Fps > 1 {
				m.Fps--
			}
			return m, nil
		}

		// Delegate navigation to tree
		var cmd tea.Cmd
		m.Tree, cmd = m.Tree.Update(msg)
		return m, cmd

	case components.TreeSelectMsg:
		// User selected a different animation
		m.FrameIdx = 0 // Reset animation
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m *PreviewModel) View() tea.View {
	if m.Quitting {
		return tea.NewView("")
	}
	if m.Width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	// Layout: left preview (55%) │ right tree (45%)
	leftW := m.Width * 55 / 100
	rightW := m.Width - leftW - 3 // 3 for separator "│"
	if leftW < 30 {
		leftW = 30
	}
	if rightW < 20 {
		rightW = 20
	}

	leftPanel := m.renderPreview(leftW)
	rightPanel := m.renderTreePanel(rightW)
	helpBar := m.Help.View(m.KeyMap)

	sep := strings.Repeat("│\n", m.Height-4)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, sep, rightPanel)

	// Add help bar at the bottom
	fullContent := lipgloss.JoinVertical(lipgloss.Left, content, "", helpBar)

	v := tea.NewView(fullContent)
	v.AltScreen = true
	return v
}

// renderPreview renders the left panel with animated ASCII art.
func (m *PreviewModel) renderPreview(width int) string {
	panelTitle := previewTitleStyle.Render(" 帧预览 ")

	node := m.Tree.Selected()
	if node == nil || !node.Selectable || node.Data == nil {
		content := lipgloss.JoinVertical(lipgloss.Left,
			panelTitle,
			"",
			"请选择一个动画",
		)
		return previewPanelStyle.Width(width - 2).Render(content)
	}

	data, ok := node.Data.(map[string]any)
	if !ok {
		content := lipgloss.JoinVertical(lipgloss.Left,
			panelTitle,
			"",
			"数据格式错误",
		)
		return previewPanelStyle.Width(width - 2).Render(content)
	}

	frameKey, ok := data["frameKey"].(string)
	if !ok {
		content := lipgloss.JoinVertical(lipgloss.Left,
			panelTitle,
			"",
			"缺少 frameKey",
		)
		return previewPanelStyle.Width(width - 2).Render(content)
	}

	frame, exists := m.Pack.Frames[frameKey]
	if !exists {
		content := lipgloss.JoinVertical(lipgloss.Left,
			panelTitle,
			"",
			fmt.Sprintf("找不到帧: %s", frameKey),
		)
		return previewPanelStyle.Width(width - 2).Render(content)
	}

	var artStr string
	frameCount := 0
	if len(frame.Frames) == 0 {
		artStr = "  (无帧数据)"
	} else {
		frameCount = len(frame.Frames)
		idx := m.FrameIdx % frameCount
		raw := strings.TrimRight(frame.Frames[idx], "\n")
		artStr = components.NormalizeArt(raw, frame.Width)
	}

	frameInfo := previewInfoStyle.Render(fmt.Sprintf("▶ %s", frameKey))
	frameStats := previewInfoStyle.Render(fmt.Sprintf(
		"帧 %d/%d  %d fps",
		m.FrameIdx%max(frameCount, 1)+1, max(frameCount, 1), m.Fps,
	))

	content := lipgloss.JoinVertical(lipgloss.Left,
		panelTitle,
		"",
		frameInfo,
		"",
		artStr,
		"",
		frameStats,
	)

	return previewPanelStyle.
		Width(width - 2).
		Height(m.Height - 4).
		Render(content)
}

// renderTreePanel renders the right panel with tree list
func (m *PreviewModel) renderTreePanel(width int) string {
	title := previewTitleStyle.Render(" 帧列表 ")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		m.Tree.View(),
	)

	return previewPanelStyle.
		Width(width - 2).
		Height(m.Height - 4).
		Render(content)
}

// buildPreviewTree creates TreeNode tree from SpeciesPack
func buildPreviewTree(pack *plugin.SpeciesPack) []*components.TreeNode {
	stagePhase := make(map[string]string)
	stageName := make(map[string]string)
	for _, s := range pack.Stages {
		stagePhase[s.ID] = s.Phase
		stageName[s.ID] = s.Name
	}

	// Group frames by phase -> stage -> animations
	type animInfo struct {
		state    string
		frameKey string
		count    int
	}
	phaseStages := make(map[string]map[string][]animInfo)

	for key, frame := range pack.Frames {
		sid := frame.StageID
		anim := frame.AnimState
		phase := stagePhase[sid]
		if phase == "" {
			phase = "unknown"
		}
		if phaseStages[phase] == nil {
			phaseStages[phase] = make(map[string][]animInfo)
		}
		phaseStages[phase][sid] = append(phaseStages[phase][sid], animInfo{
			state:    anim,
			frameKey: key,
			count:    len(frame.Frames),
		})
	}

	// Sort phases
	phaseOrder := map[string]int{
		"egg": 0, "baby": 1, "child": 2, "adult": 3, "legend": 4, "unknown": 9,
	}
	phases := make([]string, 0, len(phaseStages))
	for p := range phaseStages {
		phases = append(phases, p)
	}
	sort.Slice(phases, func(i, j int) bool {
		return phaseOrder[phases[i]] < phaseOrder[phases[j]]
	})

	// Build TreeNode tree
	var roots []*components.TreeNode

	for _, phase := range phases {
		stages := phaseStages[phase]

		// Phase node (selectable for navigation, but not a real animation)
		phaseNode := &components.TreeNode{
			ID:         "phase:" + phase,
			Label:      fmt.Sprintf("── %s ──", strings.ToUpper(phase)),
			Selectable: true, // Allow cursor to stop on this node
			Expanded:   true,
			Children:   nil,
		}

		// Sort stage IDs
		stageIDs := make([]string, 0, len(stages))
		for sid := range stages {
			stageIDs = append(stageIDs, sid)
		}
		sort.Strings(stageIDs)

		// Build stage nodes
		for _, sid := range stageIDs {
			anims := stages[sid]

			sLabel := sid
			if name, ok := stageName[sid]; ok && name != "" {
				sLabel = fmt.Sprintf("%s [%s]", name, sid)
			}

			stageNode := &components.TreeNode{
				ID:         sid,
				Label:      sLabel,
				Selectable: true, // Allow cursor to stop on this node
				Expanded:   true,
				Children:   nil,
			}

			// Sort animations (idle first)
			sort.Slice(anims, func(i, j int) bool {
				if anims[i].state == "idle" {
					return true
				}
				if anims[j].state == "idle" {
					return false
				}
				return anims[i].state < anims[j].state
			})

			// Build animation nodes (leaves)
			for _, a := range anims {
				animNode := &components.TreeNode{
					ID:         a.frameKey,
					Label:      fmt.Sprintf("%s (%d帧)", a.state, a.count),
					Selectable: true,
					Expanded:   false,
					Data: map[string]any{
						"stageID":   sid,
						"animState": a.state,
						"frameKey":  a.frameKey,
						"count":     a.count,
					},
				}
				stageNode.Children = append(stageNode.Children, animNode)
			}

			phaseNode.Children = append(phaseNode.Children, stageNode)
		}

		roots = append(roots, phaseNode)
	}

	return roots
}

func doPreviewTick(fps int) tea.Cmd {
	d := time.Second / time.Duration(fps)
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return PreviewTickMsg(t)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Styles
var (
	previewPanelStyle = styles.DevCommandStyles.Panel
	previewTitleStyle = styles.DevCommandStyles.Title
	previewInfoStyle  = styles.DevCommandStyles.Info
)
