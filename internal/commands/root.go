package commands

import (
	"fmt"

	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/tui"
	"github.com/shotgum/stg/internal/version"
	"github.com/spf13/cobra"
)

// Root returns the root Cobra command.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:     "stg [category] [script] [args...]",
		Version: version.Version,
		Short:   "ShotGum — your personal script manager",
		Long: `ShotGum (stg) is a TUI-based script manager.

Run without arguments to open the interactive interface.
Run with a category name to list scripts in that category.
Run with a category and script name to execute directly.`,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load()
			if err != nil {
				return fmt.Errorf("loading registry: %w", err)
			}

			switch len(args) {
			case 0:
				// Open full TUI
				return tui.Start(reg)

			case 1:
				// Open TUI pre-navigated to category
				return tui.StartAtCategory(reg, args[0])

			default:
				// Direct run: stg <category> <script> [extra args...]
				category := args[0]
				scriptName := args[1]
				extraArgs := args[2:]

				entry, err := reg.FindScript(category, scriptName)
				if err != nil {
					return err
				}

				return runDirect(*entry, extraArgs, reg)
			}
		},
	}

	root.AddCommand(listCmd())
	root.AddCommand(addCmd())
	root.AddCommand(initCmd())

	return root
}
