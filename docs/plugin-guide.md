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

支持两种格式，推荐使用精灵图格式。

### 格式一：精灵图（推荐）

```
{stageID}_{animState}.txt
```

同一动作的多帧放在一个文件中，用 `---` 行分隔：

```
  /\_/\
 ( o.o )
  > ^ <
 /|   |\
(_|   |_)
---
  /\_/\
 ( -.- )
  > ^ <
 /|   |\
(_|   |_)
```

**优势**：文件数少，同动作帧集中管理，编辑对比方便。

### 格式二：逐帧文件（兼容）

```
{stageID}_{animState}_{index}.txt
```

- `index` — 帧序号（从 0 开始，按字典序排列）

每个文件包含一帧画面。

**解析规则**：从文件名末尾按 `_` 分割，最后一段为 index，倒数第二段为 animState，
其余部分连接为 stageID。

> **优先级**：当同一 (stageID, animState) 同时存在精灵图和逐帧文件时，精灵图优先。

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

## 已知问题

- 蛋阶段设计为简单的声音反馈，符合生命初期不能说话的设定
- 幼年阶段使用简单重复的音节，模拟幼崽学习语言的过程
- 高级阶段的对话会变得更完整、更有深度
