# Clipet 插件设计最佳实践

## 概览

本文档提供插件开发的设计原则、最佳实践和常见陷阱，帮助创建平衡、有趣且易于维护的物种插件。

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

## 进化树设计

### 基本原则

1. **可达性**：确保所有进化路径在宠物寿命内可以完成
2. **多样性**：提供多种进化选择，鼓励重玩
3. **平衡性**：不同路径的难度应该相对平衡
4. **主题性**：每个进化路径应该有清晰的主题

### 进化条件设计

**年龄限制**：
```toml
# ❌ 错误：寿命 10 天，但进化需要 30 天
[lifecycle]
max_age_hours = 240.0  # 10 天

[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 720.0  # 30 天 - 无法达到！

# ✅ 正确：进化时间在寿命内
[lifecycle]
max_age_hours = 240.0  # 10 天

[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 200.0  # 8.3 天 - 可以达到
```

**互动次数**：
```toml
# ❌ 错误：需要 1000 次互动，不现实
[evolutions.condition]
min_interactions = 1000

# ✅ 正确：合理的互动次数
[evolutions.condition]
min_interactions = 50  # 约 5-10 天可以达到
```

**属性偏好**：
```toml
# 使用属性偏好创造不同的进化路径
[[evolutions]]
from = "child"
to = "adult_arcane"
[evolutions.condition]
attr_bias = "happiness"  # 偏向高快乐度

[[evolutions]]
from = "child"
to = "adult_feral"
[evolutions.condition]
attr_bias = "health"     # 偏向高健康度
```

### 自定义属性系统设计

自定义属性是 v3.0+ 的强大功能，允许插件创建独特的进化路径。

**设计原则**：

1. **明确主题**：每个自定义属性代表一个明确的方向或主题
2. **渐进式**：通过多个冒险事件逐步累积
3. **可平衡**：确保不同路径的难度相对平衡

**示例：三路线进化**

```toml
# 奥术路线 - 通过奥术亲和累积
[[evolutions]]
from = "child"
to = "adult_arcane_shadow"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { arcane_affinity = 50 }

# 狂野路线 - 通过狂野亲和累积
[[evolutions]]
from = "child"
to = "adult_feral_flame"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { feral_affinity = 50 }

# 机械路线 - 通过机械亲和累积
[[evolutions]]
from = "child"
to = "adult_mech_cyber"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { mech_affinity = 50 }
```

**冒险事件设计**：

```toml
# 奥术启蒙事件
[[adventures]]
id = "arcane_spark"
name = "魔法火花"
description = "一道神秘的紫色光芒在你眼前闪烁..."

[[adventures.choices]]
text = "触碰魔法火花"
outcomes = [
  { weight = 50, text = "魔法能量涌入体内！", effects = { arcane_affinity = 10, happiness = 15 } },
  { weight = 30, text = "火花轻轻环绕着你。", effects = { arcane_affinity = 6, energy = 10 } },
  { weight = 20, text = "留下了一丝印记。", effects = { arcane_affinity = 3 } },
]
```

**平衡性考虑**：

- **累积速度**：确保玩家在合理时间内可以达到目标值
- **权重分配**：高风险高回报的权重应该较低
- **多路径支持**：允许玩家通过不同事件累积同一个属性

## 对话设计

### 阶段适配对话

对话复杂度应与生命阶段匹配：

| 生命阶段 | 建议对话内容 | 示例 |
|---------|-------------|------|
| **egg** | 仅简单声响 | `"咔嗒..."`, `"咚..."`, `"蛋壳轻响"` |
| **baby** | 简单音节，重复性高 | `"喵~"`, `"喵喵~"`, `"呜..."` |
| **child** | 简单短句，语气幼稚 | `"我们一起玩吧！"`, `"我好开心~"` |
| **adult** | 完整句子，表达自己的想法 | `"我需要一些休息时间。"`, `"感谢你的照顾。"` |
| **legend** | 深刻，富有哲理，有格局 | `"守护众生是我的使命。"` |

### 心情状态反馈

**positive (happy)**: 积极、活跃、期待互动
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

**neutral (normal)**: 平静、日常、温和表达
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

**negative (unhappy/sad/miserable)**: 消极、疲惫、寻求安慰
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

## 多语言支持 (i18n)

