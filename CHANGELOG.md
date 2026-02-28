# Changelog

All notable changes to the clipet project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.1.0] - 2025-02-28

### Added
- **Internationalization (i18n) System**
  - Complete i18n framework with support for multiple languages
  - Language detection: CLIPET_LANG > LANG > config file > default
  - Support for zh-CN (Chinese Simplified) and en-US (English)
  - Template interpolation with `{{.variable}}` syntax

- **Plugin Locale Support**
  - Plugins can provide `locales/{lang}.json` files
  - Fallback chain: requested language → fallback language → inline TOML
  - cat-pack includes complete translations for all content:
    - Species name and description
    - 17 stages (egg to legend)
    - 4 traits with names and descriptions
    - ~200 dialogue lines across all stages and moods
    - 13 adventure events with full localization

- **ErrorType System**
  - Structured error types in game layer (10 error types)
  - Standardized error handling: `ErrDead`, `ErrEnergyLow`, `ErrHealthLow`, etc.
  - Game layer remains UI-agnostic (no i18n dependency)
  - TUI layer generates localized error messages

- **Complete TUI Internationalization**
  - All TUI screens fully internationalized (home, adventure, evolve, offline_settlement)
  - All hardcoded strings replaced with i18n.T() calls
  - All game error messages support multiple languages
  - Menu items, help text, success/failure messages all localized

### Changed
- All `failResult()` calls in game layer replaced with `failResultWithType()`
- Game layer ActionResult extended with ErrorType field
- TUI layer uses `localizeGameError()` for all game errors
- Plugin registry methods now use locale data when available

### Technical Details
- **Files Created**:
  - `internal/i18n/` - Complete i18n package (4 files)
  - `internal/config/config.go` - Configuration system with language detection
  - `internal/assets/locales/` - TUI locale files (zh-CN, en-US)
  - `internal/assets/builtins/cat-pack/locales/` - Plugin locale files

- **Files Modified**:
  - `internal/game/pet.go` - Added ErrorType support (18 error cases)
  - `internal/game/adventure.go` - Added AdventureCheckResult struct
  - `internal/plugin/registry.go` - Added locale-aware getter methods
  - `internal/tui/screens/*.go` - All screens internationalized
  - Translation files with 120+ keys per language

- **Backward Compatibility**:
  - ✅ Old plugins without locale files continue to work
  - ✅ Fallback to inline TOML text when locale missing
  - ✅ Old clipet versions can use new plugins
  - ✅ Gradual migration path for plugin authors

## [3.0.0] - 2025-02-27

### Added
- **Custom Attributes System**
  - Plugins can define custom attributes in species.toml
  - Evolution conditions support custom attribute accumulators
  - Adventure effects can modify custom attributes
  - Pet.CustomAttributes map for runtime storage
  - Unified field API: `Pet.SetField("custom:attr_name", value)`

- **Multi-Stage Offline Settlement**
  - Time-based decay applied in multiple rounds
  - Critical state detection and warning
  - Detailed settlement report showing attribute changes per round
  - New offline_settlement TUI screen

- **Dynamic Cooldown System**
  - Cooldown duration scales with attribute urgency
  - Low attributes → short cooldown (10%)
  - Medium attributes → medium cooldown (50%)
  - High attributes → normal cooldown (100%)

### Changed
- **Evolution System Refactoring**
  - Removed hardcoded evolution conditions (night_bias, day_bias, etc.)
  - cat-pack now uses 10 custom attributes for 3 evolution paths
  - Evolution conditions defined via custom_acc in species.toml
  - More flexible and extensible evolution mechanics

- **Balance Adjustments**
  - Adjusted action values for better gameplay balance
  - Unified action default configurations
  - Reduced dialogue cooldown for better responsiveness

### Breaking Changes
- Evolution conditions now prefer custom attributes over hardcoded fields
- Old hardcoded evolution conditions (night_bias, day_bias) deprecated
- Plugins should migrate to custom attribute system for evolution paths

## [2.0.0] - Previous Release

Initial stable release with:
- Basic pet care mechanics (feed, play, rest, heal, talk)
- Adventure system with random events
- Evolution system with multiple paths
- Plugin system for species packs
- ASCII art animation
- TUI interface with Bubble Tea
