# Clipet — 终端宠物伴侣 设计文档

## 项目概述

Clipet 是一个运行在终端中的宠物养成程序（TUI），以 ASCII 像素风动画为核心视觉表现，
支持喂食、心情、成长进化、迷你游戏、对话、数据持久化等完整交互系统。

### 项目定位

- **类型**: 终端桌宠 / TUI 养成游戏
- **风格**: 科幻 + 奇幻混搭（Fantasy Adventure）
- **运行模式**: 混合模式 — 全屏 TUI 交互 + CLI 快捷命令

## 技术栈

| 组件          | 选型                             | 版本     |
|---------------|----------------------------------|----------|
| 语言          | Go                               | 1.26+    |
| TUI 框架      | Bubble Tea v2                    | 2.0.0    |
| 样式          | Lipgloss v2                      | 2.0.0    |
| TUI 组件      | Bubbles v2                       | 2.0.0    |
| CLI 框架      | Cobra                            | 1.10.2   |
| 配置格式      | TOML (BurntSushi/toml)           | 1.6.0    |
| 动画缓动      | Harmonica                        | 0.2.0    |
| 日志          | charmbracelet/log                | 0.4.2    |

> **注意**: Bubble Tea v2 使用 `charm.land/bubbletea/v2` 导入路径，
> 不是 `github.com/charmbracelet/bubbletea/v2`。Lipgloss v2 和 Bubbles v2 同理。

## 核心特性

### 1. ASCII 像素风动画

- 多帧动画系统：idle、eating、sleeping、playing、happy、sad 等状态
- 帧文件存储在物种包 `frames/` 目录
- 命名规范：`{stageID}_{animState}_{index}.txt`
- 帧切换采用定时器驱动

### 2. 属性系统

四项核心属性（0-100）：

| 属性       | 说明                  | 影响                      |
|------------|----------------------|---------------------------|
| Hunger     | 饱腹度，越高表示越饱  | Feed 操作 +25             |
| Happiness  | 快乐度               | Play +20, Feed +5, Talk +5 |
| Health     | 健康值               | 初始由物种决定             |
| Energy     | 精力值               | Play -10                   |

### 3. 心情系统

心情由属性加权计算：
```
MoodScore = Hunger × 0.25 + Happiness × 0.35 + Health × 0.25 + Energy × 0.15
```

| 分数范围 | 心情名称   | 动画状态  |
|----------|-----------|-----------|
| 81-100   | happy     | AnimHappy |
| 61-80    | normal    | AnimIdle  |
| 41-60    | unhappy   | AnimIdle  |
| 21-40    | sad       | AnimSad   |
| 0-20     | miserable | AnimSad   |

### 4. 成长进化系统

五个生命阶段：

```
Egg → Baby → Child → Adult → Legend
        │       │       │       │
      1阶段   3分支   6分支   6分支
```

进化条件组合：
- `min_age_hours` — 最低年龄
- `attr_bias` — 属性偏好 (happiness/health/playful)
- `min_dialogues` / `min_adventures` — 互动次数
- `min_feed_regularity` — 喂食规律性
- `night_interactions_bias` / `day_interactions_bias` — 时段偏好
- `min_interactions` — 总互动次数
- `min_attr` — 属性门槛

### 5. 对话系统

- 按阶段 + 心情匹配对话内容
- 支持通配符匹配（`*` 匹配全部阶段/心情）
- 每次对话随机选取候选台词

### 6. 冒险系统

- 选项式冒险事件
- 按阶段过滤可用冒险
- 每个选项有加权随机结果
- 结果影响属性值

### 7. 持久化

- JSON 格式存档
- 原子写入（tmp + rename）
- 默认路径：`~/.local/share/clipet/save.json`

### 8. 插件系统

- 物种/进化树通过插件包加载
- 内置物种以 `go:embed` 方式打包
- 外部插件从 `~/.local/share/clipet/plugins/` 加载
- 统一的 `fs.FS` 加载接口
- TOML 描述格式

## 运行模式

### 全屏 TUI 模式

```bash
clipet          # 启动全屏 TUI 交互界面
```

### CLI 快捷命令

```bash
clipet init     # 创建新宠物
clipet status   # 查看宠物状态 (-j 输出 JSON)
clipet feed     # 快速喂食
clipet play     # 快速玩耍
```

## 数据路径

| 路径                                  | 用途             |
|---------------------------------------|-----------------|
| `~/.local/share/clipet/save.json`     | 宠物存档         |
| `~/.local/share/clipet/plugins/`      | 外部插件目录     |
