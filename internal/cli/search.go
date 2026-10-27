package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for labs by category, tag, or keyword",
	Long:  `Search the lab catalog for environments matching your query.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchLabs(args[0])
	},
}

func searchLabs(query string) {
	if catResolver == nil {
		exitWithError("catalog resolver not initialized", nil)
	}

	results := catResolver.Search(query)
	if len(results) == 0 {
		fmt.Printf("No labs found matching: %s\n", query)
		return
	}

	fmt.Printf("Search results for: %s\n", query)
	fmt.Println()
	fmt.Printf("%-20s %-30s %-12s %s\n", "ID", "Name", "Category", "Difficulty")
	fmt.Printf("%-20s %-30s %-12s %s\n", "--", "----", "--------", "-----------")
	for _, lab := range results {
		fmt.Printf("%-20s %-30s %-12s %s\n", lab.ID, lab.Name, lab.Category, lab.Difficulty)
	}
	fmt.Println()
	fmt.Println("Use 'labctl info <id>' for more details.")
}
