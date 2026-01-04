package main

import (
	"fmt"
	"os"

	"github.com/petergi/ebook-mechanic-cli/internal/cli"
	"github.com/petergi/ebook-mechanic-cli/internal/tui"
)

func main() {
	// If no arguments provided, run interactive TUI
	// Otherwise, run CLI mode
	if len(os.Args) == 1 {
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CLI mode
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
