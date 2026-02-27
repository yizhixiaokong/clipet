# Clipet 插件开发指南

## 概览

Clipet 的物种系统完全基于插件包。每个物种是一个独立的目录，包含 TOML 配置文件
和 ASCII 动画帧文件。内置物种和外部插件使用完全相同的格式和加载路径。

## 插件包目录结构

```
my-species-pack/
├── species.toml        # 必须 — 物种定义 + 进化树
├── dialogues.toml      # 可选 — 对话库
├── adventures.toml     # 可选 — 冒险事件
└── frames/             # 可选 — ASCII 动画帧
    ├── egg/              # 按阶段分级目录
    │   └── idle.txt
    ├── baby/
    │   └── xxx/
    │       ├── idle.txt
    │       ├── eating.txt
    │       └── ...
    ├── child/
    │   ├── variant_a/
    │   └── variant_b/
    ├── adult/
    │   └── .../
    └── legend/
        └── .../
```

## species.toml

### 物种元信息

```toml
[species]
id = "dragon"                    # 必须，唯一标识符
name = "龙"                      # 必须，显示名称
description = "远古巨龙" # 描述文字
author = "your-name"             # 作者
version = "1.0.0"                # 必须，语义化版本

[species.base_stats]             # 初始属性值 (0-100)
hunger = 50
happiness = 60
health = 70
energy = 65
```

### 阶段定义

每个阶段是进化树上的一个节点：

```toml
[[stages]]
id = "egg"           # 阶段唯一 ID
name = "龙之蛋"      # 显示名称
phase = "egg"        # 生命阶段: egg | baby | child | adult | legend
```

### 生命周期配置 (v2.0+)

定义宠物的生命周期参数和终局类型：

```toml
[lifecycle]
max_age_hours = 240.0       # 最大寿命（小时），默认 720 小时（30 天）
ending_type = "death"       # 终局类型: death | ascend | eternal
warning_threshold = 0.75    # 预警阈值（0.0-1.0），达到时显示温馨提示
```

**生命周期示例**：

```toml
# 短寿命物种（10天）
[lifecycle]
max_age_hours = 240.0
ending_type = "death"
warning_threshold = 0.75

# 长寿命物种（60天）
[lifecycle]
max_age_hours = 1440.0
ending_type = "ascend"
warning_threshold = 0.9
```

**Phase 5 新增 - 多风格生命周期**：

```toml
# 7天高风险（挑战模式）
[lifecycle]
max_age_hours = 168.0        # 7 天
ending_type = "death"
warning_threshold = 0.6      # 第 4 天预警

# 30天伴侣（长期陪伴）
[lifecycle]
max_age_hours = 720.0        # 30 天
ending_type = "death"
warning_threshold = 0.85     # 第 25 天预警

# 永恒物种（永不老死）
[lifecycle]
max_age_hours = 0.0          # 忽略
ending_type = "eternal"      # 永不老死
warning_threshold = 0.0      # 无预警

# 循环重生物种
[lifecycle]
max_age_hours = 240.0        # 10 天一个周期
ending_type = "loop"         # 10 天后重置年龄
warning_threshold = 0.8      # 每个周期末预警
```

**生命周期类型说明**：

- `death`: 自然离世（默认）
- `ascend`: 飞升/升华（温馨主题）
- `eternal`: 永恒存在，永不触发终局
- `loop`: 循环重生，达到寿命后重置年龄而非死亡

**向后兼容**：旧插件不包含 `[lifecycle]` 段时，使用默认值（30 天寿命，death 终局）。

### 个性特征定义 (v2.0+)

定义物种的个性特征（非战斗能力），分为三类：被动特征、主动技能和进化修正器。

**被动特征** (passive) - 自动生效的特征：

```toml
[[traits]]
id = "picky_eater"
name = "挑食"
description = "对食物比较挑剔，但喂食时心情更好"
type = "passive"
[traits.passive_effect]
feed_hunger_bonus = -0.2     # 喂食饱食度 -20%
feed_happiness_bonus = 0.1   # 但快乐度 +10%
```

**主动技能** (active) - 玩家可触发的技能：

