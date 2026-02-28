// Package cli implements the cobra command tree for clipet.
package cli

import (
	"clipet/internal/assets"
	"clipet/internal/config"
	"clipet/internal/game"
	"clipet/internal/game/capabilities"
	"clipet/internal/i18n"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	registry         *plugin.Registry
	capabilitiesReg  *capabilities.Registry
	petStore         *store.JSONStore
	i18nMgr         *i18n.Manager
	cfg             *config.Config
)

// NewRootCmd creates the root cobra command.
// When invoked without subcommands, it launches the TUI.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "clipet",
		Short: "🐾 Clipet — 你的终端宠物伴侣",
		Long:  "Clipet 是一个运行在终端中的宠物养成程序。\n喂食、玩耍、对话、冒险，观看你的宠物成长进化。",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newInitCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newResetCmd())

	return root
}

// setup initializes the plugin registry and store.
func setup() error {
	// Load user configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		// Log warning but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = &config.Config{
			Language:         config.DefaultLanguage,
			FallbackLanguage: config.DefaultFallbackLanguage,
			Version:          config.DefaultVersion,
		}
	}

	// Initialize i18n
	bundle := i18n.NewBundle()
	loader := i18n.NewLoader(assets.LocalesFS, "locales")
	if err := loader.LoadAll(bundle); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load translations: %v\n", err)
	}

	i18nMgr = i18n.NewManager(cfg.Language, cfg.FallbackLanguage, bundle)

	// Initialize plugin registry with language
	registry = plugin.NewRegistry()
	registry.SetLanguage(cfg.Language, cfg.FallbackLanguage)

	// Initialize capabilities registry
	capabilitiesReg = capabilities.NewRegistry()

	// Load builtin species packs
	if err := registry.LoadFromFS(assets.BuiltinFS, "builtins", plugin.SourceBuiltin); err != nil {
		return fmt.Errorf("load builtin packs: %w", err)
	}

	// Load external plugins
	home, err := os.UserHomeDir()
	if err == nil {
		pluginsDir := filepath.Join(home, ".local", "share", "clipet", "plugins")
		if info, err := os.Stat(pluginsDir); err == nil && info.IsDir() {
			if err := registry.LoadFromFS(os.DirFS(pluginsDir), ".", plugin.SourceExternal); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to load external plugins: %v\n", err)
			}
		}
	}

	// Register all species traits to capabilities registry
	for _, speciesInfo := range registry.ListSpecies() {
		speciesPack := registry.GetSpecies(speciesInfo.ID)
		if speciesPack != nil && len(speciesPack.Traits) > 0 {
			if err := capabilitiesReg.RegisterTraits(speciesInfo.ID, speciesPack.Traits); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to register traits for %q: %v\n", speciesInfo.ID, err)
			}
		}
	}

	// Initialize time system with registries
	game.InitTimeSystem(registry, capabilitiesReg)

	// Initialize store
	petStore, err = store.NewJSONStore("")
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	return nil
}

// runTUI launches the Bubble Tea TUI application.
func runTUI() error {
	pet, err := loadPet()
	if err != nil {
		return err
	}

	// Apply accumulated offline duration and collect results
	var offlineResults []game.DecayRoundResult
	if pet.AccumulatedOfflineDuration > 0 {
		dur := pet.AccumulatedOfflineDuration

		// Adjust age
		pet.Birthday = pet.Birthday.Add(-dur)

		// Multi-stage settlement
		offlineResults = pet.ApplyMultiStageDecay(dur)

		// Trigger time hooks (lifecycle, death check, etc.)
		pet.AdvanceTime(dur)

		// Clear cache
		pet.AccumulatedOfflineDuration = 0
		pet.LastCheckedAt = time.Now()

		// Save state
		if err := petStore.Save(pet); err != nil {
			return fmt.Errorf("save after applying offline duration: %w", err)
		}
	}

	// Import TUI package and start with offline results (if any)
	return startTUI(pet, registry, petStore, offlineResults)
}

// loadPet loads the pet from store and sets its registry reference.
func loadPet() (*game.Pet, error) {
	if !petStore.Exists() {
		fmt.Println(i18nMgr.T("cli.status.no_pet"))
		return nil, fmt.Errorf("no pet")
	}

	pet, err := petStore.Load()
	if err != nil {
		return nil, fmt.Errorf("load pet: %w", err)
	}

	// Restore registry references (not serialized)
	pet.SetRegistry(registry)
	pet.SetCapabilitiesRegistry(capabilitiesReg)

	// Accumulate natural offline time (time since last check)
	pet.AccumulateOfflineTime()

	return pet, nil
}
