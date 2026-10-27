package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var removeVolumes bool
var removeYes bool

var removeCmd = &cobra.Command{
	Use:   "remove <lab-id>",
	Short: "Remove a lab instance",
	Long:  `Stop and remove the lab containers. Use --volumes to also remove associated volumes.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		removeLab(args[0])
	},
}

func init() {
	removeCmd.Flags().BoolVar(&removeVolumes, "volumes", false, "Remove associated volumes")
	removeCmd.Flags().BoolVar(&removeYes, "yes", false, "Skip confirmation prompt")
}

func removeLab(labID string) {
	if lifecycleMgr == nil {
		exitWithError("lifecycle manager not initialized", nil)
	}

	fmt.Printf("⚠️  This will remove the lab: %s\n", labID)
	if removeVolumes {
		fmt.Println("All containers AND volumes will be removed. This cannot be undone.")
	} else {
		fmt.Println("All containers will be removed. Volumes will be preserved.")
	}

	if !removeYes {
		fmt.Print("Are you sure? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			exitWithError("Failed to read response", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Removal cancelled.")
			return
		}
	}

	fmt.Println("Removing lab...")

	if err := lifecycleMgr.Remove(labID, removeVolumes); err != nil {
		exitWithError("Failed to remove lab", err)
	}

	fmt.Println("Lab removed successfully.")
}