```toml
[[traits]]
id = "purr_heal"
name = "呼噜治愈"
description = "消耗精力通过呼噜治愈自己"
type = "active"
[traits.active_effect]
energy_cost = 10             # 消耗 10 点精力
health_restore = 15          # 恢复 15 点健康
cooldown = "30m"             # 30 分钟冷却
```

**进化修正器** (modifier) - 影响进化点数积累：

```toml
[[traits]]
id = "night_owl"
name = "夜猫子"
description = "晚上更活跃"
type = "modifier"
[traits.evolution_modifier]
night_interaction_bonus = 1.5  # 夜间互动进化点数 +50%
```

**可用字段**：

| 字段 | 类型 | 说明 |
|-----|------|------|
| `feed_hunger_bonus` | float | 喂食饱食度增益倍率（-1.0 到 1.0）|
| `feed_happiness_bonus` | float | 喂食快乐度增益倍率 |
| `play_happiness_bonus` | float | 玩耍快乐度增益倍率 |
| `sleep_energy_bonus` | float | 休息精力增益倍率 |
| `resurrect_chance` | float | 死亡时复活概率（0.0-1.0）|
| `health_restore_percent` | float | 复活时恢复生命值百分比 |
| `energy_cost` | int | 主动技能精力消耗 |
| `health_restore` | int | 主动技能健康恢复量 |
| `hunger_restore` | int | 主动技能饱食度恢复量 |
| `happiness_boost` | int | 主动技能快乐度提升量 |
| `cooldown` | duration | 主动技能冷却时间（如 "30m", "1h"）|
| `night_interaction_bonus` | float | 夜间互动进化点数倍率 |
| `day_interaction_bonus` | float | 日间互动进化点数倍率 |
| `feed_bonus` | float | 喂食进化点数倍率 |
| `play_bonus` | float | 玩耍进化点数倍率 |
| `adventure_bonus` | float | 冒险进化点数倍率 |

### 终局定义 (v2.0+)

定义物种的多种可能终局，基于宠物的生命周期质量：

```toml
# 幸福终老 - 高快乐度 + 长寿
[[endings]]
type = "blissful_passing"
name = "幸福终老"
message = "带着满足的笑容，你的猫咪安详地离开了..."
[endings.condition]
min_happiness = 80
min_age_hours = 200.0

# 冒险一生 - 完成多次冒险
[[endings]]
type = "adventurous_life"
name = "冒险一生"
message = "它度过了充满冒险的一生，成为了传奇..."
[endings.condition]
min_adventures = 30

# 平静休息 - 默认终局
[[endings]]
type = "peaceful_rest"
name = "平静休息"
message = "平静地度过了这一生，它已经离开了..."
[endings.condition]
```

**终局条件字段**：

| 字段 | 类型 | 说明 |
|-----|------|------|
| `min_happiness` | int | 最低心情分数（0-100）|
| `min_age_hours` | float | 最低存活时间（小时）|
| `min_adventures` | int | 最少完成冒险次数 |

终局按定义顺序匹配，第一个满足条件的终局将被触发。建议将特定条件高的终局放在前面，默认终局（空条件）放在最后。

### 自定义属性 (v2.0+，可选)

定义物种特有的自定义属性（扩展核心四属性）：

```toml
[[attributes]]
id = "magic"
display_name = "魔力"
min = 0
max = 100
default = 50
decay_rate = 0.5  # 每小时衰减 0.5 点
```

自定义属性会与核心属性（饥饿、快乐、健康、精力）一起管理，支持：
- 属性衰减（基于 decay_rate）
- 个性化特征效果
- 进化条件检查

### 动作配置 (v3.0+, Phase 7)

定义物种的动作行为，包括冷却时间和效果数值。**这是插件化设计的核心，让每个物种有独特的互动节奏。**

```toml
# 喂食动作
[[actions]]
id = "feed"
cooldown = "10m"              # 冷却时间
[actions.effects]
hunger = 25                   # 饱食度 +25
happiness = 5                 # 快乐度 +5

# 玩耍动作
[[actions]]
id = "play"
cooldown = "5m"
energy_cost = 10              # 需要消耗 10 点精力才能执行
[actions.effects]
happiness = 20
energy = -10                  # 精力 -10

# 休息动作
[[actions]]
id = "rest"
cooldown = "15m"
[actions.effects]
energy = 30                   # 精力 +30
health = 5                    # 健康 +5
happiness = -5                # 快乐 -5（休息可能无聊）

# 治疗动作
[[actions]]
id = "heal"
cooldown = "20m"
energy_cost = 15
[actions.effects]
health = 25

# 对话动作
[[actions]]
id = "talk"
cooldown = "2m"
[actions.effects]
happiness = 5
```

