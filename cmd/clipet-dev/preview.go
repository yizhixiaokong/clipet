package main

import (
	"clipet/internal/plugin"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func newPreviewCmd() *cobra.Command {
	var fps int

	cmd := &cobra.Command{
		Use:   "preview <pack-dir> [stage-id] [anim-state]",
		Short: "[开发] 预览 ASCII 动画帧",
		Long: `交互式 TUI 预览物种包的 ASCII 动画帧。

左侧动画预览，右侧树形列表选择。
可选参数用于定位初始位置：
  preview <dir>                    — 从第一个动画开始
  preview <dir> <stage-id>         — 定位到该阶段的 idle
  preview <dir> <stage-id> <anim>  — 定位到指定动画

操作: ↑↓/jk 选择  +/- 调速  q/Esc 退出`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]

			info, err := os.Stat(dir)
			if err != nil {
				return fmt.Errorf("无法访问 %q: %w", dir, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("%q 不是目录", dir)
			}

			pack, err := plugin.ParsePack(os.DirFS(dir), ".")
			if err != nil {
				return fmt.Errorf("解析物种包失败: %w", err)
			}

			var initStage, initAnim string
			if len(args) >= 2 {
				initStage = args[1]
			}
			if len(args) >= 3 {
				initAnim = args[2]
			}

			return runPreviewTUI(pack, fps, initStage, initAnim)
		},
	}

	cmd.Flags().IntVar(&fps, "fps", 2, "帧率 (每秒帧数)")

	return cmd
}

// ---------- Interactive TUI ----------

// treeEntry represents a selectable item in the tree list.
type treeEntry struct {
	stageID   string
	animState string
	label     string // display text
	indent    int    // nesting level (0=phase header, 1=stage, 2=anim)
	isLeaf    bool   // only leaves (anim states) can be previewed
	phase     string
	frameKey  string // key into pack.Frames (only for leaves)
	children  int    // frame count (only for leaves)
}

// previewModel is the Bubble Tea model for interactive preview.
type previewModel struct {
	pack *plugin.SpeciesPack
	tree []treeEntry
	fps  int

	cursor   int // index in tree (always on a leaf)
	frameIdx int // animation frame counter
	width    int
	height   int
	quitting bool
}

// previewTickMsg drives animation.
type previewTickMsg time.Time

func doPreviewTick(fps int) tea.Cmd {
	d := time.Second / time.Duration(fps)
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return previewTickMsg(t)
	})
}

func runPreviewTUI(pack *plugin.SpeciesPack, fps int, initStage, initAnim string) error {
	tree := buildTree(pack)
	if len(tree) == 0 {
		return fmt.Errorf("物种包中没有帧数据")
	}

	m := previewModel{
		pack: pack,
		tree: tree,
		fps:  fps,
	}

	// Locate initial cursor position based on args
	m.cursor = findInitialCursor(tree, initStage, initAnim)

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// findInitialCursor returns the tree index matching the given stage/anim.
// Falls back: exact match → stage's idle → stage's first anim → first leaf.
func findInitialCursor(tree []treeEntry, stageID, animState string) int {
	if stageID == "" {
		// No args: first leaf
		for i, e := range tree {
			if e.isLeaf {
				return i
			}
		}
		return 0
	}

	// Try exact match first
	if animState != "" {
		for i, e := range tree {
			if e.isLeaf && e.stageID == stageID && e.animState == animState {
				return i
			}
		}
	}

	// Fallback: stage's idle
	for i, e := range tree {
		if e.isLeaf && e.stageID == stageID && e.animState == "idle" {
			return i
		}
	}

	// Fallback: stage's first anim
	for i, e := range tree {
		if e.isLeaf && e.stageID == stageID {
			return i
		}
	}

	// Fallback: first leaf
	for i, e := range tree {
		if e.isLeaf {
			return i
		}
	}
	return 0
}

func (m previewModel) Init() tea.Cmd {
	return doPreviewTick(m.fps)
}

func (m previewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case previewTickMsg:
		m.frameIdx++
		return m, doPreviewTick(m.fps)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "escape":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "home":
			m.cursor = m.findLeaf(true)
			m.frameIdx = 0
		case "end":
			m.cursor = m.findLeaf(false)
			m.frameIdx = 0
		case "pgup":
			for i := 0; i < 10; i++ {
				m.moveCursor(-1)
			}
		case "pgdown":
			for i := 0; i < 10; i++ {
				m.moveCursor(1)
			}
		case "+", "=":
			if m.fps < 30 {
				m.fps++
			}
		case "-", "_":
			if m.fps > 1 {
				m.fps--
			}
		}
	}
	return m, nil
}

func (m *previewModel) moveCursor(delta int) {
	for {
		next := m.cursor + delta
		if next < 0 || next >= len(m.tree) {
			return
		}
		m.cursor = next
		if m.tree[m.cursor].isLeaf {
			m.frameIdx = 0
			return
		}
	}
}

func (m previewModel) findLeaf(first bool) int {
	if first {
		for i, e := range m.tree {
			if e.isLeaf {
				return i
			}
		}
	} else {
		for i := len(m.tree) - 1; i >= 0; i-- {
			if m.tree[i].isLeaf {
				return i
			}
		}
	}
	return 0
}

