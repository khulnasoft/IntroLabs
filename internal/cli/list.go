package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available labs",
	Long:  `List all available lab environments from the catalog.`,
	Run: func(cmd *cobra.Command, args []string) {
		listLabs()
	},
}

func listLabs() {
	if catResolver == nil {
		exitWithError("catalog resolver not initialized", nil)
	}

	labs := catResolver.List()
	if len(labs) == 0 {
		fmt.Println("No labs found in catalog.")
		return
	}

	fmt.Println("Available Labs")
	fmt.Println("==============")
	fmt.Println()
	fmt.Printf("%-20s %-30s %-12s %s\n", "ID", "Name", "Category", "Difficulty")
	fmt.Printf("%-20s %-30s %-12s %s\n", "--", "----", "--------", "-----------")
	for _, lab := range labs {
		fmt.Printf("%-20s %-30s %-12s %s\n", lab.ID, lab.Name, lab.Category, lab.Difficulty)
	}
	fmt.Println()
	fmt.Println("Use 'labctl info <id>' for more details about a specific lab.")
}