**动作字段说明**：

| 字段 | 类型 | 必需 | 说明 |
|-----|------|------|------|
| `id` | string | 是 | 动作ID，必须是 feed/play/rest/heal/talk 之一 |
| `cooldown` | duration | 是 | 冷却时间（如 "10m", "1h30m"）|
| `energy_cost` | int | 否 | 执行所需精力（不消耗则不设置）|
| `effects.hunger` | int | 否 | 饱食度变化（正数为增加）|
| `effects.happiness` | int | 否 | 快乐度变化 |
| `effects.health` | int | 否 | 健康度变化 |
| `effects.energy` | int | 否 | 精力变化（可为负数）|

**向后兼容**：如果物种包未定义 actions，系统使用内置默认值。

**设计理念**：

1. **物种差异**：不同物种有不同的互动节奏
   - 猫：快速恢复，适合频繁互动
   - 龙：恢复慢，但单次效果强
   - 机器宠物：精力恢复快，但需要频繁维护

2. **平衡性考虑**：详见"最佳实践"部分

## dialogues.toml

### 基本格式

每个对话组包含阶段匹配、心情匹配和多条备选对话：

```toml
[[dialogues]]
stage = ["baby_dragon", "child_dragon"]  # 阶段匹配（支持通配符）
mood = ["happy", "normal"]                # 心情匹配（支持 "*" 匹配全部）
lines = [
  "你好呀！我是小龙~",
  "今天天气不错喵~",
  "我们一起玩吧！",
]
```

### 对话设计建议

**对话复杂度应与阶段匹配：**

| 生命阶段 | 建议对话内容 | 示例 |
|---------|-------------|------|
| **egg** | 仅简单声响 | `"咔嗒..."`, `"咚..."`, `"蛋壳轻响"` |
| **baby** | 简单音节，重复性高 | `"喵~"`, `"喵喵~"`, `"呜..."` |
| **child** | 简单短句，语气幼稚 | `"我们一起玩吧！"`, `"我好开心~"` |
| **adult** | 完整句子，表达自己的想法 | `"我需要一些休息时间。"`, `"感谢你的照顾。"` |
| **legend** | 深刻，富有哲理，有格局 | `"守护众生是我的使命。"` |

**对话状态反馈设计：**

- **positive (happy)**: 积极、活跃、期待互动
  ```toml
  [[dialogues]]
  stage = ["adult_dragon"]
  mood = ["happy"]
  lines = [
    "今天的阳光真美好！",
    "能和你在一起太开心了！",
    "让我们来一场冒险吧！"
  ]
  ```

- **neutral (normal)**: 平静、日常、温和表达
  ```toml
  [[dialogues]]
  stage = ["child_dragon"]
  mood = ["normal"]
  lines = [
    "嗯...今天还好。",
    "想吃点东西...",
    "要不要休息一下？"
  ]
  ```

- **negative (unhappy/sad/miserable)**: 消极、疲惫、寻求安慰
  ```toml
  [[dialogues]]
  stage = ["baby_dragon"]
  mood = ["sad"]
  lines = [
    "呜...呜呜...",
    "我好难过...",
    "快来安慰我..."
  ]
  ```

**阶段过渡建议：**

- **egg → baby**: 对话突然变得有语义
  ```toml
  # egg 阶段
  [[dialogues]]
  stage = ["egg"]
  mood = ["*"]
  lines = ["咚咚...", "咔嗒..."]

  # baby 阶段
  [[dialogues]]
  stage = ["baby_dragon"]
  mood = ["happy"]
  lines = ["哇！我出生了！", "这个世界真漂亮！"]
  ```

- **baby → child**: 开始表达复杂需求
  ```toml
  [[dialogues]]
  stage = ["child_dragon"]
  mood = ["normal"]
  lines = ["我想变得更强大！", "你能教我新技能吗？"]
  ```