func (m previewModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	// Layout: left preview (55%) │ right tree (45%)
	leftW := m.width * 55 / 100
	rightW := m.width - leftW - 3 // 3 for separator "│"
	if leftW < 30 {
		leftW = 30
	}
	if rightW < 20 {
		rightW = 20
	}

	leftPanel := m.renderPreview(leftW)
	rightPanel := m.renderTree(rightW)

	sep := strings.Repeat("│\n", m.height-2)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, sep, rightPanel)
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderPreview renders the left panel with animated ASCII art.
func (m previewModel) renderPreview(width int) string {
	entry := m.tree[m.cursor]
	frame := m.pack.Frames[entry.frameKey]

	var artStr string
	frameCount := 0
	if len(frame.Frames) == 0 {
		artStr = "  (无帧数据)"
	} else {
		frameCount = len(frame.Frames)
		idx := m.frameIdx % frameCount
		raw := strings.TrimRight(frame.Frames[idx], "\n")
		artStr = pvNormalizeArt(raw, frame.Width)
	}

	title := pvTitleStyle.Render(fmt.Sprintf(" ▶ %s ", entry.frameKey))
	info := pvInfoStyle.Render(fmt.Sprintf(
		"帧 %d/%d  %d fps  (+/-调速)",
		m.frameIdx%max(frameCount, 1)+1, max(frameCount, 1), m.fps,
	))

	helpText := pvInfoStyle.Render("↑↓选择  q退出")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		artStr,
		"",
		info,
		helpText,
	)

	return pvPanelStyle.
		Width(width - 2).
		Height(m.height - 1).
		Render(content)
}

// renderTree renders the right panel with the selectable tree list.
func (m previewModel) renderTree(width int) string {
	title := pvTreeTitleStyle.Render(" 帧列表 ")

	listH := m.height - 8
	if listH < 5 {
		listH = 5
	}

	// Scrolling window around cursor
	start := 0
	if m.cursor >= listH {
		start = m.cursor - listH/2
	}
	end := start + listH
	if end > len(m.tree) {
		end = len(m.tree)
		start = max(0, end-listH)
	}

	var lines []string
	for i := start; i < end; i++ {
		e := m.tree[i]
		indent := strings.Repeat("  ", e.indent)
		prefix := "  "
		style := pvTreeItemStyle

		if i == m.cursor && e.isLeaf {
			prefix = "▸ "
			style = pvTreeSelStyle
		}

		label := e.label
		if e.isLeaf {
			label = fmt.Sprintf("%s (%d帧)", e.label, e.children)
		}

		line := indent + prefix + label
		if ansi.StringWidth(line) > width-4 {
			line = pvTruncate(line, width-6) + ".."
		}
		lines = append(lines, style.Render(line))
	}

	scrollInfo := ""
	if len(m.tree) > listH {
		pct := (m.cursor + 1) * 100 / len(m.tree)
		scrollInfo = pvInfoStyle.Render(fmt.Sprintf(" %d/%d (%d%%)", m.cursor+1, len(m.tree), pct))
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		strings.Join(lines, "\n"),
		"",
		scrollInfo,
	)

	return pvTreePanelStyle.
		Width(width - 2).
		Height(m.height - 1).
		Render(content)
}

// buildTree creates a hierarchical tree of frames: phase → stageID → animState.
func buildTree(pack *plugin.SpeciesPack) []treeEntry {
	stagePhase := make(map[string]string)
	stageName := make(map[string]string)
	for _, s := range pack.Stages {
		stagePhase[s.ID] = s.Phase
		stageName[s.ID] = s.Name
	}

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

	var tree []treeEntry

	for _, phase := range phases {
		stages := phaseStages[phase]

		tree = append(tree, treeEntry{
			label:  fmt.Sprintf("── %s ──", strings.ToUpper(phase)),
			indent: 0,
		})

		stageIDs := make([]string, 0, len(stages))
		for sid := range stages {
			stageIDs = append(stageIDs, sid)
		}
		sort.Strings(stageIDs)

		for _, sid := range stageIDs {
			anims := stages[sid]

			sLabel := sid
			if name, ok := stageName[sid]; ok && name != "" {
				sLabel = fmt.Sprintf("%s [%s]", name, sid)
			}
			tree = append(tree, treeEntry{
				stageID: sid,
				label:   sLabel,
				indent:  1,
			})

			sort.Slice(anims, func(i, j int) bool {
				if anims[i].state == "idle" {
					return true
				}
				if anims[j].state == "idle" {
					return false
				}
				return anims[i].state < anims[j].state
			})

			for _, a := range anims {
				tree = append(tree, treeEntry{
					stageID:   sid,
					animState: a.state,
					label:     a.state,
					indent:    2,
					isLeaf:    true,
					phase:     phase,
					frameKey:  a.frameKey,
					children:  a.count,
				})
			}
		}
	}

	return tree
}

// pvNormalizeArt pads each line to the same display width.
func pvNormalizeArt(art string, minWidth int) string {
	lines := strings.Split(art, "\n")
	maxW := minWidth
	for _, l := range lines {
		if w := ansi.StringWidth(l); w > maxW {
			maxW = w
		}
	}
	for i, l := range lines {
		if w := ansi.StringWidth(l); w < maxW {
			lines[i] = l + strings.Repeat(" ", maxW-w)
		}
	}
	return strings.Join(lines, "\n")
}

// pvTruncate truncates a string to fit within the given display width.
func pvTruncate(s string, maxW int) string {
	w := 0
	for i, r := range s {
		rw := ansi.StringWidth(string(r))
		if w+rw > maxW {
			return s[:i]
		}
		w += rw
	}
	return s
}

// ---------- Styles ----------

var (
	pvPanelStyle = lipgloss.NewStyle().
			Padding(0, 1)

	pvTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	pvInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555570"))

	pvTreePanelStyle = lipgloss.NewStyle().
				Padding(0, 1)

	pvTreeTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#FFD700")).
				Padding(0, 1)

	pvTreeItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAEAEA"))

	pvTreeSelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)
)
