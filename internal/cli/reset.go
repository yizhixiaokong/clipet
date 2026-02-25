package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "删除存档，重新开始",
		RunE:  runReset,
	}
	cmd.Flags().BoolP("yes", "y", false, "跳过确认直接删除")
	return cmd
}

func runReset(cmd *cobra.Command, args []string) error {
	if !petStore.Exists() {
		fmt.Println("no save file")
		return nil
	}

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Printf("delete %s? [y/N] ", petStore.Path())
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("cancelled")
			return nil
		}
	}

	if err := os.Remove(petStore.Path()); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("save deleted")
	return nil
}