- **child → adult**: 开始关心照顾者
  ```toml
  [[dialogues]]
  stage = ["adult_dragon"]
  mood = ["normal"]
  lines = ["谢谢你一直照顾我。", "我能为你做些什么呢？"]
  ```

### 阶段匹配规则

- `"*"` — 匹配所有阶段
- `"child_*"` — 前缀通配，匹配所有 `child_` 开头的阶段
- `"baby_dragon"` — 精确匹配

### 心情值

| 心情名称    | 心情分数范围 |
|------------|-------------|
| happy      | 81-100      |
| normal     | 61-80       |
| unhappy    | 41-60       |
| sad        | 21-40       |
| miserable  | 0-20        |

## adventures.toml

```toml
[[adventures]]
id = "treasure_cave"
name = "宝藏洞窟"
stage = ["child_*", "adult_*"]  # 可用阶段
description = "你发现了一个闪闪发光的洞穴入口......"

[[adventures.choices]]
text = "勇敢闯入"

[[adventures.choices.outcomes]]
weight = 60                     # 权重，用于加权随机
text = "发现了一堆美味的食物！"
[adventures.choices.outcomes.effects]
hunger = 20                     # 属性变化量
happiness = 10

[[adventures.choices.outcomes]]
weight = 40
text = "洞穴里空空如也。"
[adventures.choices.outcomes.effects]
happiness = -5

[[adventures.choices]]
text = "绕道而行"

[[adventures.choices.outcomes]]
weight = 100
text = "安全地离开了。"
[adventures.choices.outcomes.effects]
energy = 5
```

## decay 配置 (v3.0+, Phase 7)

定义物种的属性衰减率。Clipet 是一个**大多数时间离线**的宠物养成游戏，因此采用**统一慢速衰减**设计。

```toml
[decay]
hunger = 1.0        # 饱食度每小时衰减 1 点
happiness = 0.5     # 快乐度每小时衰减 0.5 点
energy = 0.3        # 精力每小时衰减 0.3 点
health = 0.2        # 饥饿时健康每小时衰减 0.2 点
```

**设计原则**：

1. **统一慢速衰减**：所有物种采用统一的慢速衰减率，无论在线/离线
2. **可调整性**：不同物种可以有不同的衰减率（例如：高能量物种精力衰减慢）
3. **离线友好**：衰减率设计为即使离线数天，宠物状态仍可控

**推荐衰减率范围**：

| 属性 | 推荐范围 | 说明 |
|-----|---------|------|
| `hunger` | 0.5 - 2.0 | 饱食度衰减（主要关注点）|
| `happiness` | 0.3 - 1.0 | 快乐度衰减 |
| `energy` | 0.2 - 0.5 | 精力衰减 |
| `health` | 0.1 - 0.5 | 健康衰减（仅在饥饿时）|

**向后兼容**：旧插件不包含 `[decay]` 段时，使用默认值（hunger=1.0, happiness=0.5, energy=0.3, health=0.2）。

## dynamic_cooldown 配置 (v3.0+, Phase 7)

定义动态冷却系统，根据属性紧急度自动调整冷却时间。解决离线游戏交互节奏问题。

```toml
[dynamic_cooldown]
# 低紧急度（属性 < 30）：非常短的冷却
low_urgency_multiplier = 0.1    # 10% 基础冷却
low_threshold = 30

# 中等紧急度（30 <= 属性 < 70）：中等冷却
medium_urgency_multiplier = 0.5  # 50% 基础冷却

# 高紧急度（属性 >= 70）：正常冷却
high_urgency_multiplier = 1.0    # 100% 基础冷却
high_threshold = 70
```

**动态冷却示例**：

假设基础喂食冷却为 10 分钟：

- 饱食度 = 10（非常低） → 冷却 = 10m × 0.1 = **1 分钟**
- 饱食度 = 50（中等） → 冷却 = 10m × 0.5 = **5 分钟**
- 饱食度 = 85（较高） → 冷却 = 10m × 1.0 = **10 分钟**

**设计理念**：

1. **紧急情况帮助**：属性低时，玩家获得额外帮助（短冷却），避免死循环
2. **收益递减**：属性健康时，冷却正常，鼓励多元化操作
3. **离线友好**：玩家短暂上线时，可以快速处理紧急状态

