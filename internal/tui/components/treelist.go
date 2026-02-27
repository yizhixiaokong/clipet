package components

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TreeNode represents a single node in the tree
type TreeNode struct {
	ID         string      // Unique identifier for the node
	Label      string      // Display text
	Children   []*TreeNode // Child nodes
	Selectable bool        // false = skip cursor (e.g., phase headers)
	Expanded   bool        // Current expansion state
	Data       any         // Custom data (frameKey, stageID, etc.)

	// Internal fields (set by TreeList.initParents)
	parent *TreeNode
	depth  int
}

// TreeList is a reusable tree navigation component
type TreeList struct {
	Roots          []*TreeNode
	Styles         TreeStyles
	KeyMap         TreeKeyMap
	ShowConnectors bool   // Display tree connectors (├── └──)
	MarkedID       string // Node to mark with special indicator
	MarkerPrefix   string // Marker for special node (default: "▸ ")

	// Internal state
	cursor    int         // Index in visible slice
	visible   []*TreeNode // Flattened visible nodes
	width     int
	height    int
	scrollOff int // Vertical scroll offset
}

// TreeStyles holds lipgloss styles for tree rendering
type TreeStyles struct {
	Cursor     lipgloss.Style // Highlighted cursor line
	Marked     lipgloss.Style // Marked node (e.g., current stage)
	Normal     lipgloss.Style // Default node style
	Connector  lipgloss.Style // Tree connector lines
	Indicator  lipgloss.Style // Expand/collapse indicator (▸/▾)
	ScrollInfo lipgloss.Style // Scroll position info
}

// TreeKeyMap defines keyboard navigation
type TreeKeyMap struct {
	Up       string
	Down     string
	Left     string
	Right    string
	Space    string
	Enter    string
	Home     string
	End      string
	PageUp   string
	PageDown string
}

// TreeSelectMsg is sent when user presses Enter on a node
type TreeSelectMsg struct {
	Node *TreeNode
}

// NewTreeList creates a new tree component
func NewTreeList(roots []*TreeNode) TreeList {
	t := TreeList{
		Roots:          roots,
		Styles:         DefaultTreeStyles(),
		KeyMap:         DefaultTreeKeyMap(),
		ShowConnectors: true,
		MarkerPrefix:   "▸ ",
		cursor:         0,
	}
	t.initParents(roots, nil, 0)
	t.rebuildVisible()
	t.findFirstSelectable()
	return t
}

// initParents recursively sets parent and depth fields
func (t *TreeList) initParents(nodes []*TreeNode, parent *TreeNode, depth int) {
	for _, node := range nodes {
		node.parent = parent
		node.depth = depth
		t.initParents(node.Children, node, depth+1)
	}
}

// rebuildVisible reconstructs the visible node list
func (t *TreeList) rebuildVisible() {
	t.visible = nil
	t.flattenVisible(t.Roots)
}

// flattenVisible recursively adds expanded nodes to visible list
func (t *TreeList) flattenVisible(nodes []*TreeNode) {
	for _, node := range nodes {
		t.visible = append(t.visible, node)
		if node.Expanded && len(node.Children) > 0 {
			t.flattenVisible(node.Children)
		}
	}
}

// DefaultTreeStyles returns the default styling
func DefaultTreeStyles() TreeStyles {
	return TreeStyles{
		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true),
		Marked: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true),
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAEAEA")),
		Connector: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666680")),
		Indicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8888AA")),
		ScrollInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555570")),
	}
}

// DefaultTreeKeyMap returns default key bindings
func DefaultTreeKeyMap() TreeKeyMap {
	return TreeKeyMap{
		Up:       "up",
		Down:     "down",
		Left:     "left",
		Right:    "right",
		Space:    " ",
		Enter:    "enter",
		Home:     "home",
		End:      "end",
		PageUp:   "pgup",
		PageDown: "pgdown",
	}
}

// Init implements tea.Model
func (t TreeList) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (t TreeList) Update(msg tea.Msg) (TreeList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return t.handleKeyPress(msg)
	}
	return t, nil
}

