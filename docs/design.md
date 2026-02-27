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
| 语言          | Go                               | 1.25+    |
| TUI 框架      | Bubble Tea v2                    | 2.0.0    |
| 样式          | Lipgloss v2                      | 2.0.0    |
| TUI 组件      | Bubbles v2                       | 2.0.0 (间接依赖) |
| CLI 框架      | Cobra                            | 1.10.2   |
| 配置格式      | TOML (BurntSushi/toml)           | 1.6.0    |
| 动画缓动      | Harmonica                        | 0.2.0 (间接依赖) |

> **注意**: Bubble Tea v2 使用 `charm.land/bubbletea/v2` 导入路径，
> 不是 `github.com/charmbracelet/bubbletea/v2`。Lipgloss v2 和 Bubbles v2 同理。

## 核心特性

### 1. ASCII 像素风动画

- 多帧动画系统：idle、eating、sleeping、playing、happy、sad 等状态
- 帧文件存储在物种包 `frames/` 目录，按阶段分级组织
- 推荐多级目录：`frames/{phase}/{variant}/idle.txt`（路径拼接为 stageID）
- 支持精灵图格式：`{animState}.txt`（多帧用 `---` 分隔）
- 兼容根级格式：`{stageID}_{animState}.txt` 和逐帧 `{stageID}_{animState}_{index}.txt`
- 帧切换采用定时器驱动

### 2. 属性系统

四项核心属性（0-100）：

| 属性       | 说明                  | 操作影响（base值）               |
|------------|----------------------|---------------------------------|
| Hunger     | 饱腹度，越高表示越饱  | Feed +25                        |
| Happiness  | 快乐度               | Play +20, Feed +5, Talk +5, Rest -5 |
| Health     | 健康值               | Heal +25, Rest +5               |
| Energy     | 精力值               | Rest +30, Play -10, Heal -15    |

#### 收益递减

所有增益操作使用递减公式，属性越接近满值增长越小：

```
gain = base × (100 - current) / 100   (最小为 1)
```

例如：Feed base=25，当前 Hunger=80 时实际增益 = 25×20/100 = 5。

#### 属性衰减（核心机制）

属性随**现实时间**自然衰减，即使程序未运行：

| 属性       | 衰减速率      | 说明                           |
|------------|--------------|--------------------------------|
| Hunger     | -3 / 小时    | 最快衰减                       |
| Happiness  | -2 / 小时    | 中速衰减                       |
| Energy     | -1 / 小时    | 最慢衰减                       |
| Health     | -0.5 / 小时  | 仅在 Hunger < 20 时触发        |

离线补偿：程序启动时计算 `time.Since(LastCheckedAt)` 并一次性扣除衰减量。

#### 死亡机制

- **触发条件**: Health 降至 0
- **永久性**: 当前版本死亡不可逆，无复活机制
- **行为**: 死亡后所有操作返回 "宠物已经不在了..."
- **典型死亡路径**: 长期不喂食 → 饱腹归零 → 健康持续下降 → 健康归零 → 死亡
- 从满属性到饿死大约需要 **~40 小时**持续忽视

#### 操作冷却

每个操作有最低间隔时间，防止无限刷属性：

| 操作    | 冷却时间 | 前置条件       |
|---------|---------|----------------|
| Feed    | 10 分钟 | Hunger < 95    |
| Play    | 5 分钟  | Energy ≥ 10    |
| Talk    | 2 分钟  | —              |
| Rest    | 15 分钟 | Energy < 90    |
| Heal    | 20 分钟 | Energy ≥ 10    |
| 冒险    | 10 分钟 | Energy ≥ 15    |

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

#### 自动闲聊

宠物会在 TUI 界面主动说话：
- 每 3 分钟有 30% 概率触发自动对话气泡
- 失败后 1 分钟重试
- 匹配当前阶段 + 心情的对话库
- 在游戏/冒险进行时不会触发

### 6. 冒险系统

- 选项式冒险事件，4 阶段流程：介绍 → 选择 → 动画 → 结果
- 按阶段过滤可用冒险（支持通配符如 `child_*`）
- 每个选项有加权随机结果（`weight` 字段控制概率）
- 结果影响属性值（显示彩色 +/- 变化）
- 固定消耗 10 精力，冷却 10 分钟
- 完成次数 `AdventuresCompleted` 计入进化条件

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