**平衡性建议**：

- `low_urgency_multiplier`: 0.05 - 0.2（5%-20% 冷却）
- `medium_urgency_multiplier`: 0.4 - 0.6（40%-60% 冷却）
- `high_urgency_multiplier`: 0.8 - 1.2（80%-120% 冷却）

**向后兼容**：旧插件不包含 `[dynamic_cooldown]` 段时，使用默认值（0.1/0.5/1.0，阈值 30/70）。

## 动画帧文件

支持三种目录布局，推荐使用多级子目录 + 精灵图格式。

### 目录布局

**布局一：多级子目录（推荐）**

```
frames/
  {phase}/
    {variant}/
      {animState}.txt     # 精灵图
      {animState}_{index}.txt  # 逐帧（兼容）
```

路径中各级目录名用 `_` 拼接还原为 stageID。
例如 `frames/adult/arcane_shadow/idle.txt` → stageID = `adult_arcane_shadow`。
单级目录也有效：`frames/egg/idle.txt` → stageID = `egg`。

**布局二：根级精灵图**

```
frames/{stageID}_{animState}.txt
```

**布局三：根级逐帧文件（兼容）**

```
frames/{stageID}_{animState}_{index}.txt
```

> **优先级**：多级子目录 > 根级精灵图 > 根级逐帧。同 (stageID, animState) 内精灵图优先于逐帧。

### 精灵图格式

同一动作的多帧放在一个文件中，用 `---` 行分隔：

```
 /-/\
(' ' )
 | \/
U-U(_/
---
 \-/\
(' ' )
 | \/
U-U(_/
```

**优势**：文件数少，同动作帧集中管理，编辑对比方便。

### 支持的动画状态

| animState  | 触发条件              |
|------------|----------------------|
| idle       | 默认状态 (必须提供)   |
| eating     | 喂食时               |
| sleeping   | 精力低于 15          |
| playing    | 玩耍时               |
| happy      | 心情 > 80            |
| sad        | 心情 < 40            |

如果某状态没有帧文件，会自动 fallback 到 `idle`。

### 帧内容规范

纯 ASCII 文本。建议每帧保持一致的宽高（推荐 12-16 字符宽，6-10 行高）。
系统会自动计算最大宽度并将所有帧补齐到相同宽度，避免渲染偏移。

## 安装外部插件

将插件目录放入 `~/.local/share/clipet/plugins/`：

```
~/.local/share/clipet/plugins/
└── my-dragon-pack/
    ├── species.toml
    ├── dialogues.toml
    ├── adventures.toml
    └── frames/
        ├── egg/
        │   └── idle.txt
        ├── baby/
        │   └── dragon/
        └── ...
```

程序启动时会自动扫描该目录并加载所有有效的物种包。

## 验证规则

加载时自动执行以下校验：

1. **必填字段**: `species.id`, `species.name`, `species.version`
2. **阶段完整性**: 至少一个 egg 阶段
3. **进化路径有效性**: from/to 引用的阶段 ID 必须存在
4. **进化链连通性**: 所有非 egg 阶段必须从某个 egg 阶段可达
5. **对话引用**: 非通配符的 stage 引用必须指向已定义的阶段
6. **冒险结构**: 每个冒险至少有一个选项，每个选项至少有一个结果
7. **帧文件**: egg 阶段必须有 idle 帧

校验失败时，整个插件包将被拒绝加载，并输出详细的错误信息列表。

## Phase 6: 安全约束系统

为防止插件滥用（过短寿命、极端数值、过多危机），系统实施轻量级安全边界。

### 默认约束

| 约束类型 | 默认值 | 说明 |
|---------|--------|------|
| 最小寿命 | 24 小时 | 防止瞬间死亡（eternal 除外）|
| 最大寿命 | 10 年 | 防止永不死亡（eternal 除外）|
| 属性修正器 | 10% - 300% | 防止极端增益/惩罚 |
| 冒险数量上限 | 20 | 防止玩家 overwhelmed |
| 对话组数量上限 | 100 | 防止内容过载 |

### 约束验证

**生命周期验证**：

