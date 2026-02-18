// Package commands implements the Cobra CLI sub-commands for ShotGum:
// the root dispatcher, init, add, and list. The root command also handles
// pre-navigation and direct script execution without opening the TUI.
package commands

import (
	"fmt"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/version"
	"github.com/spf13/cobra"
)

// Root returns the root Cobra command.
func Root() *cobra.Command {
	var showExecutionLogs bool
	var printRepoURL bool

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
			if printRepoURL {
				fmt.Println(version.RepoURL)
				return nil
			}

			reg, err := registry.Load()
			if err != nil {
				return fmt.Errorf("loading registry: %w", err)
			}

			switch len(args) {
			case 0:
				// Open full TUI
				return tui.StartWithLogs(reg, showExecutionLogs)

			case 1:
				// Open TUI pre-navigated to category
				return tui.StartAtCategoryWithLogs(reg, args[0], showExecutionLogs)

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
	root.PersistentFlags().BoolVar(&showExecutionLogs, "logs", false, "Show ShotGum execution logs in a TUI footer")
	root.Flags().BoolVar(&printRepoURL, "repo-url", false, "Print the canonical repository URL and exit")
	_ = root.Flags().MarkHidden("repo-url")

	root.AddCommand(listCmd())
	root.AddCommand(addCmd())
	root.AddCommand(initCmd())

	return root
}
