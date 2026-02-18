package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/shotgum/stg/internal/commands"
	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/runner"
)

// knownSubcommands are the Cobra-handled subcommands.
var knownSubcommands = map[string]bool{
	"list":       true,
	"add":        true,
	"init":       true,
	"help":       true,
	"completion": true,
	"--help":     true,
	"-h":         true,
}

func main() {
	args := os.Args[1:]

	// If we have at least 2 args where the first two are non-flag words not
	// matching known subcommands, treat as: stg <category> <script> [extra-args...]
	// This lets all extra flags (including --foo) pass through to the script.
	if len(args) >= 2 &&
		!strings.HasPrefix(args[0], "-") &&
		!strings.HasPrefix(args[1], "-") &&
		!knownSubcommands[args[0]] {

		reg, err := registry.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
			os.Exit(1)
		}

		entry, err := reg.FindScript(args[0], args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := runner.Run(*entry, args[2:], reg); err != nil {
			if runErr, ok := err.(*runner.RunError); ok {
				os.Exit(runErr.ExitCode)
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := commands.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