```toml
# 错误：寿命过短（< 24 小时）
[lifecycle]
max_age_hours = 0.5  # 验证失败

# 错误：寿命过长（> 10 年）
[lifecycle]
max_age_hours = 100000.0  # 验证失败

# 正确：eternal 类型不受寿命限制
[lifecycle]
ending_type = "eternal"  # 验证通过
```

**属性修正器验证**：

```toml
# 错误：修正器过大（> 300%）
[[traits]]
[traits.passive_effect]
feed_hunger_bonus = 5.0  # 600% 增益，验证失败

# 错误：修正器过小（< 10%）
[traits.passive_effect]
play_happiness_bonus = -0.95  # 5% 增益，验证失败

# 正确：合理范围
[traits.passive_effect]
feed_hunger_bonus = -0.2  # 80% 增益，验证通过
```

### 约束覆盖机制

如需覆盖默认约束，必须提供理由说明（至少 50 字符）：

```toml
# 物种级别约束覆盖
[species]
id = "challenge_beetle"
name = "挑战甲虫"

[constraints]
min_lifespan_hours = 1.0      # 覆盖：1 小时寿命
reason = "Roguelike 挑战模式：玩家需在 1 小时内完成目标，体验紧张刺激的极限养成"
```

**验证规则**：
- 覆盖默认约束时，`reason` 字段必须存在且至少 50 字符
- 系统会验证 `reason` 字段的完整性
- 无覆盖时不需要 `reason` 字段

### 默认回退

如果插件配置超出安全边界且未提供约束覆盖，系统会自动钳制到安全范围：

```go
// 解析时自动钳制过短寿命
if max_age_hours < 24.0 && ending_type != "eternal" {
    max_age_hours = 24.0  // 回退到最小安全值
    // 输出警告日志
}
```

**建议**：对于特殊玩法需求，显式声明约束覆盖并提供理由，而不是依赖自动钳制。


## 参考：内置猫物种包

查看 `internal/assets/builtins/cat-pack/` 目录获取完整的参考实现。

进化树概览：

```
egg (神秘之蛋)
 └── baby (小猫咪)
      ├── child_arcane (咒术小猫)  ← happiness偏好 + 对话数
      │    ├── adult_arcane_shadow (暗影魅猫)  ← 夜间偏好
      │    │    └── legend_arcane_shadow (虚空行者)
      │    └── adult_arcane_crystal (水晶预言猫)  ← 日间偏好
      │         └── legend_arcane_crystal (星辰贤者)
      ├── child_feral (战斗小猫)  ← health偏好 + 冒险数
      │    ├── adult_feral_flame (烈焰狮)  ← hunger偏好
      │    │    └── legend_feral_flame (不灭炎帝)
      │    └── adult_feral_frost (霜暴豹)  ← energy偏好
      │         └── legend_feral_frost (极寒霜神)
      └── child_mech (机甲小猫)  ← playful偏好 + 喂食规律
           ├── adult_mech_cyber (赛博猞猁)  ← 对话数
           │    └── legend_mech_cyber (量子幽灵)
           └── adult_mech_chrome (合金猎豹)  ← 冒险数
                └── legend_mech_chrome (星际掠夺者)
```

## 已知问题

- 蛋阶段设计为简单的声音反馈，符合生命初期不能说话的设定
- 幼年阶段使用简单重复的音节，模拟幼崽学习语言的过程
- 高级阶段的对话会变得更完整、更有深度

---

# 最佳实践与设计指南

## 离线游戏交互节奏设计

### 核心挑战

Clipet 是一个**大多数时候离线**的宠物养成游戏。玩家通常只在打开界面的短时间内进行互动。这带来了独特的设计挑战：

**问题场景**：
```
1. 玩家打开游戏，发现宠物饱腹度只有 20
2. 喂食一次，恢复到 45，还远远不够
3. 但有 10 分钟冷却，无法立即再次喂食
4. 玩家只能等待或离开游戏，体验不流畅
```

### 设计原则

#### 1. **单次操作要有意义**

```toml
# ❌ 错误：效果太弱，单次操作无意义
[[actions]]
id = "feed"
cooldown = "10m"
[actions.effects]
hunger = 10   # 太少了！

# ✅ 正确：单次操作有明显的效果
[[actions]]
id = "feed"
cooldown = "10m"
[actions.effects]
hunger = 35   # 从 20 → 55，明显改善
happiness = 5
```

