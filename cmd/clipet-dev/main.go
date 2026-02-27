package main

import (
	"clipet/internal/assets"
	"clipet/internal/game"
	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	registry         *plugin.Registry
	capabilitiesReg  *capabilities.Registry
	petStore         *store.JSONStore
	packDir          string
)

func main() {
	root := &cobra.Command{
		Use:   "clipet-dev",
		Short: "Clipet developer tool",
		Long:  "clipet-dev is a Clipet plugin developer tool for timeskip, set, evolve, validate, and preview.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "validate" || cmd.Name() == "preview" {
				return nil
			}
			return setup()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&packDir, "pack-dir", "", "load species pack from directory")

	root.AddCommand(newTimeskipCmd())
	root.AddCommand(newSetCmd())
	root.AddCommand(newEvoCmd())
	root.AddCommand(newValidateCmd())
	root.AddCommand(newPreviewCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func setup() error {
	registry = plugin.NewRegistry()
	capabilitiesReg = capabilities.NewRegistry()

	if err := registry.LoadFromFS(assets.BuiltinFS, "builtins", plugin.SourceBuiltin); err != nil {
		return fmt.Errorf("load builtins: %w", err)
	}

	home, err := os.UserHomeDir()
	if err == nil {
		dir := filepath.Join(home, ".local", "share", "clipet", "plugins")
		if info, e := os.Stat(dir); e == nil && info.IsDir() {
			_ = registry.LoadFromFS(os.DirFS(dir), ".", plugin.SourceExternal)
		}
	}

	if packDir != "" {
		pack, err := plugin.ParsePack(os.DirFS(packDir), ".")
		if err != nil {
			return fmt.Errorf("parse pack-dir %q: %w", packDir, err)
		}
		pack.Source = plugin.SourceExternal
		registry.Register(pack)
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

	petStore, err = store.NewJSONStore("")
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	return nil
}

func requirePet() error {
	if !petStore.Exists() {
		return fmt.Errorf("no save file found, run clipet init first")
	}
	return nil
}

// loadPet loads the pet from store and sets its registry reference.
func loadPet() (*game.Pet, error) {
	if err := requirePet(); err != nil {
		return nil, err
	}

	pet, err := petStore.Load()
	if err != nil {
		return nil, fmt.Errorf("load save: %w", err)
	}

	// Restore registry references (not serialized)
	pet.SetRegistry(registry)
	pet.SetCapabilitiesRegistry(capabilitiesReg)

	return pet, nil
}
