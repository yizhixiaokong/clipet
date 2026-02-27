package main

import (
	"clipet/internal/cli"
	"clipet/internal/game"
	"fmt"
	"os"
)

func main() {
	game.InitTimeSystem()

	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
