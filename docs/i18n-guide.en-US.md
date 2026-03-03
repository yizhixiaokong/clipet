# Internationalization (i18n) Guide

**[中文](i18n-guide.md) | English**

Clipet supports multi-language internationalization, allowing users to use the application in different language environments.

## Quick Start

### Switching Languages

**Temporary Switch** (Recommended):
```bash
# Use English interface
CLIPET_LANG=en-US clipet

# Use Chinese interface
CLIPET_LANG=zh-CN clipet
```

**Permanent Switch**:
Edit the configuration file `~/.config/clipet/config.json`:
```json
{
  "language": "en-US",
  "fallback_language": "zh-CN",
  "version": "1.0"
}
```

### Language Detection Priority

1. **`CLIPET_LANG`** environment variable (highest priority)
2. **`LANG`** environment variable
3. **`LC_ALL`** environment variable
4. **Configuration file** `~/.config/clipet/config.json`
5. **Default value** `zh-CN` (lowest priority)

## Architecture Design

### Core Components

```
internal/i18n/
├── i18n.go          # Manager (language detection, T() function)
├── bundle.go        # Translation bundle management (fallback chain)
├── loader.go        # Load translation files (embed + filesystem)
└── plural.go        # Plural rules
```

### API Usage

**Simple Translation**:
```go
i18n.T("ui.home.feed_success", "oldHunger", 50, "newHunger", 75)
// Output: "Feeding successful! Hunger 50 → 75"
```

**Plural Translation**:
```go
i18n.TN("game.stats.interactions", count, "count", count)
// Automatically selects "interactions_one" or "interactions_other"
```

**Runtime Language Switching**:
```go
i18n.SetLanguage("en-US")
```

## Translation File Format

### Directory Structure

```
internal/assets/locales/
├── zh-CN/
│   ├── tui.json       # TUI interface text
│   ├── game.json      # Game logic messages
│   └── cli.json       # CLI command output
└── en-US/
    ├── tui.json
    ├── game.json
    └── cli.json
```

### JSON Format

```json
{
  "ui": {
    "home": {
      "feed_success": "Feeding successful! Hunger {{.oldHunger}} → {{.newHunger}}",
      "play_success": "Playtime! Happiness {{.oldHappiness}} → {{.newHappiness}}"
    },
    "common": {
      "quit": "Goodbye!"
    }
  }
}
```

### Naming Conventions

Use hierarchical naming: `<domain>.<component>.<item>[.<variant>]`

Examples:
- `ui.home.feed_success` - UI interface, home component, feed success message
- `game.stats.hunger` - Game logic, stats, hunger
- `cli.init.welcome` - CLI command, init command, welcome message

### Template Variables

Use Go's `text/template` syntax:
- `{{.variableName}}` - Variable interpolation
- Variables passed as key-value pairs: `i18n.T("key", "name", value, ...)`

## Plugin Multi-language Support

### Plugin Locale Files

Plugins can provide translation files in their own directories:

```
internal/assets/builtins/cat-pack/
├── locales/
│   ├── zh-CN.json     # Chinese translation
│   └── en-US.json     # English translation
├── species.toml
├── dialogues.toml
└── adventures.toml
```

### locale.json Structure

```json
{
  "species": {
    "cat": {
      "name": "Cat",
      "description": "Agile little kitty..."
    }
  },
  "stages": {
    "egg": "Mysterious Egg",
    "baby": "Kitten",
    "adult_arcane_crystal": "Arcane Crystal Cat"
  },
  "dialogues": {
    "baby": {
      "happy": ["Meow~", "Meow meow~"],
      "sad": ["Meow...", "Aww..."]
    }
  },
  "adventures": {
    "explore_garden": {
      "name": "Explore Garden",
      "description": "The kitten wants to explore the garden...",
      "choices": {
        "follow": "Follow quietly"
      }
    }
  }
}
```

### Fallback Chain

When the requested language is unavailable, the system falls back in order:

1. **Requested language** (e.g., `en-US`)
2. **Fallback language** (e.g., `zh-CN`)
3. **Inline TOML text** (from species.toml, dialogues.toml)

This ensures plugins work even without translation files.

## Adding New Translations

### Step 1: Extract Strings

Find hardcoded strings in code:
```go
fmt.Printf("Feeding successful! Hunger %d → %d", old, new)
```

### Step 2: Add to Translation Files

`internal/assets/locales/zh-CN/tui.json`:
```json
{
  "ui": {
    "home": {
      "feed_success": "喂食成功！饱腹度 {{.oldHunger}} → {{.newHunger}}"
    }
  }
}
```

`internal/assets/locales/en-US/tui.json`:
```json
{
  "ui": {
    "home": {
      "feed_success": "Feeding successful! Hunger {{.oldHunger}} → {{.newHunger}}"
    }
  }
}
```

### Step 3: Update Code

```go
// Before
fmt.Printf("Feeding successful! Hunger %d → %d", old, new)

// After
i18n.T("ui.home.feed_success", "oldHunger", old, "newHunger", new)
```

### Step 4: Recompile

```bash
go build ./cmd/clipet
```

## Plural Handling

### Translation Files

Provide singular and plural forms:
```json
{
  "game.stats.interactions_one": "{{.count}} interaction",
  "game.stats.interactions_other": "{{.count}} interactions"
}
```

### Code Usage

```go
i18n.TN("game.stats.interactions", count, "count", count)
```

### Supported Languages

- **Chinese/Japanese**: No plural forms
- **English/German**: singular (1) vs plural (n != 1)
- **French**: singular (0, 1) vs plural (n > 1)
- **Polish/Russian**: Complex plural rules

## Testing Translations

### Unit Tests

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

### Manual Testing

```bash
# Test Chinese
CLIPET_LANG=zh-CN clipet init

# Test English
CLIPET_LANG=en-US clipet init
```

## Performance Considerations

- **Compile-time embed**: Translation files embedded in binary via `go:embed`
- **Memory cache**: Translations cached in memory after first load
- **Thread-safe**: Uses `sync.RWMutex` for concurrent access safety

## Troubleshooting

### Translations Not Showing

1. **Check language settings**:
   ```bash
   echo $CLIPET_LANG
   cat ~/.config/clipet/config.json
   ```

2. **Check translation keys**: Ensure corresponding keys exist in JSON files

3. **Check logs**: Missing translations show warnings in logs

### Plugin Locale Not Loading

1. **Check file path**: `locales/zh-CN.json` (must be lowercase)
2. **Check JSON format**: Validate with `jq`
   ```bash
   jq . locales/zh-CN.json
   ```

## Future Extensions

Planned features:
- More language support (Japanese, Korean, Spanish, etc.)
- Dynamic language switching (no restart needed)
- Translation management tools (automatic string extraction)
- Community translation platform integration

## Contributing Translations

Contributions of new translations are welcome!

1. Fork the project
2. Copy `locales/en-US/` to `locales/{your-lang}/`
3. Translate strings in JSON files
4. Submit a Pull Request

---

**Currently Supported Languages**:
- 🇨🇳 Chinese (Simplified) - `zh-CN`
- 🇺🇸 English - `en-US`
