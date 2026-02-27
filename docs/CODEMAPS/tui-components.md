<!-- Generated: 2026-02-27 | Files scanned: 15 | Token estimate: ~650 -->

# TUI Components & Screens

## Framework
Bubble Tea v2 + Lipgloss v2

## Component Hierarchy

```
internal/tui/
├── app.go (root model + screen routing)
├── components/
│   ├── petview.go       (animated ASCII pet display)
│   ├── dialoguebubble.go (speech bubbles)
│   ├── treelist.go      (reusable tree navigation)
│   ├── progressbar.go   (attribute bars)
│   └── inputfield.go    (text input wrapper)
├── screens/
│   ├── home.go          (main menu + pet view)
│   ├── evolve.go        (evolution selection)
│   └── adventure.go     (adventure flow)
├── dev/                 (dev tools TUI)
│   ├── preview.go       (frame viewer)
│   ├── evolve.go        (force evolution)
│   ├── set.go           (attribute editor)
│   └── timeskip.go      (time simulation)
└── styles/
    ├── theme.go         (color schemes)
    ├── common.go        (shared styles)
    └── colors.go        (color utilities)
```

## Reusable Components

### TreeList (treelist.go)

**Purpose**: Generic tree navigation with cursor + expand/collapse

**Used by**: evolve, preview, evoinfo commands

**Key Features**:
- `View()` → TUI rendering
- `RenderPlain()` → CLI output (no colors/cursor)
- Visible list pattern (flattened expanded nodes)
- Cursor navigation (↑↓)
- Expand/collapse (←→)
- Node marking (current stage highlight)

**Data Structure**:
```go
type TreeNode struct {
    ID, Label string
    Children []*TreeNode
    Selectable, Expanded bool
    Data interface{}  // Custom payload
}

type TreeList struct {
    Roots []*TreeNode
    Cursor int
    MarkedID string
    ShowConnectors bool
}
```

**Messages**:
```go
TreeSelectMsg{Node *TreeNode}  // Enter pressed
```

### PetView (petview.go)

**Purpose**: Animated ASCII pet display with frame cycling

**Key Functions**:
```go
NormalizeArt(art string, width int) string
  // Pad lines to consistent width for centering

DisplayWidth(s string) int
  // Use charmbracelet/x/ansi.StringWidth()
```

**Animation Loop**:
```
FrameTickMsg → FrameIdx++ → View() → select frame[idx % len]
```

### DialogueBubble (dialoguebubble.go)

**Purpose**: Render pet dialogue in speech bubble style

**Style**: Lipgloss rounded borders with padding

### ProgressBar (progressbar.go)

**Purpose**: Attribute visualization (Hunger/Happiness/etc.)

**API**:
```go
NewProgressBar().
    SetValue(75).
    SetMax(100).
    SetWidth(20).
    SetFilledStyle(style).
    Render() → string
```

## Screens

### Home Screen (home.go)

**Structure**:
```
┌─────────────────────────┐
│  [Pet ASCII Art]        │  PetView component
│  [Dialogue Bubble]      │  DialogueBubble
├─────────────────────────┤
│  [Menu]                 │
│  ▸ 喂食                 │  2nd level menu
│    玩耍                 │
│    休息                 │
└─────────────────────────┘
```

**States**:
- `homeStateMenu` → navigation
- `homeStateAction` → action animation

### Evolve Screen (evolve.go)

**Flow**:
```
showEvolveSelect → TreeList (choose stage)
    ↓
showEvolveAnim → evolution animation
    ↓
showEvolveDone → completion message
```

### Adventure Screen (adventure.go)

**Flow**:
```
showAdventureIntro → story text
    ↓
showAdventureChoice → choice selection
    ↓
showAdventureAnim → result animation
    ↓
showAdventureResult → attribute changes
```

## Dev Tools TUI (dev/)

### Shared Patterns

All dev commands use:
- `help.Model` for keybinding display
- `key.Binding` for key mappings
- `KeyMap` interface (ShortHelp/FullHelp)
- Toggle help with `?` key

### Preview Command (preview.go)

**Layout**: Left 55% (frame preview) | Right 45% (tree list)

**Features**:
- Frame animation with FPS control
- Speed adjustment (+/-)
- Tree navigation for frame selection

### Set Command (set.go)

**Phases**:
1. `setPhaseSelect` → field list (↑↓, Enter)
2. `setPhaseInput` → text input (Enter to save, Esc to cancel)

**Dual KeyMaps**: SetSelectKeyMap + SetInputKeyMap

### Timeskip Command (timeskip.go)

**Phases**:
1. `timeskipPhaseInput` → hours input
2. `timeskipPhasePreview` → preview changes, confirm/cancel

**Preview Calculation**:
```
OldStats → simulate decay → NewStats
Show delta with +/- indicators
Warn if pet would die (health ≤ 0)
```

## Styling System (styles/)

### Theme Definition (theme.go)

```go
var DevCommandStyles = struct {
    Panel, Title, Info lipgloss.Style
    Item, SelItem lipgloss.Style
    BarFilled, BarEmpty lipgloss.Style
    InputLabel, Error lipgloss.Style
}{...}
```

### Color Utilities (colors.go)

```go
DimColor(), GoldColor(), TextColor() → color.Color
```

### Common Styles (common.go)

```go
MakeTitleStyle(bgColor) → lipgloss.Style
  // Standardized title formatting
```
