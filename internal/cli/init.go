package cli

import (
	"clipet/internal/game"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a new pet",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	if petStore.Exists() {
		return fmt.Errorf(i18nMgr.T("cli.init.pet_exists", "path", petStore.Path()))
	}

	species := registry.ListSpecies()
	if len(species) == 0 {
		return fmt.Errorf(i18nMgr.T("cli.init.no_species"))
	}

	// Sort species by name
	sort.Slice(species, func(i, j int) bool {
		return species[i].Name < species[j].Name
	})

	// Display available species
	fmt.Println(i18nMgr.T("cli.init.welcome"))
	fmt.Println()
	fmt.Println(i18nMgr.T("cli.init.available_species"))
	for i, s := range species {
		source := ""
		if s.Source == "external" {
			source = i18nMgr.T("cli.init.external_plugin")
		}
		fmt.Printf("  %d. %s — %s%s\n", i+1, s.Name, s.Description, source)
	}
	fmt.Println()

	// Get species choice
	var choice int
	for {
		fmt.Printf(i18nMgr.T("cli.init.select_species", "count", len(species)))
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > len(species) {
			fmt.Println(i18nMgr.T("cli.init.invalid_selection"))
			continue
		}
		break
	}
	selected := species[choice-1]

	// Get pet name
	var name string
	for {
		fmt.Print(i18nMgr.T("cli.init.enter_name"))
		_, err := fmt.Scanln(&name)
		if err != nil || name == "" {
			fmt.Println(i18nMgr.T("cli.init.name_empty"))
			continue
		}
		break
	}

	// Get base stats and egg stage
	baseStats := registry.GetBaseStats(selected.ID)
	eggStage := registry.GetEggStage(selected.ID)
	if baseStats == nil || eggStage == nil {
		return fmt.Errorf(i18nMgr.T("cli.init.incomplete_species", "species", selected.ID))
	}

	// Create pet
	pet := game.NewPet(name, selected.ID, eggStage.ID,
		baseStats.Hunger, baseStats.Happiness, baseStats.Health, baseStats.Energy, registry)
	pet.SetCapabilitiesRegistry(capabilitiesReg)

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf(i18nMgr.T("cli.init.save_failed", "error", err.Error()))
	}

	fmt.Println()
	fmt.Println(i18nMgr.T("cli.init.pet_created", "name", name, "stage", eggStage.Name))
	fmt.Println(i18nMgr.T("cli.init.species_label", "species", selected.Name))
	fmt.Println(i18nMgr.T("cli.init.stage_label", "stage", eggStage.Name))
	fmt.Println()
	fmt.Println(i18nMgr.T("cli.init.run_hint"))

	return nil
}
