# TUI 组件重构 TODO

## 背景

dev 工具的 TUI 代码散落在 `cmd/clipet-dev/` 各文件中，存在大量重复：
- 进化树渲染有 3 份实现（evolve.go、evoinfo.go、preview.go）
- `normalizeArt` 有 2 份拷贝（petview.go、preview.go）
- 树导航逻辑各自实现，行为不一致
- 样式定义重复（同色值定义 4+ 次）

目标：将可复用的 TUI 逻辑提取到 `internal/tui/components/`，消除重复。

---

## 1. 创建 TreeList 组件

**位置**: `internal/tui/components/treelist.go`

### 1.1 设计原则

- 遵循 bubbles 子组件模式：`Update(tea.Msg) (TreeList, tea.Cmd)` + `View() string`
- 通过 `tea.Cmd` 发送事件消息（`TreeSelectMsg`），不使用 `HandleKey() bool` hack
- 不依赖 `charm.land/bubbles/v2`（避免新增依赖），仅依赖 `bubbletea/v2` 和 `lipgloss/v2`

### 1.2 数据结构

```go
// TreeNode 树节点
type TreeNode struct {
    ID         string
    Label      string
    Children   []*TreeNode
    Selectable bool        // false 时跳过选择（如 phase 分组标题）
    Expanded   bool        // 是否展开子节点
    Data       any         // 调用方自定义数据（如 frameKey、stageID）
    parent     *TreeNode   // 内部字段，由 InitTree 设置
}

// TreeList 组件模型
type TreeList struct {
    Roots          []*TreeNode
    KeyMap         TreeKeyMap
    Styles         TreeStyles
    ShowConnectors bool       // true: 显示树形连接线（├── └──）; false: 纯缩进
    MarkedID       string     // 标记特殊节点（如当前阶段），显示 MarkerPrefix
    MarkerPrefix   string     // 默认 "▸ "

    cursor    int           // visible 数组中的索引
    visible   []*TreeNode   // 当前可见行（已展开的节点列表）
    width     int
    height    int
    scrollOff int           // 滚动偏移
}
```

### 1.3 导航行为（标准树控件）

| 按键 | 行为 |
|------|------|
| ↑/k | 移动到上一个可见行（跳过 `Selectable=false` 的节点） |
| ↓/j | 移动到下一个可见行（跳过 `Selectable=false` 的节点） |
| ← /h | 若当前节点已展开 → 折叠；否则 → 跳到父节点 |
| → /l | 若当前节点已折叠 → 展开；否则 → 进入第一个子节点 |
| Space | 切换展开/折叠 |
| Enter | 选中当前节点，通过 `tea.Cmd` 发送 `TreeSelectMsg` |
| Home | 跳到第一个可选节点 |
| End | 跳到最后一个可选节点 |
| PgUp/PgDn | 翻页 |

### 1.4 显示

- 折叠指示器：`▸`（折叠）/ `▾`（展开），无子节点不显示
- 连接线（`ShowConnectors=true` 时）：`├── ` / `└── ` / `│   ` / `    `
- 光标行高亮（`Styles.Cursor`）
- 标记节点高亮（`Styles.Marked`）
- 滚动窗口：基于 `height` 自动裁剪可见范围
- 折叠时：光标在被折叠子树内 → 自动移到折叠节点

### 1.5 公开 API

```go
// 构造
func NewTreeList(roots []*TreeNode) TreeList

// bubbles 模式
func (t TreeList) Update(msg tea.Msg) (TreeList, tea.Cmd)
func (t TreeList) View() string

// 事件消息
type TreeSelectMsg struct { Node *TreeNode }

// 辅助方法
func (t *TreeList) SetSize(w, h int)
func (t TreeList) Selected() *TreeNode        // 当前光标节点
func (t *TreeList) SetCursor(id string)       // 按 ID 定位（自动展开祖先）
func (t *TreeList) ExpandAll()
func (t *TreeList) CollapseAll()
func (t TreeList) RenderPlain() string         // 纯文本输出（用于 stdout/evo info）

// 样式与键绑定
func DefaultTreeStyles() TreeStyles
func DefaultTreeKeyMap() TreeKeyMap
```