// View implements tea.Model (for TUI)
func (t TreeList) View() string {
	if len(t.visible) == 0 {
		return ""
	}

	// Calculate scroll window
	start, end := t.calculateScrollWindow()

	var lines []string
	for i := start; i < end; i++ {
		node := t.visible[i]
		line := t.renderNode(node, i == t.cursor)
		lines = append(lines, line)
	}

	result := strings.Join(lines, "\n")

	// Add scroll info if needed
	if len(t.visible) > t.height && t.height > 0 {
		pct := (t.cursor + 1) * 100 / len(t.visible)
		info := t.Styles.ScrollInfo.Render(
			fmt.Sprintf(" %d/%d (%d%%)", t.cursor+1, len(t.visible), pct),
		)
		result = result + "\n" + info
	}

	return result
}

// RenderPlain returns plain text output (for CLI)
func (t TreeList) RenderPlain() string {
	var sb strings.Builder
	t.renderPlainTree(t.Roots, "", true, &sb)
	return sb.String()
}

// handleKeyPress processes keyboard input
func (t TreeList) handleKeyPress(msg tea.KeyPressMsg) (TreeList, tea.Cmd) {
	key := msg.String()

	// Check against key map
	if key == t.KeyMap.Up || key == "k" {
		t.moveCursor(-1)
		return t, nil
	}
	if key == t.KeyMap.Down || key == "j" {
		t.moveCursor(1)
		return t, nil
	}
	if key == t.KeyMap.Left || key == "h" {
		t.handleLeft()
		return t, nil
	}
	if key == t.KeyMap.Right || key == "l" {
		t.handleRight()
		return t, nil
	}
	if key == t.KeyMap.Space {
		t.toggleExpand()
		return t, nil
	}
	if key == t.KeyMap.Enter {
		node := t.Selected()
		if node != nil && node.Selectable {
			return t, func() tea.Msg {
				return TreeSelectMsg{Node: node}
			}
		}
		return t, nil
	}
	if key == t.KeyMap.Home {
		t.cursorHome()
		return t, nil
	}
	if key == t.KeyMap.End {
		t.cursorEnd()
		return t, nil
	}
	if key == t.KeyMap.PageUp {
		if t.height > 0 {
			t.moveCursor(-t.height)
		}
		return t, nil
	}
	if key == t.KeyMap.PageDown {
		if t.height > 0 {
			t.moveCursor(t.height)
		}
		return t, nil
	}

	return t, nil
}

// moveCursor moves cursor by delta, skipping non-selectable nodes
func (t *TreeList) moveCursor(delta int) {
	if len(t.visible) == 0 {
		return
	}

	start := t.cursor
	direction := 1
	if delta < 0 {
		direction = -1
	}

	steps := 0
	maxSteps := len(t.visible)

	for steps < maxSteps {
		t.cursor += direction

		// Bounds check
		if t.cursor < 0 {
			t.cursor = 0
			return
		}
		if t.cursor >= len(t.visible) {
			t.cursor = len(t.visible) - 1
			return
		}

		// Check if selectable
		if t.visible[t.cursor].Selectable {
			return
		}

		steps++
	}

	// Fallback: restore original position
	t.cursor = start
}

// handleLeft collapses or moves to parent
func (t *TreeList) handleLeft() {
	if t.cursor < 0 || t.cursor >= len(t.visible) {
		return
	}

	node := t.visible[t.cursor]
	if node.Expanded && len(node.Children) > 0 {
		// Collapse
		node.Expanded = false
		t.rebuildVisible()
		t.clampCursor()
	} else if node.parent != nil {
		// Move to parent
		t.setCursorToNode(node.parent)
	}
}

// handleRight expands or moves to first child
func (t *TreeList) handleRight() {
	if t.cursor < 0 || t.cursor >= len(t.visible) {
		return
	}

	node := t.visible[t.cursor]
	if len(node.Children) == 0 {
		return // No children
	}

	if !node.Expanded {
		// Expand
		node.Expanded = true
		t.rebuildVisible()
	} else {
		// Move to first child
		t.moveCursor(1)
	}
}

// toggleExpand flips expansion state
func (t *TreeList) toggleExpand() {
	if t.cursor < 0 || t.cursor >= len(t.visible) {
		return
	}

	node := t.visible[t.cursor]
	if len(node.Children) == 0 {
		return
	}

	node.Expanded = !node.Expanded
	t.rebuildVisible()
	t.clampCursor()
}

