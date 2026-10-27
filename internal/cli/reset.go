package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset <lab-id>",
	Short: "Reset a lab instance (recreate containers)",
	Long:  `Stop and recreate the lab containers, preserving volumes. This will reset the application state.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resetLab(args[0])
	},
}

func resetLab(labID string) {
	if lifecycleMgr == nil {
		exitWithError("lifecycle manager not initialized", nil)
	}

	fmt.Printf("⚠️  This will reset the lab: %s\n", labID)
	fmt.Println("All containers will be recreated. Volumes will be preserved.")
	fmt.Print("Are you sure? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		exitWithError("Failed to read response", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" {
		fmt.Println("Reset cancelled.")
		return
	}

	fmt.Println("Resetting lab...")

	if err := lifecycleMgr.Reset(labID); err != nil {
		exitWithError("Failed to reset lab", err)
	}

	fmt.Println("Lab reset successfully.")
}