**建议**：单次操作应该能恢复 30-40% 的属性，让玩家有成就感。

#### 2. **冷却时间要合理**

```toml
# 不同物种的冷却节奏示例

# 快节奏物种（猫）
[[actions]]
id = "feed"
cooldown = "5m"    # 短冷却，适合频繁互动
[actions.effects]
hunger = 25

# 慢节奏物种（龙）
[[actions]]
id = "feed"
cooldown = "30m"   # 长冷却，但单次效果强
[actions.effects]
hunger = 60
```

**建议**：
- 快节奏物种：5-10 分钟冷却
- 慢节奏物种：15-30 分钟冷却
- 冷却越长，单次效果应该越强

#### 3. **动态冷却系统（推荐）**

Clipet 使用**动态冷却**来平衡离线游戏体验。冷却时间根据属性紧急度自动调整：

```toml
[dynamic_cooldown]
# 当饱食度很低（< 30）时，喂食冷却极短
low_urgency_multiplier = 0.1    # 10% 冷却
low_threshold = 30

# 当饱食度中等（30-70）时，冷却适中
medium_urgency_multiplier = 0.5  # 50% 冷却

# 当饱食度较高（>= 70）时，冷却正常
high_urgency_multiplier = 1.0    # 100% 冷却
high_threshold = 70
```

**示例**（基础喂食冷却 = 10 分钟）：

| 饱食度 | 冷却时间 | 设计意图 |
|-------|---------|---------|
| 10 | **1 分钟** | 紧急情况，玩家可快速喂食多次 |
| 40 | **5 分钟** | 中等紧急，节奏适中 |
| 85 | **10 分钟** | 状态良好，正常冷却 |

**设计理由**：
- **避免死循环**：属性低时不会陷入"无法恢复"的困境
- **离线友好**：短暂上线即可处理紧急状态
- **收益递减**：属性健康时冷却正常，鼓励多元化操作

#### 4. **多元化恢复路径**

```toml
# 不要让玩家只能等待冷却

# 主动技能：消耗资源快速恢复
[[traits]]
id = "emergency_feed"
name = "紧急喂食"
description = "消耗双倍精力，无视冷却喂食"
type = "active"
[traits.active_effect]
energy_cost = 20      # 比正常高
hunger_restore = 30
cooldown = "1h"       # 长冷却，但提供选择
```

### 具体数值建议

**基础冷却时间**（在 `[[actions]]` 中配置）：

| 物种类型 | 基础冷却 | 单次效果 | 特点 |
|---------|---------|---------|------|
| 快节奏（猫、兔子）| 5-10m | 25-35 | 频繁互动，快速反馈 |
| 平衡型（狗、狐狸）| 10-15m | 30-40 | 标准节奏 |
| 慢节奏（龙、凤凰）| 20-30m | 50-70 | 少量但强力 |
| 神话级 | 30-60m | 70-90 | 稀有但强力 |

**实际冷却时间**（动态冷却系统自动调整）：

例如：基础冷却 = 10 分钟，饱食度 = 15（紧急）

```toml
[dynamic_cooldown]
low_urgency_multiplier = 0.1
low_threshold = 30

# 实际冷却 = 10m × 0.1 = 1 分钟
```

**衰减率建议**（在 `[decay]` 中配置）：

| 物种类型 | Hunger | Happiness | Energy | Health |
|---------|--------|-----------|--------|--------|
| 低维护 | 0.5-0.8 | 0.3-0.5 | 0.2-0.3 | 0.1-0.2 |
| 标准型 | 1.0 | 0.5 | 0.3 | 0.2 |
| 高维护 | 1.5-2.0 | 0.8-1.0 | 0.4-0.5 | 0.3-0.5 |

**推荐配置**：使用统一慢速衰减（hunger=1.0, happiness=0.5, energy=0.3, health=0.2），适合离线游戏体验。

## 动作系统平衡性指南

### 属性效果范围

```toml
# 推荐的效果范围（单次操作）

[actions.effects]
hunger = 20-60      # 饱食度
happiness = 10-30   # 快乐度
health = 15-40      # 健康度
energy = 20-50      # 精力
```

**平衡原则**：
1. **正面效果总和**：单次操作的正面效果总和不超过 80
2. **负面效果限制**：负面效果不应超过正面效果的 30%
3. **冷却与效果成正比**：效果越强，冷却越长

