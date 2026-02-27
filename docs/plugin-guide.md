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
