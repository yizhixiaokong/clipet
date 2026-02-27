// Package cli implements the cobra command tree for clipet.
package cli

import (
	"clipet/internal/assets"
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	registry *plugin.Registry
	petStore *store.JSONStore
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
	root.AddCommand(newFeedCmd())
	root.AddCommand(newPlayCmd())
	root.AddCommand(newResetCmd())

	return root
}

// setup initializes the plugin registry and store.
func setup() error {
	// Initialize registry
	registry = plugin.NewRegistry()

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

	// Initialize time system with registry
	game.InitTimeSystem(registry)

	// Initialize store
	petStore, err = store.NewJSONStore("")
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	return nil
}

// runTUI launches the Bubble Tea TUI application.
func runTUI() error {
	if !petStore.Exists() {
		fmt.Println("还没有宠物呢！请先运行 clipet init 创建一只。")
		return nil
	}

	pet, err := petStore.Load()
	if err != nil {
		return fmt.Errorf("load pet: %w", err)
	}

	// Import TUI package here to avoid circular dependency in the future
	return startTUI(pet, registry, petStore)
}