### 示例：平衡的动作组合

```toml
# 喂食：主恢复，小副作用
[[actions]]
id = "feed"
cooldown = "10m"
[actions.effects]
hunger = 35      # 主要效果
happiness = 5    # 小加成

# 玩耍：强快乐，消耗精力
[[actions]]
id = "play"
cooldown = "8m"
energy_cost = 10  # 前置条件
[actions.effects]
happiness = 25   # 主要效果
energy = -15     # 消耗（大于 cost，因为还有体力消耗）

# 休息：强精力恢复，小负面
[[actions]]
id = "rest"
cooldown = "15m"
[actions.effects]
energy = 40      # 主要效果
health = 5       # 小加成
happiness = -5   # 小负面（休息无聊）

# 治疗：强健康恢复，高消耗
[[actions]]
id = "heal"
cooldown = "20m"
energy_cost = 15
[actions.effects]
health = 35      # 主要效果
```

## 生命周期设计

### 寿命与进化匹配

**常见错误**：
```toml
[lifecycle]
max_age_hours = 240.0  # 10 天

[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 720.0  # 30 天 - 无法达到！
```

**正确做法**：
```toml
[lifecycle]
max_age_hours = 240.0  # 10 天

# 进化时间线
# 0-1h: egg → baby
# 1-24h: baby → child  
# 24-72h: child → adult
# 72-200h: adult → legend  ✓ 在寿命内
[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 200.0  # 8.3 天 - 可以达到！
```

**原则**：确保玩家能在宠物寿命内体验完整的进化链。

### 生命周期风格选择

```toml
# 短期挑战（7天）
[lifecycle]
max_age_hours = 168.0
warning_threshold = 0.6

# 标准体验（10-30天）
max_age_hours = 240.0-720.0
warning_threshold = 0.75-0.85

# 长期陪伴（30-90天）
max_age_hours = 720.0-2160.0
warning_threshold = 0.9

# 永恒宠物
ending_type = "eternal"
```

## 测试与调试

### 使用 clipet-dev 工具

```bash
# 验证物种包
./clipet-dev validate internal/assets/builtins/cat-pack

# 测试进化条件
./clipet-dev evo info

# 强制进化测试
./clipet-dev evo to legend_arcane_shadow

# 时间跳跃测试
./clipet-dev timeskip --hours 200

# 修改属性测试
./clipet-dev set happiness 95
./clipet-dev set age_hours 200
```

### 测试清单

创建新物种包时，确保测试：

- [ ] **基本功能**：所有阶段都有 idle 帧
- [ ] **进化路径**：至少一条路径可以在寿命内完成
- [ ] **动作平衡**：单次操作有明显效果
- [ ] **对话覆盖**：每个阶段至少有 3-5 条对话
- [ ] **冒险可用**：至少有 3-5 个冒险
- [ ] **冷却合理**：不会让玩家陷入无法操作的死循环
- [ ] **个性特征**：特征效果合理，不会过强或过弱

## 常见陷阱

### 1. 过度平衡

```toml
# ❌ 错误：所有效果都很小，操作没有意义
[actions.effects]
hunger = 5
happiness = 2
```

**解决**：让操作有明显的效果，保持游戏乐趣。

### 2. 冷却过长

```toml
# ❌ 错误：冷却 1 小时，玩家早就关闭游戏了
cooldown = "1h"
```

**解决**：冷却时间不超过 30 分钟，除非效果非常强。

### 3. 进化条件过严

```toml
# ❌ 错误：需要 1000 次互动
[evolutions.condition]
min_interactions = 1000
```

**解决**：确保进化条件在合理时间内可以达成（< 500 次互动）。

### 4. 忽视收益递减

```toml
# ❌ 错误：高属性时喂食仍然恢复很多
# 应该在代码中实现收益递减，而不是在配置中硬编码
```

**解决**：代码中实现 `diminish()` 函数，高属性时效果自动降低。

---

## 扩展阅读

- **架构设计**：`docs/architecture.md`
- **进化系统**：`docs/design.md`
- **开发路线**：`docs/roadmap.md`
- **内置示例**：`internal/assets/builtins/cat-pack/`

