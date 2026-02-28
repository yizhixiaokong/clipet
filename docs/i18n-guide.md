# Internationalization (i18n) Guide

Clipet 支持多语言国际化，允许用户在不同语言环境下使用应用。

## 快速开始

### 切换语言

**临时切换**（推荐）：
```bash
# 使用英文界面
CLIPET_LANG=en-US clipet

# 使用中文界面
CLIPET_LANG=zh-CN clipet
```

**永久切换**：
编辑配置文件 `~/.config/clipet/config.json`：
```json
{
  "language": "en-US",
  "fallback_language": "zh-CN",
  "version": "1.0"
}
```

### 语言检测优先级

1. **`CLIPET_LANG`** 环境变量（最高优先级）
2. **`LANG`** 环境变量
3. **`LC_ALL`** 环境变量
4. **配置文件** `~/.config/clipet/config.json`
5. **默认值** `zh-CN`（最低优先级）

## 架构设计

### 核心组件

```
internal/i18n/
├── i18n.go          # Manager（语言检测、T() 函数）
├── bundle.go        # 翻译包管理（回退链）
├── loader.go        # 加载翻译文件（embed + 文件系统）
└── plural.go        # 复数规则
```

### API 使用

**简单翻译**：
```go
i18n.T("ui.home.feed_success", "oldHunger", 50, "newHunger", 75)
// Output: "喂食成功！饱腹度 50 → 75"
```

**复数翻译**：
```go
i18n.TN("game.stats.interactions", count, "count", count)
// Automatically selects "interactions_one" or "interactions_other"
```

**运行时切换语言**：
```go
i18n.SetLanguage("en-US")
```

## 翻译文件格式

### 目录结构

```
internal/assets/locales/
├── zh-CN/
│   ├── tui.json       # TUI 界面文本
│   ├── game.json      # 游戏逻辑消息
│   └── cli.json       # CLI 命令输出
└── en-US/
    ├── tui.json
    ├── game.json
    └── cli.json
```

### JSON 格式

```json
{
  "ui": {
    "home": {
      "feed_success": "喂食成功！饱腹度 {{.oldHunger}} → {{.newHunger}}",
      "play_success": "玩耍愉快！快乐度 {{.oldHappiness}} → {{.newHappiness}}"
    },
    "common": {
      "quit": "再见！"
    }
  }
}
```

### 命名约定

使用层级命名：`<domain>.<component>.<item>[.<variant>]`

示例：
- `ui.home.feed_success` - UI 界面，home 组件，喂食成功消息
- `game.stats.hunger` - 游戏逻辑，统计，饱腹度
- `cli.init.welcome` - CLI 命令，init 命令，欢迎消息

### 模板变量

使用 Go 的 `text/template` 语法：
- `{{.variableName}}` - 变量插值
- 变量通过 key-value 对传递：`i18n.T("key", "name", value, ...)`

## 插件多语言支持

### 插件 locale 文件

插件可以在自己的目录中提供翻译文件：

```
internal/assets/builtins/cat-pack/
├── locales/
│   ├── zh-CN.json     # 中文翻译
│   └── en-US.json     # 英文翻译
├── species.toml
├── dialogues.toml
└── adventures.toml
```

### locale.json 结构

```json
{
  "species": {
    "cat": {
      "name": "猫",
      "description": "灵动的小猫咪..."
    }
  },
  "stages": {
    "egg": "神秘之蛋",
    "baby": "小猫咪",
    "adult_arcane_crystal": "奥术晶能猫"
  },
  "dialogues": {
    "baby": {
      "happy": ["喵~", "喵喵~"],
      "sad": ["喵...", "呜..."]
    }
  },
  "adventures": {
    "explore_garden": {
      "name": "探索花园",
      "description": "小猫咪想去花园里探险...",
      "choices": {
        "follow": "悄悄跟随"
      }
    }
  }
}
```

### 回退链

当请求的语言不可用时，系统会按以下顺序回退：

1. **请求语言** (e.g., `en-US`)
2. **回退语言** (e.g., `zh-CN`)
3. **内联 TOML 文本** (from species.toml, dialogues.toml)

这确保了即使没有翻译文件，插件也能正常工作。

## 添加新翻译

### 步骤 1：提取字符串

找到代码中的硬编码字符串：
```go
fmt.Printf("喂食成功！饱腹度 %d → %d", old, new)
```

### 步骤 2：添加到翻译文件

`internal/assets/locales/zh-CN/tui.json`：
```json
{
  "ui": {
    "home": {
      "feed_success": "喂食成功！饱腹度 {{.oldHunger}} → {{.newHunger}}"
    }
  }
}
```

`internal/assets/locales/en-US/tui.json`：
```json
{
  "ui": {
    "home": {
      "feed_success": "Feeding successful! Hunger {{.oldHunger}} → {{.newHunger}}"
    }
  }
}
```

### 步骤 3：更新代码

```go
// 之前
fmt.Printf("喂食成功！饱腹度 %d → %d", old, new)

// 之后
i18n.T("ui.home.feed_success", "oldHunger", old, "newHunger", new)
```

### 步骤 4：重新编译

```bash
go build ./cmd/clipet
```

## 复数处理

### 翻译文件

提供单数和复数形式：
```json
{
  "game.stats.interactions_one": "{{.count}} 次互动",
  "game.stats.interactions_other": "{{.count}} 次互动"
}
```

### 代码使用

```go
i18n.TN("game.stats.interactions", count, "count", count)
```

### 支持的语言

- **中文/日文**：无复数形式
- **英文/德文**：singular (1) vs plural (n != 1)
- **法文**：singular (0, 1) vs plural (n > 1)
- **波兰/俄文**：复杂复数规则

## 测试翻译

### 单元测试

```go
func TestTranslation(t *testing.T) {
    bundle := i18n.NewBundle()
    // Load test translations
    mgr := i18n.NewManager("en-US", "zh-CN", bundle)

    result := mgr.T("ui.home.feed_success", "oldHunger", 50, "newHunger", 75)
    expected := "Feeding successful! Hunger 50 → 75"

    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

### 手动测试

```bash
# 测试中文
CLIPET_LANG=zh-CN clipet init

# 测试英文
CLIPET_LANG=en-US clipet init
```

## 性能考虑

- **编译时 embed**：翻译文件通过 `go:embed` 嵌入二进制文件
- **内存缓存**：翻译在首次加载后缓存在内存中
- **线程安全**：使用 `sync.RWMutex` 保证并发访问安全

## 故障排查

### 翻译不显示

1. **检查语言设置**：
   ```bash
   echo $CLIPET_LANG
   cat ~/.config/clipet/config.json
   ```

2. **检查翻译键**：确保 JSON 文件中存在对应的键

3. **检查日志**：缺失的翻译会在日志中显示警告

### 插件 locale 不加载

1. **检查文件路径**：`locales/zh-CN.json`（必须是小写）
2. **检查 JSON 格式**：使用 `jq` 验证
   ```bash
   jq . locales/zh-CN.json
   ```

## 未来扩展

计划中的功能：
- 更多语言支持（日语、韩语、西班牙语等）
- 动态语言切换（无需重启）
- 翻译管理工具（自动提取字符串）
- 社区翻译平台集成

## 贡献翻译

欢迎贡献新的翻译！

1. Fork 项目
2. 复制 `locales/en-US/` 到 `locales/{your-lang}/`
3. 翻译 JSON 文件中的字符串
4. 提交 Pull Request

---

**当前支持的语言**：
- 🇨🇳 中文（简体）- `zh-CN`
- 🇺🇸 English - `en-US`