### 1.6 TreeStyles

```go
type TreeStyles struct {
    Cursor    lipgloss.Style  // 光标行
    Marked    lipgloss.Style  // 标记节点（如当前阶段）
    Normal    lipgloss.Style  // 普通节点
    Connector lipgloss.Style  // 连接线颜色
    Indicator lipgloss.Style  // ▸/▾ 指示器
    ScrollInfo lipgloss.Style // 滚动信息
}
```

---

## 2. 导出 NormalizeArt

**位置**: `internal/tui/components/petview.go`

- `normalizeArt` → `NormalizeArt`（导出）
- `displayWidth` → `DisplayWidth`（导出）
- 删除 `preview.go` 中的 `pvNormalizeArt` 重复实现

---

## 3. 改造 preview.go 使用 TreeList

**文件**: `cmd/clipet-dev/preview.go`

### 要删除的代码
- `treeEntry` 结构体
- `moveCursor()` / `findLeaf()` 方法
- `renderTree()` 方法
- `pvNormalizeArt()` 函数
- `pvTreeItemStyle` / `pvTreeSelStyle` / `pvTreePanelStyle` / `pvTreeTitleStyle` 等树相关样式

### 要修改的代码
- `buildTree()` 返回 `[]*components.TreeNode` 而不是 `[]treeEntry`
  - phase 标题节点：`Selectable: false`
  - stage 节点：`Selectable: false`
  - 动画叶子节点：`Selectable: true`，`Data` 存放 frameKey / count 等信息
- `previewModel` 中用 `tree components.TreeList` 替代 `tree []treeEntry` + `cursor int`
- `ShowConnectors: false`（preview 不需要连接线）
- `Update()` 中将 `tea.KeyPressMsg` 先传给 `tree.Update()`，监听 `TreeSelectMsg` 切换动画
- `View()` 中右侧用 `tree.View()` 替代 `renderTree()`
- 用 `components.NormalizeArt()` 替代 `pvNormalizeArt()`

### 配置
- 初始展开所有节点（`ExpandAll`）— preview 需要看到全部列表
- `findInitialCursor` → `tree.SetCursor(id)`

---

## 4. 改造 evolve.go 使用 TreeList

**文件**: `cmd/clipet-dev/evolve.go`

### 要删除的代码
- `evoNode` 结构体
- `moveSibling()` / `moveToParent()` / `moveToChild()` / `getSiblings()` 方法
- `renderTree()` 方法

### 要修改的代码
- 构建 `[]*components.TreeNode`，从 `pack.Stages` + `pack.Evolutions` 生成树
  - 所有节点 `Selectable: true`
  - `Data` 存放 stageID
- `evoModel` 中用 `tree components.TreeList` 替代 `roots`/`allNodes`/`cursor`
- `ShowConnectors: true`
- `MarkedID: pet.StageID`（标记当前阶段）
- `Update()` 传递消息给 `tree.Update()`，监听 `TreeSelectMsg` 执行进化
- `View()` 使用 `tree.View()`

### 配置
- 初始展开所有节点
- `tree.SetCursor(pet.StageID)`

---

## 5. 改造 evoinfo.go 使用 TreeList

**文件**: `cmd/clipet-dev/evoinfo.go`

### 要修改的代码
- `printEvoTree()` 函数改为构建 `TreeList` 后调用 `RenderPlain()`
- 或直接用 `fmt.Print(tree.RenderPlain())` 输出

### 注意
- evoinfo 是纯 CLI 输出，不是 TUI，所以只用 `RenderPlain()`
- `MarkedID` 设为当前阶段 ID 以显示 `▸` 标记
- `condSummary()`、`checkMark()`、`attrName()` 等辅助函数保留不动

---

## 6. 构建验证

```bash
go build ./...
```

确保所有文件编译通过，无导入错误。

---

## 暂不处理

- **样式统一**：各文件的 lipgloss 样式仍各自定义，暂不提取到 `styles/theme.go`
- **进度条组件**：`set.go` / `timeskip.go` 的 `renderBar()` 暂不提取
- **键盘导航组件**：TreeList 之外的导航逻辑暂不抽象
