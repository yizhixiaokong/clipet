<!-- Generated: 2026-02-28 | Files scanned: 4 | Token estimate: ~600 -->

# i18n System Architecture (v3.1)

## Package: internal/i18n/

Zero-dependency lightweight internationalization framework

### Core Components

```
internal/i18n/
├── i18n.go       (Manager - public API)
├── bundle.go     (Bundle - translation storage)
├── loader.go     (Loader - file loading)
└── plural.go     (Plural rules by language)
```

## Manager (i18n.go)

**Thread-safe public API for translations**

```go
type Manager struct {
    mu       sync.RWMutex
    language string
    bundle   *Bundle
    fallback string
}

// Primary translation function
func (m *Manager) T(key string, args ...interface{}) string
// key: "ui.home.feed_success"
// args: "oldHunger", 50, "newHunger", 75 (key-value pairs)

// Plural-aware translation
func (m *Manager) TN(key string, count int, args ...interface{}) string
// Automatically selects singular/plural form
```

**Key Features**:
- Thread-safe with RWMutex
- Automatic fallback to secondary language
- Template interpolation via Go's `text/template`
- Zero external dependencies

## Bundle (bundle.go)

**Translation storage with nested key lookup**

```go
type Bundle struct {
    mu     sync.RWMutex
    stores map[string]*Store  // language → translations
}

type Store struct {
    Language string
    Data     map[string]interface{}  // nested JSON structure
}

// Lookup navigates nested maps: "ui.home.feed_success"
func (b *Bundle) Lookup(lang, key string, vars map[string]interface{}) (string, error)
```

**Lookup Algorithm**:
1. Split key by `.` → `["ui", "home", "feed_success"]`
2. Navigate nested map → `data["ui"]["home"]["feed_success"]`
3. Execute template with variables
4. Return rendered string

## Loader (loader.go)

**Filesystem-based locale loader**

```go
type Loader struct {
    fsys       fs.FS      // embed.FS or os.DirFS
    root       string     // "locales" directory
    language   string
    fallback   string
}

// LoadAll loads all languages from filesystem
func (l *Loader) LoadAll(bundle *Bundle) error

// Structure expected:
// locales/
// ├── zh-CN/
// │   ├── tui.json
// │   └── game.json
// └── en-US/
//     ├── tui.json
//     └── game.json
```

**File Loading**:
- Reads JSON files from embedded or real filesystem
- Merges multiple files per language
- Validates JSON structure

## Plural Rules (plural.go)

**Language-specific pluralization**

```go
type PluralRule func(n int) int

// Returns plural form index:
// 0 = singular, 1 = plural, etc.

var PluralRules = map[string]PluralRule{
    "zh-CN": ChinesePlural,  // No plurals
    "en-US": EnglishPlural,  // 1 vs n
    "fr-FR": FrenchPlural,   // 0,1 vs n
    // ... more languages
}

func EnglishPlural(n int) int {
    if n == 1 {
        return 0  // singular
    }
    return 1  // plural
}
```

## Configuration Integration

**Language detection chain** (internal/config/config.go):

```go
type Config struct {
    Language         string `json:"language"`
    FallbackLanguage string `json:"fallback_language"`
    Version          string `json:"version"`
}

func Load() (*Config, error) {
    // 1. Load from ~/.config/clipet/config.json
    // 2. Override with environment variables:
    //    - CLIPET_LANG
    //    - LANG
    //    - LC_ALL
}
```

## TUI Integration

**Initialization flow** (internal/cli/root.go):

```go
func setup() error {
    // 1. Load config
    cfg, _ := config.Load()

    // 2. Create bundle
    bundle := i18n.NewBundle()
    loader := i18n.NewLoader(assets.LocalesFS, "locales")
    loader.LoadAll(bundle)

    // 3. Create manager
    i18nMgr := i18n.NewManager(cfg.Language, cfg.FallbackLanguage, bundle)

    // 4. Pass to TUI
    app := tui.NewApp(pet, registry, store, i18nMgr, offlineResults)
}
```

**Usage in TUI screens**:

```go
type HomeModel struct {
    i18n *i18n.Manager
    // ...
}

func (h HomeModel) someAction() {
    // Simple translation
    msg := h.i18n.T("ui.home.feed_success")

    // With interpolation
    msg := h.i18n.T("ui.home.feed_success",
        "oldHunger", 50,
        "newHunger", 75)
}
```

## Plugin Locale Support

**Plugin locale structure**:

```
internal/assets/builtins/cat-pack/
├── species.toml
├── dialogues.toml
├── adventures.toml
└── locales/
    ├── zh-CN.json
    └── en-US.json
```

**Plugin locale file format**:

```json
{
  "species": {
    "cat": {
      "name": "Cat",
      "description": "An agile little kitten..."
    }
  },
  "stages": {
    "egg": "Mysterious Egg",
    "baby": "Little Kitten"
  },
  "traits": {
    "purr_heal": {
      "name": "Purr Heal",
      "description": "Consume energy to heal..."
    }
  },
  "dialogues": {
    "baby": {
      "happy": ["Meow~", "Meow meow~"]
    }
  },
  "adventures": {
    "cat_fish_pond": {
      "name": "Mysterious Fish Pond",
      "choices": {
        "0": "Let it catch fish"
      }
    }
  }
}
```

## Registry Locale Integration

**Localized getters** (internal/plugin/registry.go):

```go
// Get localized dialogue
func (r *Registry) GetDialogue(speciesID, stageID, mood string) string {
    // 1. Try plugin locale first
    if pack.Locale != nil {
        dialogueKey := fmt.Sprintf("dialogues.%s.%s", stageID, mood)
        if lines := getLocaleArray(pack.Locale.Data, dialogueKey); len(lines) > 0 {
            return lines[rand.Intn(len(lines))]
        }
    }

    // 2. Fallback to inline TOML dialogues
    return selectFromTOMLDialogues(...)
}

// Similar for: GetAdventures, GetStage, GetTraitName, etc.
```

## ErrorType → i18n Bridge

**Game layer returns structured errors**:

```go
// internal/game/pet.go
func (p *Pet) Feed() ActionResult {
    if p.Energy < minEnergy {
        return failResultWithType(ErrEnergyLow, "精力不足")  // Internal log
    }
    // ...
}
```

**TUI layer generates localized messages**:

```go
// internal/tui/screens/home.go
func (h HomeModel) localizeGameError(res game.ActionResult) string {
    switch res.ErrorType {
    case game.ErrEnergyLow:
        return h.i18n.T("game.errors.energy_low")  // "Not enough energy!"
    case game.ErrDead:
        return h.i18n.T("game.errors.dead")  // "Your pet has passed away..."
    // ... 10 error types
    }
}
```

## File Statistics

- **i18n package**: 4 files, ~500 LoC
- **config package**: 1 file, ~150 LoC
- **TUI translations**: 120+ keys per language
- **Plugin translations (cat-pack)**: ~750 lines per language
- **Total i18n-related code**: ~2,000 lines

## Performance Characteristics

- **Startup cost**: ~5ms to load all locales (cached in memory)
- **Lookup cost**: O(depth) where depth = key segments (avg 3)
- **Memory overhead**: ~50KB per language loaded
- **Thread safety**: Full RWMutex protection

## Backward Compatibility

- Old plugins without locale files: ✅ Work with inline TOML text
- Old clipet with new plugins: ✅ Ignores locale files
- Partial translations: ✅ Falls back gracefully
