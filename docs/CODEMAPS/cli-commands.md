<!-- Generated: 2026-02-27 | Files scanned: 8 | Token estimate: ~500 -->

# CLI Commands

## Main Binary: clipet

### Command Structure

```
clipet
├── [default] → TUI mode
├── init <name> [species]
├── status
├── feed
├── play
├── reset
└── evolve-check
```

### Implementation Map

| Command | File | Handler | Purpose |
|---------|------|---------|---------|
| root | cli/root.go | runTUI() | Launch Bubble Tea app |
| init | cli/init.go | runInit() | Create new pet |
| status | cli/status.go | runStatus() | Show pet status (CLI) |
| feed | cli/feed.go | runFeed() | Feed pet (CLI) |
| play | cli/play.go | runPlay() | Play with pet (CLI) |
| reset | cli/reset.go | runReset() | Delete save file |
| evolve-check | cli/evolve_check.go | runEvolveCheck() | Check evolution status |

### Initialization Flow (root.go)

```go
PersistentPreRunE: setup()
  ↓
1. plugin.NewRegistry()
2. registry.LoadFromFS(assets.BuiltinFS, "builtins", builtin)
3. registry.LoadFromFS(os.DirFS(pluginsDir), ".", external)
4. store.NewJSONStore("")
  ↓
Store in viper flags → accessible by all subcommands
```

### Individual Commands

#### init

```
Args: name (required), species (optional, default: "cat")
  ↓
petStore.Exists() → error if exists
  ↓
game.NewPet(name, species, stageID)
  ↓
petStore.Save(pet)
  ↓
Output: "创建成功！"
```

#### status

```
petStore.Load() → error if not exists
  ↓
Format pet info:
  - Name, Species, StageID
  - Age (hours/days)
  - Attributes with progress bars
  - Mood
  - Cooldowns
```

#### feed / play

```
petStore.Load()
  ↓
game.Feed(pet) / game.Play(pet)
  ↓
Check evolution → game.CheckEvolution()
  ↓
petStore.Save()
  ↓
Output result + evolution notification
```

#### reset

```
Confirm: "确定要重置吗？"
  ↓
petStore.Delete()
```

#### evolve-check

```
petStore.Load()
  ↓
game.CheckEvolution(pet, registry)
  ↓
Print available evolutions + conditions
```

## Dev Binary: clipet-dev

### Command Structure

```
clipet-dev
├── validate <pack-dir>
├── preview <pack-dir> [stage] [anim]
├── evo
│   ├── to [stage-id]
│   └── info
├── set [attr] [value]
└── timeskip [--hours N]
```

### Implementation Map

| Command | File | TUI Model | Purpose |
|---------|------|-----------|---------|
| validate | clipet-dev/validate.go | - | Validate pack structure |
| preview | clipet-dev/preview.go | dev.PreviewModel | Frame viewer |
| evo to | clipet-dev/evolve.go | dev.EvolveModel | Force evolution |
| evo info | clipet-dev/evolve.go | - | Show evolution info |
| set | clipet-dev/set.go | dev.SetModel | Edit attributes |
| timeskip | clipet-dev/timeskip.go | dev.TimeskipModel | Time simulation |

### Dev Commands Detail

#### validate

```
plugin.LoadPackFromFS(dir)
  ↓
plugin.ValidatePack(pack)
  ↓
Print validation results:
  - Missing required fields
  - Invalid references
  - Frame file errors
```

#### preview

```
Load species pack
  ↓
dev.NewPreviewModel(pack, fps, stage, anim)
  ↓
TUI: Left panel (animated frame) | Right panel (tree list)
  ↓
Key bindings:
  - ↑↓: Navigate tree
  - +/-: Adjust FPS
  - q/Esc: Quit
```

#### evo to

```
Args: stage-id (optional)
  ↓
If stage-id provided → direct evolution
Else → dev.NewEvolveModel() → TUI tree selection
  ↓
OnEvolve callback → force set StageID (no condition check)
  ↓
Output: "evolve: old -> new"
```

#### set

```
Args: attr value (optional)
  ↓
If both provided → direct set
Else → dev.NewSetModel() → TUI field editor
  ↓
Phases:
  1. Select field (↑↓)
  2. Edit value (Enter to save, Esc to cancel)
  ↓
Output: "set attr: old -> new"
```

#### timeskip

```
Flag: --hours (optional, default: 24)
  ↓
If --hours provided → direct simulation
Else → dev.NewTimeskipModel() → TUI
  ↓
Phases:
  1. Input hours
  2. Preview changes (confirm/cancel)
  ↓
Apply decay to pet
  ↓
Output: stats delta
```

## Shared Utilities

### tui_bridge.go

```go
startTUI(pet, registry, store)
  ↓
tea.NewProgram(tui.NewAppModel(...))
  ↓
Run Bubble Tea event loop
```

### Output Formatting

All commands use lipgloss for styled CLI output:
- Progress bars
- Status indicators
- Error messages (red)
- Success messages (green)
