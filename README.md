# 🐾 Clipet — 终端宠物伴侣

Clipet 是一个运行在终端中的虚拟宠物养成游戏，使用 Go 和 [Bubble Tea](https://charm.sh/) 构建。

## 功能

- **TUI 交互界面** — ASCII 艺术宠物 + 状态面板 + 二级分类菜单
- **宠物养成** — 喂食、玩耍、对话、休息、治疗，属性随时间衰减
- **进化系统** — 基于属性和互动的条件触发进化，多路径进化树
- **迷你游戏** — 反应速度测试、猜数字，胜负影响宠物属性
- **对话系统** — 阶段/心情对应的对话内容，自动闲聊气泡
- **插件化物种** — 通过 TOML 配置自定义物种、进化条件、对话
- **国际化支持** — 支持中文和英文界面，可配置语言切换
- **离线衰减** — 关闭后属性自动衰减，重新打开时补偿计算
- **冷却 & 收益递减** — 操作有冷却时间，属性越高收益越低

## 快速开始

```bash
# 构建
make build

# 初始化宠物
./clipet init

# 启动 TUI（默认中文）
./clipet

# 使用英文界面
CLIPET_LANG=en-US ./clipet

# CLI 命令
./clipet status
```

## 操作指南

### TUI 快捷键

| 键 | 功能 |
|----|------|
| `←` `→` | 切换分类 / 选择操作 |
| `↓` / `Enter` | 进入分类 / 确认操作 |
| `↑` / `Esc` | 返回上级 |
| `f` `p` `r` `c` `t` | 快捷：喂食 / 玩耍 / 休息 / 治疗 / 对话 |
| `q` | 退出 |

### 游戏中

| 键 | 功能 |
|----|------|
| 任意键 | 反应速度测试：出现 GO! 时按下 |
| `0`-`9` + `Enter` | 猜数字：输入数字确认 |
| `Esc` | 退出游戏 |
| `Enter` | 确认结果返回主界面 |

## 项目结构

```
clipet/
├── cmd/
│   ├── clipet/          # 主程序入口
│   └── clipet-dev/      # 开发者工具 (timeskip, set, evolve, validate, preview)
├── internal/
│   ├── assets/          # 内置物种包 (go:embed)
│   ├── cli/             # Cobra CLI 命令
│   ├── game/            # 核心游戏逻辑 (pet, evolution)
│   │   └── games/       # 迷你游戏 (types, manager, reaction, guess)
│   ├── plugin/          # 插件系统 (types, parser, validator, loader, registry)
│   ├── store/           # 持久化 (JSON)
│   └── tui/             # Bubble Tea TUI
│       ├── components/  # UI 组件 (petview, dialoguebubble)
│       ├── screens/     # 屏幕 (home, evolve)
│       └── styles/      # 主题和颜色
└── docs/                # 文档
```

## 开发

```bash
# 开发者工具
make dev

# 时间跳跃测试衰减
./clipet-dev timeskip --hours 2

# 手动设属性（交互式）
./clipet-dev set

# 强制进化（交互式）
./clipet-dev evo to

# 查看进化信息
./clipet-dev evo info
```

## 自定义物种

Clipet 支持完全可定制的插件系统。每个物种是一个独立的插件包，包含：
- 物种定义和进化树
- 生命周期和终局配置
- 对话库和冒险事件
- ASCII 动画帧文件
- 个性特征（被动/主动/修正器）

### 自定义属性系统 (v3.0+)

插件可以通过冒险事件定义自定义属性累积器，用于创建独特的进化路径。

**工作原理**：

1. **定义自定义属性**：在冒险事件的效果中使用任意名称作为自定义属性
2. **累积属性值**：玩家通过完成冒险事件累积自定义属性
3. **检查进化条件**：进化系统可以检查自定义属性是否达到阈值

**示例**：内置猫物种包定义了 10 个自定义属性，支持三条进化路线：

| 自定义属性 | 进化路线 | 获得方式 |
|-----------|---------|---------|
| `arcane_affinity` | 奥术路线 | 完成奥术主题冒险 |
| `feral_affinity` | 狂野路线 | 完成战斗主题冒险 |
| `mech_affinity` | 机械路线 | 完成科技主题冒险 |

**TOML 配置示例**：

```toml
# 冒险事件中累积自定义属性
[[adventures]]
id = "mystic_shrine"
name = "神秘神殿"
[[adventures.choices]]
text = "探索神殿"
outcomes = [
  { weight = 50, text = "感受到奥术能量！", effects = { arcane_affinity = 10 } },
]

# 进化条件中检查自定义属性
[[evolutions]]
from = "child"
to = "adult_arcane_shadow"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { arcane_affinity = 50 }  # 需要奥术亲和达到 50
```

这种设计允许插件创建独特的进化路径，而无需修改核心框架代码。

参考 [插件开发指南](docs/plugin-guide.md) 和 [插件设计最佳实践](docs/plugin-best-practices.md) 创建自定义物种包。

## 技术栈

- **Go** 1.25+ — 语言
- **Bubble Tea v2** — TUI 框架 (事件驱动)
- **Lipgloss v2** — 终端样式
- **Cobra** — CLI 框架
- **TOML** — 物种/对话配置
- **JSON** — 存档持久化 + 配置文件
- **i18n** — 轻量级国际化框架 (自研，零外部依赖)

## 国际化 (i18n)

Clipet 支持多语言界面，当前支持中文和英文。

### 切换语言

**临时切换**（推荐）：
```bash
# 使用英文
CLIPET_LANG=en-US ./clipet

# 使用中文
CLIPET_LANG=zh-CN ./clipet
```

**永久切换**：
编辑配置文件 `~/.config/clipet/config.json`：
```json
{
  "language": "en-US",
  "fallback_language": "zh-CN"
}
```

### 语言检测优先级

1. `CLIPET_LANG` 环境变量
2. `LANG` 环境变量
3. `LC_ALL` 环境变量
4. 配置文件 `~/.config/clipet/config.json`
5. 默认值 `zh-CN`

详细的 i18n 使用和开发指南，请参考 [docs/i18n-guide.md](docs/i18n-guide.md)。

## 许可证

MIT
