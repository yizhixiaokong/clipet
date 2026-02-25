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
    ├── egg_idle_0.txt
    ├── egg_idle_1.txt
    ├── baby_xxx_idle_0.txt
    ├── baby_xxx_eating_0.txt
    └── ...
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

**要求**：
- 至少有一个 `phase = "egg"` 的阶段
- 所有非 egg 阶段必须通过进化路径从 egg 可达

### 进化路径

```toml
[[evolutions]]
from = "egg"           # 来源阶段 ID
to = "baby_dragon"     # 目标阶段 ID

[evolutions.condition]
min_age_hours = 1.0              # 最低年龄（小时）
attr_bias = "happiness"          # 属性偏好: happiness | health | playful
min_dialogues = 10               # 最低对话次数
min_adventures = 5               # 最低冒险完成数
min_feed_regularity = 0.7        # 喂食规律性 (0.0-1.0)
night_interactions_bias = true   # 夜间互动偏好
day_interactions_bias = true     # 日间互动偏好
min_interactions = 500           # 最低总互动次数

[evolutions.condition.min_attr]  # 属性门槛
happiness = 90
health = 85
```

所有条件字段均为可选，未指定的条件视为已满足。

## dialogues.toml

```toml
[[dialogues]]
stage = ["baby_dragon"]     # 匹配阶段 ID 列表，"*" 匹配全部
mood = ["happy", "normal"]  # 匹配心情列表，"*" 匹配全部
lines = [                   # 随机候选台词
    "吼～感觉好开心！",
    "想要飞起来！",
]

[[dialogues]]
stage = ["*"]               # 通配符：所有阶段
mood = ["sad"]
lines = [
    "呜呜......肚子好饿......",
]
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

### 命名规范

```
{stageID}_{animState}_{index}.txt
```

- `stageID` — 阶段 ID，可包含下划线（如 `baby_dragon`）
- `animState` — 动画状态名
- `index` — 帧序号（从 0 开始，按字典序排列）

**解析规则**：从文件名末尾按 `_` 分割，最后一段为 index，倒数第二段为 animState，
其余部分连接为 stageID。

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

### 帧文件内容

纯 ASCII 文本，每个文件是一帧画面：

```
  /\_/\  
 ( o.o ) 
  > ^ <  
```

建议每帧保持一致的宽高（推荐 12-16 字符宽，6-10 行高）。

## 安装外部插件

将插件目录放入 `~/.local/share/clipet/plugins/`：

```
~/.local/share/clipet/plugins/
└── my-dragon-pack/
    ├── species.toml
    ├── dialogues.toml
    ├── adventures.toml
    └── frames/
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

## 参考：内置猫物种包

查看 `internal/assets/builtins/cat-pack/` 目录获取完整的参考实现。

进化树概览：

```
egg (神秘之蛋)
 └── baby_cat (小猫咪)
      ├── child_arcane (咒术小猫)  ← happiness偏好 + 对话数
      │    ├── adult_shadow_mage (暗影魅猫)  ← 夜间偏好
      │    │    └── legend_void_walker (虚空行者)
      │    └── adult_crystal_oracle (水晶预言猫)  ← 日间偏好
      │         └── legend_astral_sage (星辰贤者)
      ├── child_feral (战斗小猫)  ← health偏好 + 冒险数
      │    ├── adult_flame_lion (烈焰狮)  ← hunger偏好
      │    │    └── legend_immortal_inferno (不灭炎帝)
      │    └── adult_frost_panther (霜暴豹)  ← energy偏好
      │         └── legend_cryostorm_deity (极寒霜神)
      └── child_mech (机甲小猫)  ← playful偏好 + 喂食规律
           ├── adult_cyber_lynx (赛博猞猁)  ← 对话数
           │    └── legend_quantum_phantom (量子幽灵)
           └── adult_chrome_jaguar (合金猎豹)  ← 冒险数
                └── legend_stellar_predator (星际掠夺者)
```