// clampCursor ensures cursor is valid after tree structure change
func (t *TreeList) clampCursor() {
	if len(t.visible) == 0 {
		t.cursor = 0
		return
	}

	if t.cursor >= len(t.visible) {
		t.cursor = len(t.visible) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}

	// Ensure cursor is on selectable node (only if there are multiple nodes)
	if len(t.visible) > 1 && !t.visible[t.cursor].Selectable {
		t.moveCursor(1)
		if len(t.visible) > 0 && t.cursor < len(t.visible) && !t.visible[t.cursor].Selectable {
			t.moveCursor(-1)
		}
	}

	// Final bounds check after moveCursor
	if len(t.visible) == 0 {
		t.cursor = 0
	} else if t.cursor >= len(t.visible) {
		t.cursor = len(t.visible) - 1
	}
}

// renderNode renders a single node
func (t TreeList) renderNode(node *TreeNode, isCursor bool) string {
	// Build raw prefix (without styles) for consistent styling
	var rawPrefix strings.Builder
	if t.ShowConnectors {
		rawPrefix.WriteString(t.buildConnectorPrefixRaw(node))
	} else {
		rawPrefix.WriteString(strings.Repeat("  ", node.depth))
	}

	// Add expand/collapse indicator (raw text)
	indicator := "  "
	if len(node.Children) > 0 {
		if node.Expanded {
			indicator = "▾ "
		} else {
			indicator = "▸ "
		}
	}
	rawPrefix.WriteString(indicator)

	// Get label
	label := node.Label

	// Calculate marker width for truncation
	markerWidth := 0
	if node.ID == t.MarkedID {
		markerWidth = ansi.StringWidth(" [当前]")
	}

	// Truncate label if needed
	if t.width > 0 {
		maxWidth := t.width - markerWidth - ansi.StringWidth(rawPrefix.String())
		if maxWidth > 0 && ansi.StringWidth(label) > maxWidth {
			label = t.truncate(label, maxWidth-2) + ".."
		}
	}

	// Build complete raw line (without any styles)
	rawLine := rawPrefix.String() + label

	// Determine which style to apply
	var targetStyle lipgloss.Style
	if isCursor {
		targetStyle = t.Styles.Cursor
	} else if node.ID == t.MarkedID {
		targetStyle = t.Styles.Marked
	} else {
		targetStyle = t.Styles.Normal
	}

	// Apply style to entire line (this ensures full background coverage)
	styledLine := targetStyle.Render(rawLine)

	// Add marker suffix at the end (with appropriate style)
	if node.ID == t.MarkedID {
		marker := " [当前]"
		// Use same style as the line for consistent appearance
		styledLine = styledLine + targetStyle.Render(marker)
	}

	return styledLine
}

// buildConnectorPrefixRaw builds connector prefix without any styles
func (t TreeList) buildConnectorPrefixRaw(node *TreeNode) string {
	if node.depth == 0 {
		return ""
	}

	var prefix strings.Builder
	current := node
	var ancestors []bool
	visited := make(map[*TreeNode]bool)
	maxDepth := 100

	for current.parent != nil && len(ancestors) < maxDepth {
		if visited[current] {
			break
		}
		visited[current] = true

		parent := current.parent
		if parent == nil {
			break
		}

		isLast := true
		for i, sibling := range parent.Children {
			if sibling == current {
				isLast = (i == len(parent.Children)-1)
				break
			}
		}
		ancestors = append([]bool{isLast}, ancestors...)
		current = parent
	}

	for i, isLast := range ancestors {
		if i == len(ancestors)-1 {
			if isLast {
				prefix.WriteString("└── ")
			} else {
				prefix.WriteString("├── ")
			}
		} else {
			if isLast {
				prefix.WriteString("    ")
			} else {
				prefix.WriteString("│   ")
			}
		}
	}

	return prefix.String()
}

