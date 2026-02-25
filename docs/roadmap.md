# Clipet 开发路线图

## 里程碑计划

### M1: 核心骨架 ✅ / 🔧 进行中

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
| TUI 桥接 tui_bridge.go        | 🔧   |
| TUI 样式 styles/theme.go      | 🔧   |
| TUI 组件 petview.go           | 🔧   |
| TUI 主屏幕 home.go            | 🔧   |
| TUI 应用 app.go               | 🔧   |
| 入口 main.go                  | 🔧   |
| 编译测试                       | 🔧   |
| 文档                           | ✅    |
| Git 提交                       | 🔧   |

### M2: 属性衰减与进化引擎（计划中）

- 属性随时间自然衰减
- 离线时间补偿计算
- 进化条件判定引擎
- 自动/手动确认进化
- 进化时的特殊动画

### M3: 迷你游戏（计划中）

- 反应速度类游戏
- 猜数字 / 猜谜语类
- 游戏胜负 → 属性影响
- TUI 内嵌游戏界面

### M4: 冒险系统（计划中）

- 冒险事件触发机制
- 选项式交互
- 加权随机结果计算
- 冒险完成 → 属性影响

### M5: 对话系统（计划中）

- 按阶段 + 心情匹配对话
- TUI 对话气泡界面
- 对话触发频率控制

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
