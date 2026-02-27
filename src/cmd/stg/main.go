package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/commands"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/runner"
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
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {

	// If we have at least 2 args where the first two are non-flag words not
	// matching known subcommands, treat as: stg <category> <script> [extra-args...]
	// This lets all extra flags (including --foo) pass through to the script.
	if isDirectRun(args) {

		reg, err := registry.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
			return 1
		}

		entry, err := reg.FindScript(args[0], args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}

		if err := runner.Run(*entry, args[2:], reg); err != nil {
			if runErr, ok := err.(*runner.RunError); ok {
				return runErr.ExitCode
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	cmd := commands.Root()
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func isDirectRun(args []string) bool {
	if len(args) < 2 {
		return false
	}
	if strings.HasPrefix(args[0], "-") || strings.HasPrefix(args[1], "-") {
		return false
	}
	if knownSubcommands[args[0]] {
		return false
	}
	return true
}
