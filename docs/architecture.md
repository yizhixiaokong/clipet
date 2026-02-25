# Clipet 架构文档

## 模块结构

```
clipet/
├── cmd/clipet/main.go          # 入口点（待实现）
├── internal/
│   ├── assets/
│   │   ├── embed.go            # go:embed 嵌入内置物种包
│   │   └── builtins/           # 内置物种包数据
│   │       └── cat-pack/       # 猫物种包
│   ├── plugin/                 # 插件系统
│   │   ├── types.go            # 数据类型定义
│   │   ├── parser.go           # TOML 解析 + 帧扫描
│   │   ├── validator.go        # 插件包校验
│   │   ├── loader.go           # 文件系统发现与加载
│   │   └── registry.go         # 中央注册表
│   ├── game/
│   │   └── pet.go              # 核心游戏逻辑（UI 无关）
│   ├── store/
│   │   ├── store.go            # 持久化接口
│   │   └── jsonstore.go        # JSON 文件实现
│   ├── cli/
│   │   ├── root.go             # Cobra 根命令 + 初始化
│   │   ├── init.go             # clipet init 命令
│   │   ├── status.go           # clipet status 命令
│   │   ├── feed.go             # clipet feed 命令
│   │   ├── play.go             # clipet play 命令
│   │   └── tui_bridge.go       # TUI 启动桥接（待实现）
│   └── tui/                    # TUI 层（待实现）
│       ├── app.go              # 顶层 Model + 屏幕路由
│       ├── styles/
│       │   └── theme.go        # Lipgloss 样式常量
│       ├── components/
│       │   └── petview.go      # 宠物 ASCII 渲染组件
│       └── screens/
│           └── home.go         # 主屏幕
├── docs/                       # 文档
├── go.mod
├── go.sum
└── .gitignore
```

## 分层架构

```
┌─────────────────────────────────────────┐
│              cmd/clipet/main.go         │  入口
├─────────────────────────────────────────┤
│         internal/cli/ (Cobra)           │  CLI 命令 + 路由
├──────────────────┬──────────────────────┤
│  internal/tui/   │  internal/cli/*.go   │  TUI 界面 / CLI 快捷命令
│  (Bubble Tea)    │  (Cobra subcommands) │
├──────────────────┴──────────────────────┤
│           internal/game/                │  核心游戏逻辑（UI 无关）
├─────────────────────────────────────────┤
│           internal/store/               │  持久化层
├─────────────────────────────────────────┤
│           internal/plugin/              │  插件系统
├─────────────────────────────────────────┤
│           internal/assets/              │  嵌入资源
└─────────────────────────────────────────┘
```

### 设计原则

1. **游戏逻辑与 UI 分离**: `internal/game/` 不依赖任何 TUI/CLI 包
2. **统一插件接口**: 内置和外部插件通过同一个 `fs.FS` → `Registry` 路径加载
3. **持久化接口抽象**: `Store` 接口可替换为其他实现
4. **单一初始化入口**: `root.go` 的 `PersistentPreRunE` 统一完成 Registry + Store 初始化

## 关键类型

### plugin.SpeciesPack

物种包的顶层结构，包含：
- `SpeciesConfig` — 元信息 (ID, 名称, 作者, 版本, 基础属性)
- `[]Stage` — 进化阶段节点
- `[]Evolution` — 进化路径（有向图的边）
- `[]DialogueGroup` — 对话库
- `[]Adventure` — 冒险事件
- `map[string]Frame` — ASCII 动画帧
- `PluginSource` — 来源标记 (builtin/external)

### plugin.Registry

线程安全的物种注册中心：
- `LoadFromFS(fsys, root, source)` — 批量扫描加载
- `GetSpecies(id)` — 获取物种包
- `GetFrames(species, stage, anim)` — 获取动画帧（自动 fallback 到 idle）
- `GetDialogue(species, stage, mood)` — 随机获取对话
- `GetEvolutionsFrom(species, stage)` — 查询可用进化路径
- `GetAdventures(species, stage)` — 获取可用冒险

### game.Pet

宠物实体，包含：
- 基本信息 (Name, Species, Stage, StageID)
- 四项属性 (Hunger, Happiness, Health, Energy)
- 时间戳 (Birthday, LastFedAt, LastPlayedAt, LastCheckedAt)
- 统计数据 (TotalInteractions, GamesWon, etc.)
- 进化累积分 (AccHappiness, AccHealth, AccPlayful, Night/Day counts)
- 方法: `Feed()`, `Play()`, `Talk()`, `MoodScore()`, `MoodName()`, `UpdateAnimation()`

### store.JSONStore

JSON 文件持久化：
- 原子写入 (tmp → rename)
- 默认路径 `~/.local/share/clipet/save.json`
- 实现 `Store` 接口 (Save, Load, Exists)

## 初始化流程

```
main() → cli.NewRootCmd().Execute()
                │
                ├─ PersistentPreRunE: setup()
                │      │
                │      ├─ plugin.NewRegistry()
                │      ├─ registry.LoadFromFS(assets.BuiltinFS, "builtins", builtin)
                │      ├─ registry.LoadFromFS(os.DirFS(pluginsDir), ".", external)
                │      └─ store.NewJSONStore("")
                │
                ├─ 无子命令 → runTUI() → startTUI(pet, registry, store)
                └─ 有子命令 → init / status / feed / play
```

## 依赖图

```
cmd/clipet/main
  └── internal/cli
        ├── internal/assets (embed.FS)
        ├── internal/plugin
        │     └── github.com/BurntSushi/toml
        ├── internal/game
        ├── internal/store
        │     └── internal/game
        ├── internal/tui (待实现)
        │     ├── charm.land/bubbletea/v2
        │     ├── charm.land/lipgloss/v2
        │     ├── charm.land/bubbles/v2
        │     └── internal/game, internal/plugin
        └── github.com/spf13/cobra
```
