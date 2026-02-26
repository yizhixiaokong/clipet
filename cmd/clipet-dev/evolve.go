package main

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// newEvoCmd creates the parent "evo" command with subcommands "to" and "info".
func newEvoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evo",
		Short: "[开发] 进化相关命令",
		Long:  "进化相关的开发工具集，包含强制进化和查看进化信息子命令。",
	}

	cmd.AddCommand(newEvoToCmd())
	cmd.AddCommand(newEvoInfoSubCmd())

	return cmd
}

func newEvoToCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "to [stage-id]",
		Short: "强制进化到指定阶段",
		Long: `跳过所有进化条件，直接将宠物设置为指定的 stage ID。

不带参数时进入交互式界面：左侧进化树，右侧选择目标阶段。
带参数时直接执行进化。`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
			}

			pack := registry.GetSpecies(pet.Species)
			if pack == nil {
				return fmt.Errorf("species %q not found in registry", pet.Species)
			}

			// Direct mode with argument
			if len(args) == 1 {
				return doEvolve(cmd, pet, pack, args[0])
			}

			// Interactive mode
			return runEvoTUI(pet, pack)
		},
	}
}

func doEvolve(cmd *cobra.Command, pet *game.Pet, pack *plugin.SpeciesPack, targetID string) error {
	stage := registry.GetStage(pet.Species, targetID)
	if stage == nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "stage %q not found for species %q\n\n", targetID, pet.Species)
		fmt.Fprintln(cmd.ErrOrStderr(), "可用阶段:")
		for _, s := range pack.Stages {
			fmt.Fprintf(cmd.ErrOrStderr(), "  %-25s (%s)\n", s.ID, s.Phase)
		}
		return fmt.Errorf("invalid stage ID %q", targetID)
	}

	oldStageID := pet.StageID
	oldPhase := string(pet.Stage)

	pet.StageID = targetID
	pet.Stage = game.PetStage(stage.Phase)

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Printf("🔄 进化完成: %s (%s) → %s (%s)\n", oldStageID, oldPhase, pet.StageID, stage.Phase)
	fmt.Printf("   阶段名称: %s\n", stage.Name)
	return nil
}

// ---------- Interactive evo to TUI ----------

// evoNode is a node in the evolution tree for navigation.
type evoNode struct {
	id       string
	name     string
	phase    string
	parent   *evoNode
	children []*evoNode
}

type evoModel struct {
	pet      *game.Pet
	pack     *plugin.SpeciesPack
	roots    []*evoNode
	allNodes map[string]*evoNode
	cursor   *evoNode // currently highlighted node
	width    int
	height   int
	done     bool
	result   string
	quitting bool
}

func runEvoTUI(pet *game.Pet, pack *plugin.SpeciesPack) error {
	// Build tree structure
	allNodes := make(map[string]*evoNode)
	for _, s := range pack.Stages {
		allNodes[s.ID] = &evoNode{id: s.ID, name: s.Name, phase: s.Phase}
	}

	// Link parent-child via evolutions
	for _, e := range pack.Evolutions {
		parent := allNodes[e.From]
		child := allNodes[e.To]
		if parent != nil && child != nil {
			child.parent = parent
			parent.children = append(parent.children, child)
		}
	}

	// Find roots (nodes with no parent — typically egg phase)
	var roots []*evoNode
	for _, s := range pack.Stages {
		node := allNodes[s.ID]
		if node.parent == nil {
			roots = append(roots, node)
		}
	}

	// Initial cursor: current stage
	cursor := allNodes[pet.StageID]
	if cursor == nil && len(roots) > 0 {
		cursor = roots[0]
	}

	m := evoModel{
		pet:      pet,
		pack:     pack,
		roots:    roots,
		allNodes: allNodes,
		cursor:   cursor,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	fm := finalModel.(evoModel)
	if fm.done && fm.result != "" {
		stage := registry.GetStage(pet.Species, fm.result)
		oldStageID := pet.StageID
		oldPhase := string(pet.Stage)

		pet.StageID = fm.result
		pet.Stage = game.PetStage(stage.Phase)

		if err := petStore.Save(pet); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		fmt.Printf("🔄 进化完成: %s (%s) → %s (%s)\n", oldStageID, oldPhase, pet.StageID, stage.Phase)
		fmt.Printf("   阶段名称: %s\n", stage.Name)
	}
	return nil
}

func (m evoModel) Init() tea.Cmd { return nil }

func (m evoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "escape":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			m.moveSibling(-1)
		case "down", "j":
			m.moveSibling(1)
		case "left", "h":
			m.moveToParent()
		case "right", "l":
			m.moveToChild()
		case "enter":
			m.done = true
			m.result = m.cursor.id
			return m, tea.Quit
		}
	}
	return m, nil
}

