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

### M4: 冒险系统 ✅

> 目标：随机冒险事件，选项式交互，加权结果

| 任务 | 状态 |
|------|------|
| 冒险引擎 game/adventure.go (CanAdventure, PickAdventure, ResolveOutcome) | ✅ |
| 加权随机结果计算 + 属性影响 (ApplyAdventureOutcome) | ✅ |
| TUI 冒险屏幕 screens/adventure.go (4阶段: 介绍→选择→动画→结果) | ✅ |
| App 冒险屏幕切换 (screenAdventure + 完整生命周期) | ✅ |
| Home 菜单集成冒险入口 (🗺️ 冒险) | ✅ |
| Pet 冒险冷却字段 LastAdventureAt (10分钟冷却) | ✅ |

### M7: 插件系统重构 ✅

> 目标：支持生命周期、个性特征、自定义属性和终局系统

#### Phase 1: 核心边界定义 ✅

| 任务 | 状态 |
|------|------|
| 定义能力类型和接口 (capabilities/types.go) | ✅ |
| 创建能力注册表 (capabilities/registry.go) | ✅ |
| 扩展插件类型支持生命周期和特征 | ✅ |
| 解析新配置段 (lifecycle, traits, endings) | ✅ |
| 添加生命周期追踪到 Pet 模型 | ✅ |

#### Phase 2: 生命周期系统 ✅

| 任务 | 状态 |
|------|------|
| 创建生命周期管理器 (lifecycle_manager.go) | ✅ |
| 创建生命周期时间钩子 (hook_lifecycle.go) | ✅ |
| 集成生命周期检查到时间系统 | ✅ |
| 实现终局触发和预警机制 | ✅ |

#### Phase 3: 灵活属性系统 ✅

| 任务 | 状态 |
|------|------|
| 实现属性系统接口 (attributes/system.go) | ✅ |
| 添加自定义属性支持到 Pet 模型 | ✅ |
| 统一属性访问接口 (GetAttr/SetAttr) | ✅ |

#### Phase 4: 猫插件示例重构 ✅

| 任务 | 状态 |
|------|------|
| 添加生命周期配置 (10天寿命) | ✅ |
| 定义个性特征 (九条命、呼噜治愈、夜猫子、挑食) | ✅ |
| 定义多种终局 (幸福终老、冒险一生、平静休息) | ✅ |
| 更新插件开发指南文档 | ✅ |

#### Phase 5: 多风格生命周期支持 ✅

| 任务 | 状态 |
|------|------|
| 扩展 LifecycleState (IsEternal, IsLooping) | ✅ |
| 实现 eternal 生命周期逻辑（永不死亡） | ✅ |
| 实现 loop 生命周期逻辑（循环重生） | ✅ |
| 修改时间钩子处理不同生命周期类型 | ✅ |
| 创建生命周期风格示例配置 | ✅ |

#### Phase 6: 插件生态安全设计 ✅

| 任务 | 状态 |
|------|------|
| 创建约束系统 (capabilities/constraints.go) | ✅ |
| 添加生命周期边界验证 | ✅ |
| 添加属性修正器限制 | ✅ |
| 增强插件验证器 | ✅ |
| 实现默认回退机制 | ✅ |
| 文档化约束和覆盖机制 | ✅ |

#### Phase 7: 动作系统插件化 ✅

| 任务 | 状态 |
|------|------|
| 添加 ActionConfig 和 ActionEffects 类型 (plugin/types.go) | ✅ |
| 添加 DecayConfig 和 DynamicCooldownConfig 类型 (capabilities/types.go) | ✅ |
| 扩展 Registry 支持获取动作、衰减和动态冷却配置 | ✅ |
| 创建 action_config.go 辅助函数 (GetActionCooldown, GetActionEffects, CalculateDynamicCooldown) | ✅ |
| 修改 AttrDecayHook 使用插件控制的衰减率 | ✅ |
| 为猫插件添加完整动作配置 (feed, play, rest, heal, talk) | ✅ |
| 为猫插件添加衰减配置 (统一慢速衰减) | ✅ |
| 为猫插件添加动态冷却配置 (基于紧急度的动态冷却) | ✅ |
| 完善 plugin-guide.md 最佳实践文档 | ✅ |
| 文档化离线游戏交互节奏设计 | ✅ |

#### 向后兼容性 ✅

- 旧插件自动使用默认生命周期配置
- 新旧系统并存，不破坏现有功能
- 所有现有测试通过
- 手动测试验证（timeskip、evolve、adventure）

---

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
- **可扩展能力系统**：核心提供能力接口，插件定义具体实现

## 编码规范

- 所有 Go 文件使用 `gofmt` 标准格式
- 内部包统一在 `internal/` 下
- 中文注释和中文 UI 文本
- TOML 用于外部可配置数据，JSON 用于存档

---

## M8: 统计面板（未来）

> 目标：宠物成长历史、交互统计可视化

| 任务 | 状态 |
|------|------|
| 交互历史记录系统 | ⏳ |
| 统计数据可视化 | ⏳ |
| TUI 统计屏幕 | ⏳ |

---

## 未来扩展方向

- **脚本系统**：Lua/WASM 支持，允许复杂逻辑
- **外部插件管理**：`clipet plugin install/list/remove`
- **多宠物存档**：支持同时养成多只宠物
- **云同步**：存档跨设备同步
- **社区插件仓库**：共享物种包生态
