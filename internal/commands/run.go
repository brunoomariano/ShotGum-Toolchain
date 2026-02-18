package commands

import (
	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/runner"
)

// runDirect runs a script directly (no TUI), forwarding extra args to the script.
func runDirect(entry registry.ScriptEntry, args []string, reg *registry.Registry) error {
	return runner.Run(entry, args, reg)
}