// moveSibling moves up/down among siblings (same parent).
func (m *evoModel) moveSibling(delta int) {
	siblings := m.getSiblings()
	idx := 0
	for i, s := range siblings {
		if s.id == m.cursor.id {
			idx = i
			break
		}
	}
	next := idx + delta
	if next >= 0 && next < len(siblings) {
		m.cursor = siblings[next]
	}
}

// moveToParent navigates to the parent node.
func (m *evoModel) moveToParent() {
	if m.cursor.parent != nil {
		m.cursor = m.cursor.parent
	}
}

// moveToChild navigates to the first child node.
func (m *evoModel) moveToChild() {
	if len(m.cursor.children) > 0 {
		m.cursor = m.cursor.children[0]
	}
}

// getSiblings returns the sibling list that contains the cursor.
func (m *evoModel) getSiblings() []*evoNode {
	if m.cursor.parent != nil {
		return m.cursor.parent.children
	}
	return m.roots
}

func (m evoModel) View() tea.View {
	if m.quitting || m.done {
		return tea.NewView("")
	}
	if m.width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	title := evoHeaderStyle.Render(fmt.Sprintf(" 进化树 — %s [%s] ", m.pet.Name, m.pet.StageID))

	treeStr := m.renderTree()

	selected := m.cursor
	info := evoInfoStyle.Render(fmt.Sprintf("选中: %s [%s] (%s)", selected.name, selected.id, selected.phase))

	isCurrent := ""
	if selected.id == m.pet.StageID {
		isCurrent = evoInfoStyle.Render("  (当前阶段)")
	}

	help := evoInfoStyle.Render("↑↓同级切换  ←父级 →子级  Enter确认  q退出")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		treeStr,
		"",
		info+isCurrent,
		help,
	)

	panel := evoLeftStyle.
		Width(m.width - 2).
		Height(m.height - 1).
		Render(content)

	v := tea.NewView(panel)
	v.AltScreen = true
	return v
}

func (m evoModel) renderTree() string {
	var sb strings.Builder
	cursorID := m.cursor.id
	currentID := m.pet.StageID

	var printNode func(node *evoNode, prefix string, isLast bool)
	printNode = func(node *evoNode, prefix string, isLast bool) {
		conn := "├── "
		if isLast {
			conn = "└── "
		}

		name := node.name
		if name == "" {
			name = node.id
		}

		marker := "  "
		if node.id == currentID {
			marker = "▸ "
		}

		line := fmt.Sprintf("%s%s%s%s (%s)", prefix, conn, marker, name, node.phase)

		if node.id == cursorID {
			line = evoSelStyle.Render(line)
		} else if node.id == currentID {
			line = evoHighlightStyle.Render(line)
		}
		sb.WriteString(line + "\n")

		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}
		for i, child := range node.children {
			printNode(child, childPrefix, i == len(node.children)-1)
		}
	}

	for i, root := range m.roots {
		name := root.name
		if name == "" {
			name = root.id
		}
		marker := "  "
		if root.id == currentID {
			marker = "▸ "
		}

		line := fmt.Sprintf("%s%s (%s)", marker, name, root.phase)
		if root.id == cursorID {
			line = evoSelStyle.Render(line)
		} else if root.id == currentID {
			line = evoHighlightStyle.Render(line)
		}
		sb.WriteString(line + "\n")

		for j, child := range root.children {
			printNode(child, "", j == len(root.children)-1)
		}

		if i < len(m.roots)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

var (
	evoLeftStyle = lipgloss.NewStyle().Padding(0, 1)

	evoHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	evoInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555570"))

	evoSelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)

	evoHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)
)
