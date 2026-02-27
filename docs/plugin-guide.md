# Clipet 插件开发指南

## 概览

Clipet 的物种系统完全基于插件包。每个物种是一个独立的目录，包含 TOML 配置文件和 ASCII 动画帧文件。内置物种和外部插件使用完全相同的格式和加载路径。

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
description = "远古巨龙"          # 描述文字
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
ending_type = "death"       # 终局类型: death | ascend | eternal | loop
warning_threshold = 0.75    # 预警阈值（0.0-1.0），达到时显示温馨提示
```

**生命周期类型**：

| 类型 | 说明 |
|-----|------|
| `death` | 自然离世（默认）|
| `ascend` | 飞升/升华（温馨主题）|
| `eternal` | 永恒存在，永不触发终局 |
| `loop` | 循环重生，达到寿命后重置年龄而非死亡 |

### 个性特征定义 (v2.0+)

定义物种的个性特征，分为三类：被动特征、主动技能和进化修正器。

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

| 被动效应字段 | 类型 | 说明 |
|-------------|------|------|
| `feed_hunger_bonus` | float | 喂食饱食度增益倍率（-1.0 到 1.0）|
| `feed_happiness_bonus` | float | 喂食快乐度增益倍率 |
| `play_happiness_bonus` | float | 玩耍快乐度增益倍率 |
| `sleep_energy_bonus` | float | 休息精力增益倍率 |
| `resurrect_chance` | float | 死亡时复活概率（0.0-1.0）|
| `health_restore_percent` | float | 复活时恢复生命值百分比 |

| 主动技能字段 | 类型 | 说明 |
|-------------|------|------|
| `energy_cost` | int | 主动技能精力消耗 |
| `health_restore` | int | 主动技能健康恢复量 |
| `hunger_restore` | int | 主动技能饱食度恢复量 |
| `happiness_boost` | int | 主动技能快乐度提升量 |
| `cooldown` | duration | 主动技能冷却时间（如 "30m", "1h"）|

| 进化修正器字段 | 类型 | 说明 |
|-------------|------|------|
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

### 动作配置 (v3.0+, Phase 7)

定义物种的动作行为，包括冷却时间和效果数值：

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

### decay 配置 (v3.0+, Phase 7)

定义物种的属性衰减率：

```toml
[decay]
hunger = 1.0        # 饱食度每小时衰减 1 点
happiness = 0.5     # 快乐度每小时衰减 0.5 点
energy = 0.3        # 精力每小时衰减 0.3 点
health = 0.2        # 饥饿时健康每小时衰减 0.2 点
```

### dynamic_cooldown 配置 (v3.0+, Phase 7)

定义动态冷却系统，根据属性紧急度自动调整冷却时间：

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

### 进化条件

进化条件支持多种检查类型：

```toml
[[evolutions]]
from = "child"
to = "adult_fire"
[evolutions.condition]
min_age_hours = 72.0            # 最低年龄
min_interactions = 50           # 最低互动次数
attr_bias = "happiness"         # 属性偏好
min_attr = { happiness = 70 }   # 最低属性要求

# 自定义属性累积器（v3.0+）
custom_acc = { fire_power = 30 }  # 需要火之力量达到 30

# 时间偏好
night_bias = true                # 夜间偏好
day_bias = false                 # 日间偏好
```

**进化条件字段**：

| 字段 | 类型 | 说明 |
|-----|------|------|
| `min_age_hours` | float | 最低年龄（小时）|
| `min_attr` | map | 最低属性要求（如 `{happiness = 70}`）|
| `min_interactions` | int | 最低互动次数 |
| `min_feed_count` | int | 最低喂食次数 |
| `min_dialogues` | int | 最低对话次数 |
| `min_adventures` | int | 最低冒险次数 |
| `min_feed_regularity` | float | 最低喂食规律性（0-1）|
| `attr_bias` | string | 属性偏好（"happiness", "health", "playful"）|
| `night_bias` | bool | 夜间偏好 |
| `day_bias` | bool | 日间偏好 |
| `custom_acc` | map | 自定义累积器要求（v3.0+）|

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

### 阶段匹配规则

- `"*"` — 匹配所有阶段
- `"child_*"` — 前缀通配，匹配所有 `child_` 开头的阶段
- `"baby_dragon"` — 精确匹配

### 心情值

| 心情名称 | 心情分数范围 |
|---------|-------------|
| happy | 81-100 |
| normal | 61-80 |
| unhappy | 41-60 |
| sad | 21-40 |
| miserable | 0-20 |

## adventures.toml

### 基本格式

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

# 自定义属性累积器（v3.0+）
fire_power = 5                  # 火之力量 +5

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

### 冒险效果字段

| 字段 | 类型 | 说明 |
|-----|------|------|
| `hunger` | int | 饱食度变化 |
| `happiness` | int | 快乐度变化 |
| `health` | int | 健康度变化 |
| `energy` | int | 精力变化 |
| `{custom_attr}` | int | 自定义属性变化（v3.0+）|

## 动画帧文件

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

### 支持的动画状态

| animState | 触发条件 |
|-----------|---------|
| idle | 默认状态 (必须提供) |
| eating | 喂食时 |
| sleeping | 精力低于 15 |
| playing | 玩耍时 |
| happy | 心情 > 80 |
| sad | 心情 < 40 |

如果某状态没有帧文件，会自动 fallback 到 `idle`。

## 自定义属性系统 (v3.0+)

插件可以通过冒险事件定义自定义属性累积器，用于创建独特的进化路径。

### 定义自定义属性

在冒险事件的效果中使用自定义属性名称：

```toml
[[adventures]]
id = "fire_shrine"
name = "火焰神殿"
description = "一座燃烧着永恒之火的古老神殿..."

[[adventures.choices]]
text = "献祭力量"
outcomes = [
  { weight = 50, text = "火焰之力涌入体内！", effects = { fire_power = 10, happiness = 15 } },
  { weight = 30, text = "感受到了温暖的火焰。", effects = { fire_power = 5, energy = 10 } },
]
```

### 在进化条件中使用

```toml
[[evolutions]]
from = "child"
to = "adult_fire"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { fire_power = 30 }  # 需要火之力量达到 30
```

### 工作原理

1. **累积**：通过冒险事件累积自定义属性值
2. **检查**：进化时检查是否满足自定义属性要求
3. **持久化**：自定义属性值保存在存档中

**注意事项**：
- 自定义属性值从 0 开始累积
- 同一个自定义属性可以通过多个冒险事件增加
- 自定义属性不影响核心四属性（饥饿、快乐、健康、精力）

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

## 开发工具

使用 `clipet-dev` 工具进行开发和测试：

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

## 扩展阅读

- **设计最佳实践**: [plugin-best-practices.md](plugin-best-practices.md)
- **架构设计**: [CODEMAPS/architecture.md](CODEMAPS/architecture.md)
- **核心逻辑**: [CODEMAPS/core-logic.md](CODEMAPS/core-logic.md)
- **数据结构**: [CODEMAPS/data-structures.md](CODEMAPS/data-structures.md)
