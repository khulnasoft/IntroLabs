package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <lab-id>",
	Short: "Stop a running lab instance",
	Long:  `Stop the services for a lab while retaining its state and volumes.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stopLab(args[0])
	},
}

func stopLab(labID string) {
	if lifecycleMgr == nil {
		exitWithError("lifecycle manager not initialized", nil)
	}

	if err := lifecycleMgr.Stop(labID); err != nil {
		exitWithError("Failed to stop lab", err)
	}

	fmt.Printf("Lab %s stopped successfully.\n", labID)
	fmt.Printf("Use 'labctl run %s' to start it again.\n", labID)
	fmt.Printf("Use 'labctl remove %s' to remove it completely.\n", labID)
}