### 何时创建 locale 文件

**推荐**：如果你想支持多语言，或计划未来支持多语言

**不必要**：仅用于单语言社区（如仅中文用户）

### locale 文件最佳实践

#### 1. **完整覆盖所有文本**

```json
{
  "species": {
    "dragon": {
      "name": "龙",
      "description": "远古巨龙"
    }
  },
  "stages": {
    "egg": "神秘之蛋",
    "baby": "幼龙",
    "child_fire": "火焰幼龙",
    "adult_fire": "火焰巨龙"
  },
  "dialogues": {
    "baby": {
      "happy": ["嗷呜~", "吼~"],
      "sad": ["呜...", "嗷..."]
    }
  },
  "adventures": {
    "fire_shrine": {
      "name": "火焰神殿",
      "description": "一座燃烧着永恒之火的古老神殿...",
      "choices": {
        "enter": "进入神殿",
        "wait": "在外面等待"
      }
    }
  }
}
```

#### 2. **保留内联 TOML 文本作为回退**

即使创建 locale 文件，也要在 TOML 中保留默认语言的文本：

```toml
# species.toml
[species]
name = "龙"  # 保留中文作为回退
description = "远古巨龙"

[[stages]]
id = "egg"
name = "神秘之蛋"  # 保留中文作为回退
```

**原因**：
- 向后兼容旧版本
- 回退链：`en-US` → `zh-CN` → TOML 内联文本
- 避免缺失翻译时显示空白

#### 3. **渐进式翻译策略**

不必一次性翻译所有内容，可以渐进式添加：

**Phase 1**：先翻译核心内容
- 物种名称和描述
- 阶段名称

**Phase 2**：再翻译常见内容
- 常用对话（happy、normal 心情）
- 主要冒险事件

**Phase 3**：最后翻译详细内容
- 所有心情的对话
- 所有冒险文本

#### 4. **locale 文件命名规范**

```
locales/
├── zh-CN.json     # 中文（简体）
├── en-US.json     # 英文（美国）
├── ja-JP.json     # 日文
└── ko-KR.json     # 韩文
```

使用标准的语言代码：`{language}-{region}`（如 `zh-CN`、`en-US`）

#### 5. **测试 locale 文件**

使用 `jq` 验证 JSON 格式：
```bash
jq . locales/zh-CN.json
jq . locales/en-US.json
```

测试不同语言：
```bash
CLIPET_LANG=en-US ./clipet
CLIPET_LANG=zh-CN ./clipet
```

### 多语言插件发布建议

1. **默认包含主语言**：在 locale 文件中至少包含一种完整语言
2. **标注支持语言**：在插件 README 或 species.toml 中说明支持的语言
3. **社区贡献**：欢迎社区贡献新语言翻译
4. **版本控制**：locale 文件变更时更新插件版本号

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

## 安全约束系统

为防止插件滥用（过短寿命、极端数值、过多危机），系统实施轻量级安全边界。

### 默认约束

| 约束类型 | 默认值 | 说明 |
|---------|--------|------|
| 最小寿命 | 24 小时 | 防止瞬间死亡（eternal 除外）|
| 最大寿命 | 10 年 | 防止永不死亡（eternal 除外）|
| 属性修正器 | 10% - 300% | 防止极端增益/惩罚 |
| 冒险数量上限 | 20 | 防止玩家 overwhelmed |
| 对话组数量上限 | 100 | 防止内容过载 |

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
- [ ] **自定义属性**：自定义属性可以通过冒险事件正常累积

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

### 5. 对话不匹配阶段

```toml
# ❌ 错误：幼年阶段说复杂的话
[[dialogues]]
stage = ["baby"]
mood = ["happy"]
lines = [
  "我认为这种哲学观点值得深入探讨。"  # 太复杂了！
]
```

**解决**：确保对话复杂度与生命阶段匹配。

## 扩展阅读

- **插件开发指南**: [plugin-guide.md](plugin-guide.md)
- **架构设计**: [CODEMAPS/architecture.md](CODEMAPS/architecture.md)
- **核心逻辑**: [CODEMAPS/core-logic.md](CODEMAPS/core-logic.md)
- **数据结构**: [CODEMAPS/data-structures.md](CODEMAPS/data-structures.md)
