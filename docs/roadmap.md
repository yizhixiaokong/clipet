# Clipet 开发路线图

## 里程碑计划

### M1: 核心骨架 ✅

> 目标：可编译运行，CLI 快捷命令可用，TUI 基础交互

| 任务 | 状态 |
|------|------|
| 项目初始化 (go mod, git, deps) | ✅ |
| 插件系统 (types, parser, validator, loader, registry) | ✅ |
| 猫内置物种包 (species.toml, dialogues.toml, adventures.toml, 16 帧) | ✅ |
| assets/embed.go | ✅ |
| 游戏逻辑 pet.go | ✅ |
| 持久化 store/ | ✅ |
| CLI 命令 (init, status, feed, play) | ✅ |
| TUI (app, home, petview, theme, bridge) | ✅ |
| 入口 main.go | ✅ |
| clipet-dev 开发者工具 | ✅ |

### M2: 属性衰减与进化引擎 ✅

> 目标：属性随时间衰减，离线补偿计算，进化条件引擎，TUI 进化界面

| 任务 | 状态 |
|------|------|
| 进化引擎 game/evolution.go | ✅ |
| 离线衰减 ApplyOfflineDecay | ✅ |
| 辅助方法 GetAttr, UpdateFeedRegularity | ✅ |
| TUI 进化屏幕 screens/evolve.go | ✅ |
| TUI app.go 屏幕切换与集成 | ✅ |
| CLI 命令集成衰减+进化检查 | ✅ |

### M2.5: 属性系统重构 ✅

> 目标：操作冷却、收益递减、前置条件，让宠物互动更真实

| 任务 | 状态 |
|------|------|
| ActionResult 返回值结构体 | ✅ |
| diminish 收益递减公式 | ✅ |
| 操作冷却 (Feed 10m, Play 5m, Rest 15m, Heal 20m, Talk 2m) | ✅ |
| 前置条件检查 | ✅ |
| TUI 警告消息样式 | ✅ |
| CLI 适配新签名 | ✅ |

### M3: 迷你游戏 ✅

> 目标：内嵌小型游戏，胜负影响宠物属性

| 任务 | 状态 |
|------|------|
| MiniGame 接口设计 (纯状态机，无阻塞 I/O) | ✅ |
| 游戏管理器工厂 games/manager.go | ✅ |
| 反应速度测试 reaction.go (Start/HandleKey/Tick/View) | ✅ |
| 猜数字游戏 guess.go (键盘输入缓冲 + 猜测历史) | ✅ |
| TUI 游戏覆盖层 (全屏渲染 + Esc 退出) | ✅ |
| 属性影响：精力消耗(前置) + 快乐度奖惩(结算) | ✅ |

### M5: 对话系统 ✅

> 目标：丰富对话体验，自动闲聊，气泡界面

| 任务 | 状态 |
|------|------|
| 对话气泡组件 components/dialoguebubble.go | ✅ |
| 闲聊自动触发 (每3分钟30%概率，失败后1分钟重试) | ✅ |
| TUI 集成气泡显示在宠物上方 | ✅ |

### TUI 架构优化 ✅

> 代码审查后的整体重构

| 任务 | 状态 |
|------|------|
| 二级菜单 (分类标签 + 子操作) | ✅ |
| 游戏从阻塞 I/O 重写为 Bubble Tea 事件驱动状态机 | ✅ |
| 修复 gameResultHandler 值接收者修改丢失 | ✅ |
| 修复 PlayGame 三重属性修改 → 分离为消耗(前置)+奖惩(结算) | ✅ |
| 修复对话/消息双系统混乱 → 统一为 bubble + message | ✅ |
| 修复 t 快捷键缺少 return | ✅ |
| 进化屏幕添加 Esc 取消 | ✅ |
| 导出 Clamp、移除重复 clamp | ✅ |
| 移除 RenderAligned 死代码 | ✅ |
| 移除 lipgloss Copy() 弃用调用 | ✅ |
| 自动对话从 Update 移至 Tick（避免每次按键重复触发） | ✅ |
| 进化检查只在用户操作后触发（非游戏中） | ✅ |
| store.Save 错误提示到消息区 | ✅ |
| executeAction 重复代码提取为 failMsg/okMsg/infoMsg 辅助方法 | ✅ |
| 添加 GamePanel 主题样式 | ✅ |

### M4: 冒险系统（计划中）

> 目标：随机冒险事件，选项式交互，加权结果

- [ ] 冒险事件触发机制
- [ ] 选项式交互 (Bubble Tea 状态机)
- [ ] 加权随机结果计算
- [ ] 冒险完成 → 属性影响
- [ ] TUI 冒险覆盖层

### M6: 高级特性（计划中）

- [ ] 外部插件热安装 `clipet plugin install <path>`
- [ ] 死亡与重生机制
- [ ] 成就系统
- [ ] 统计面板
- [ ] 多宠物存档

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.25+ |
| TUI 框架 | Bubble Tea v2 (`charm.land/bubbletea/v2`) |
| 样式 | Lipgloss v2 (`charm.land/lipgloss/v2`) |
| CLI | Cobra |
| 配置 | TOML (物种/对话) + JSON (存档) |
| 资源嵌入 | go:embed |

## 架构原则

- **纯逻辑 / UI 分离**：`game/` 包不依赖任何 UI 框架
- **事件驱动**：所有 UI 交互通过 Bubble Tea 的 `Update(msg)` / `View()` 模式，禁止阻塞 I/O
- **值语义模型**：Bubble Tea model 使用值接收者；需修改状态时通过返回新值传递
- **插件化物种**：通过 TOML 配置定义物种、进化树、对话、冒险

## 编码规范

- 所有 Go 文件使用 `gofmt` 标准格式
- 内部包统一在 `internal/` 下
- 中文注释和中文 UI 文本
- TOML 用于外部可配置数据，JSON 用于存档
