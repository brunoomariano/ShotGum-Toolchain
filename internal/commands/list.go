package commands

import (
	"fmt"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [category]",
		Short: "List categories or scripts in a category",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load()
			if err != nil {
				return fmt.Errorf("loading registry: %w", err)
			}

			if len(args) == 0 {
				return listCategories(reg)
			}
			return listScripts(reg, args[0])
		},
	}
}

func listCategories(reg *registry.Registry) error {
	cats := reg.GetCategories()
	if len(cats) == 0 {
		fmt.Println(styles.DescStyle.Render("No categories found. Run `stg init` to get started."))
		return nil
	}

	title := styles.TitleStyle.Render("Categories")
	sep := styles.DescStyle.Render(strings.Repeat("─", 50))
	fmt.Println(title)
	fmt.Println(sep)

	for _, c := range cats {
		badge := styles.Badge(c.Source)
		name := styles.CategoryStyle.Render(c.Name)
		desc := styles.DescStyle.Render(c.Description)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			fmt.Sprintf("  %s %-20s  %s", badge, name, desc),
		)
		fmt.Println(row)
	}
	fmt.Println()
	return nil
}

func listScripts(reg *registry.Registry, category string) error {
	scripts := reg.GetScripts(category)
	if len(scripts) == 0 {
		fmt.Printf("%s\n", styles.DescStyle.Render(fmt.Sprintf("No scripts found in category %q.", category)))
		return nil
	}

	title := styles.TitleStyle.Render(category) + styles.DescStyle.Render(" — scripts")
	sep := styles.DescStyle.Render(strings.Repeat("─", 50))
	fmt.Println(title)
	fmt.Println(sep)

	for _, s := range scripts {
		badge := styles.Badge(s.Source)
		name := styles.ScriptStyle.Render(s.Name)
		desc := styles.DescStyle.Render(s.Description)

		fmt.Printf("  %s %-20s  %s\n", badge, name, desc)
	}
	fmt.Println()
	return nil
}