// renderPlainTree renders tree for CLI output (no colors/cursor)
func (t TreeList) renderPlainTree(nodes []*TreeNode, prefix string, isLast bool, sb *strings.Builder) {
	for i, node := range nodes {
		isLastChild := i == len(nodes)-1

		// Connector for this node
		connector := "├── "
		if isLastChild {
			connector = "└── "
		}

		// Marker for marked node
		marker := "  "
		if node.ID == t.MarkedID {
			marker = t.MarkerPrefix
		}

		// Build line
		line := fmt.Sprintf("%s%s%s%s", prefix, connector, marker, node.Label)
		sb.WriteString(line + "\n")

		// Recursively render children
		if len(node.Children) > 0 {
			childPrefix := prefix
			if isLastChild {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
			t.renderPlainTree(node.Children, childPrefix, isLastChild, sb)
		}
	}
}

// SetSize updates component dimensions
func (t *TreeList) SetSize(w, h int) {
	t.width = w
	t.height = h
}

// Selected returns the currently selected node
func (t TreeList) Selected() *TreeNode {
	if t.cursor >= 0 && t.cursor < len(t.visible) {
		return t.visible[t.cursor]
	}
	return nil
}

// SetCursor moves cursor to node with given ID, expanding ancestors
func (t *TreeList) SetCursor(id string) {
	// Find node
	target := t.findNode(t.Roots, id)
	if target == nil {
		return
	}

	// Expand all ancestors
	current := target.parent
	for current != nil {
		current.Expanded = true
		current = current.parent
	}

	// Rebuild visible list
	t.rebuildVisible()

	// Set cursor to target
	t.setCursorToNode(target)
}

// findNode recursively searches for node by ID
func (t TreeList) findNode(nodes []*TreeNode, id string) *TreeNode {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
		if found := t.findNode(node.Children, id); found != nil {
			return found
		}
	}
	return nil
}

// setCursorToNode moves cursor to specific node
func (t *TreeList) setCursorToNode(target *TreeNode) {
	if len(t.visible) == 0 || target == nil {
		return
	}
	for i, node := range t.visible {
		if node == target {
			t.cursor = i
			return
		}
	}
}

// ExpandAll expands all nodes
func (t *TreeList) ExpandAll() {
	t.expandAllRecursive(t.Roots)
	t.rebuildVisible()
}

func (t *TreeList) expandAllRecursive(nodes []*TreeNode) {
	for _, node := range nodes {
		node.Expanded = true
		t.expandAllRecursive(node.Children)
	}
}

// CollapseAll collapses all nodes
func (t *TreeList) CollapseAll() {
	t.collapseAllRecursive(t.Roots)
	t.rebuildVisible()
	t.clampCursor()
}

func (t *TreeList) collapseAllRecursive(nodes []*TreeNode) {
	for _, node := range nodes {
		node.Expanded = false
		t.collapseAllRecursive(node.Children)
	}
}

// ExpandToLevel expands nodes up to the specified level (0 = only roots, 1 = roots + 1 level deep, etc.)
func (t *TreeList) ExpandToLevel(maxLevel int) {
	t.expandToLevelRecursive(t.Roots, 0, maxLevel)
	t.rebuildVisible()
}

func (t *TreeList) expandToLevelRecursive(nodes []*TreeNode, currentLevel, maxLevel int) {
	for _, node := range nodes {
		node.Expanded = currentLevel < maxLevel
		t.expandToLevelRecursive(node.Children, currentLevel+1, maxLevel)
	}
}

// calculateScrollWindow determines visible range based on cursor position
func (t TreeList) calculateScrollWindow() (start, end int) {
	if t.height <= 0 || len(t.visible) <= t.height {
		return 0, len(t.visible)
	}

	// Keep cursor in view with some context
	halfHeight := t.height / 2

	start = t.cursor - halfHeight
	if start < 0 {
		start = 0
	}

	end = start + t.height
	if end > len(t.visible) {
		end = len(t.visible)
		start = end - t.height
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

// truncate truncates string to display width
func (t TreeList) truncate(s string, maxW int) string {
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

// findFirstSelectable moves cursor to first selectable node
func (t *TreeList) findFirstSelectable() {
	for i, node := range t.visible {
		if node.Selectable {
			t.cursor = i
			return
		}
	}
}

// cursorHome moves to first selectable node
func (t *TreeList) cursorHome() {
	t.cursor = 0
	if len(t.visible) > 0 && !t.visible[0].Selectable {
		t.moveCursor(1)
	}
}

// cursorEnd moves to last selectable node
func (t *TreeList) cursorEnd() {
	if len(t.visible) == 0 {
		return
	}

	t.cursor = len(t.visible) - 1
	if !t.visible[t.cursor].Selectable {
		t.moveCursor(-1)
	}
}
