# Clipet 开发路线图

## 里程碑计划

### M1: 核心骨架 ✅

> 目标：可编译运行，CLI 快捷命令可用，TUI 基础交互

| 任务                          | 状态 |
|-------------------------------|------|
| 项目初始化 (go mod, git, deps) | ✅    |
| 插件系统 (types, parser, validator, loader, registry) | ✅ |
| 猫内置物种包 (species.toml, dialogues.toml, adventures.toml, 16 帧) | ✅ |
| assets/embed.go               | ✅    |
| 游戏逻辑 pet.go               | ✅    |
| 持久化 store/                  | ✅    |
| CLI 命令 (init, status, feed, play) | ✅ |
| CLI root.go                    | ✅    |
| TUI (app, home, petview, theme, bridge) | ✅ |
| 入口 main.go                  | ✅    |
| clipet-dev 开发者工具         | ✅    |
| 文档                           | ✅    |

### M2: 属性衰减与进化引擎 ✅

> 目标：属性随时间衰减，离线补偿计算，进化条件引擎，TUI 进化界面

| 任务                          | 状态 |
|-------------------------------|------|
| 进化引擎 game/evolution.go     | ✅    |
| 离线衰减 ApplyOfflineDecay     | ✅    |
| 辅助方法 GetAttr, UpdateFeedRegularity | ✅ |
| TUI 进化屏幕 screens/evolve.go | ✅   |
| TUI app.go 屏幕切换与集成      | ✅    |
| CLI 命令集成衰减+进化检查       | ✅    |
| CLI evolve_check.go 辅助       | ✅    |
| Makefile                       | ✅    |

### M2.5: 属性系统重构 ✅

> 目标：操作冷却、收益递减、前置条件，让宠物互动更真实

| 任务                          | 状态 |
|-------------------------------|------|
| ActionResult 返回值结构体      | ✅    |
| diminish 收益递减公式           | ✅    |
| 操作冷却 (Feed 10m, Play 5m, Rest 15m, Heal 20m, Talk 2m) | ✅ |
| 前置条件检查 (精力不足不能玩/治疗, 太饱不能喂, 精力充沛不能休息) | ✅ |
| TUI 警告消息样式 (红色边框)     | ✅    |
| CLI feed/play 适配新签名        | ✅    |

### M3: 迷你游戏 ✅

> 目标：内嵌小型游戏，胜负影响宠物属性

| 任务                          | 状态 |
|-------------------------------|------|
| 游戏管理器框架 games/manager.go | ✅    |
| 反应速度测试游戏 (⚡)          | ✅    |
| 猜数字游戏 (🎲)              | ✅    |
| TUI 集成到二级菜单            | ✅    |
| 游戏胜负 → 属性影响机制       | ✅    |

### M4: 冒险系统（计划中）

### M4: 冒险系统（计划中）

- 冒险事件触发机制
- 选项式交互
- 加权随机结果计算
- 冒险完成 → 属性影响

### M5: 对话系统 ✅

> 目标：丰富对话体验，自动闲聊，气泡界面

| 任务                          | 状态 |
|-------------------------------|------|
| 对话气泡组件 components/dialoguebubble.go | ✅    |
| 闲聊自动触发 (每3分钟30%概率) | ✅    |
| TUI 集成气泡显示在宠物上方   | ✅    |

### M3: 迷你游戏（计划中）

### M6: 高级特性（计划中）

- 外部插件热安装 `clipet plugin install <path>`
- 死亡与重生机制
- 成就系统
- 统计面板
- 多宠物存档

## 技术债与注意事项

### 已知问题

1. **外部格式化工具破坏文件**: 工作区配置的格式化工具会破坏 Go 源文件
   （表现为：行倒序压缩成单行，出现重复 package 声明）。
   修复方式：使用终端 `cat` heredoc 覆盖重写并 `gofmt -w` 验证。

### 编码规范

- 所有 Go 文件使用 `gofmt` 标准格式
- 内部包统一在 `internal/` 下
- 中文注释和中文 UI 文本
- TOML 用于外部可配置数据，JSON 用于存档
