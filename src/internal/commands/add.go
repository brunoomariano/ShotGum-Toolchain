package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/spf13/cobra"
)

// loadOrInitGlobalConfig loads the global config, creating a minimal default if
// it does not exist yet (before stg init has been run).
func loadOrInitGlobalConfig() (*config.Config, error) {
	cfg, err := config.LoadGlobal()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &config.Config{Version: "1", HelpFlag: "--help"}
	}
	return cfg, nil
}

func addCmd() *cobra.Command {
	add := &cobra.Command{
		Use:   "add",
		Short: "Add a folder (category) or script to ShotGum",
	}

	add.AddCommand(addFolderCmd())
	add.AddCommand(addScriptCmd())

	return add
}

func addFolderCmd() *cobra.Command {
	var name, desc string

	cmd := &cobra.Command{
		Use:   "folder <path>",
		Short: "Register a directory as a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			folderPath, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolving path: %w", err)
			}

			info, err := os.Stat(folderPath)
			if err != nil {
				return fmt.Errorf("accessing path %q: %w", folderPath, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("%q is not a directory", folderPath)
			}

			if name == "" {
				name = filepath.Base(folderPath)
			}

			globalCfg, err := loadOrInitGlobalConfig()
			if err != nil {
				return err
			}

			// Check for duplicate category name
			for _, c := range globalCfg.Categories {
				if c.Name == name {
					return fmt.Errorf("category %q already exists", name)
				}
			}

			cat := config.Category{
				Name:        name,
				Description: desc,
				ScriptsPath: folderPath,
			}
			globalCfg.Categories = append(globalCfg.Categories, cat)

			// Scan for .sh files and suggest registering them (best-effort).
			scripts, _ := filepath.Glob(filepath.Join(folderPath, "*.sh")) //nolint:errcheck // Glob only errors on malformed patterns; ours is static.

			if err := config.Save(globalCfg, config.GlobalConfigPath()); err != nil {
				return err
			}

			fmt.Printf("%s Category %s added (path: %s)\n",
				styles.LocalBadge.Render("✓"),
				styles.CategoryStyle.Render(name),
				styles.DescStyle.Render(folderPath),
			)

			if len(scripts) > 0 {
				fmt.Println(styles.DescStyle.Render(fmt.Sprintf(
					"\nFound %d script(s) in directory. To register them individually:", len(scripts),
				)))
				for _, s := range scripts {
					base := filepath.Base(s)
					scriptName := strings.TrimSuffix(base, ".sh")
					fmt.Printf("  stg add script %s --category %s --name %s\n", s, name, scriptName)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Category name (default: directory basename)")
	cmd.Flags().StringVar(&desc, "desc", "", "Category description")

	return cmd
}

func addScriptCmd() *cobra.Command {
	var category, name, desc, scriptExec string

	cmd := &cobra.Command{
		Use:   "script <path>",
		Short: "Register a single script under a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptPath, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolving path: %w", err)
			}

			if _, err := os.Stat(scriptPath); err != nil {
				return fmt.Errorf("accessing script %q: %w", scriptPath, err)
			}

			if category == "" {
				return fmt.Errorf("--category is required")
			}

			if name == "" {
				base := filepath.Base(scriptPath)
				name = strings.TrimSuffix(base, filepath.Ext(base))
			}

			globalCfg, err := loadOrInitGlobalConfig()
			if err != nil {
				return err
			}

			// Verify the category exists
			catFound := false
			for _, c := range globalCfg.Categories {
				if c.Name == category {
					catFound = true
					break
				}
			}
			if !catFound {
				return fmt.Errorf("category %q not found — create it first with: stg add folder <path> --name %s", category, category)
			}

			// Check for duplicate script name in category
			for _, s := range globalCfg.Scripts {
				if s.Category == category && s.Name == name {
					return fmt.Errorf("script %q already exists in category %q", name, category)
				}
			}

			script := config.Script{
				Name:        name,
				Category:    category,
				Description: desc,
				Executable:  scriptExec,
				Path:        scriptPath,
			}
			globalCfg.Scripts = append(globalCfg.Scripts, script)

			if err := config.Save(globalCfg, config.GlobalConfigPath()); err != nil {
				return err
			}

			fmt.Printf("%s Script %s added to category %s\n",
				styles.LocalBadge.Render("✓"),
				styles.ScriptStyle.Render(name),
				styles.CategoryStyle.Render(category),
			)
			return nil
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Category to add the script to (required)")
	cmd.Flags().StringVar(&name, "name", "", "Script name (default: filename without extension)")
	cmd.Flags().StringVar(&desc, "desc", "", "Script description")
	cmd.Flags().StringVar(&scriptExec, "executable", "", "Executable for this script (defaults to config default_executable)")

	return cmd
}
