package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize ShotGum configuration and directories",
		Long:  "Creates ~/.config/shotgum/config.yaml and ~/.shotgum/scripts/ on first run.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := config.GlobalConfigPath()

			// Check if already initialized
			if _, err := os.Stat(cfgPath); err == nil {
				fmt.Printf("%s ShotGum is already initialized at %s\n",
					styles.UserBadge.Render("ℹ"),
					styles.DescStyle.Render(cfgPath),
				)
				return nil
			}

			if err := config.EnsureDefault(); err != nil {
				return fmt.Errorf("initializing ShotGum: %w", err)
			}

			home, _ := os.UserHomeDir()
			scriptsHome := filepath.Join(home, ".shotgum", "scripts")

			fmt.Println(styles.TitleStyle.Render("ShotGum initialized!"))
			fmt.Println()
			fmt.Printf("  Config file:  %s\n", styles.DescStyle.Render(cfgPath))
			fmt.Printf("  Scripts home: %s\n", styles.DescStyle.Render(scriptsHome))
			fmt.Println()
			fmt.Println(styles.DescStyle.Render("Next steps:"))
			fmt.Println(styles.DescStyle.Render("  1. Add a script folder:   stg add folder ~/my-scripts --name shotgum"))
			fmt.Println(styles.DescStyle.Render("  2. Add a script:          stg add script ~/my-scripts/deploy.sh --category shotgum"))
			fmt.Println(styles.DescStyle.Render("  3. List your scripts:     stg list"))
			fmt.Println(styles.DescStyle.Render("  4. Open interactive TUI:  stg"))
			fmt.Println()
			fmt.Println(styles.HelpStyle.Render("Tip: Add SHOTGUM_HOME to your shell profile to customize the scripts directory."))

			return nil
		},
	}
}
