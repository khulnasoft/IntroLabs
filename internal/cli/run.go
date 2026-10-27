package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <lab-id>",
	Short: "Launch a lab environment",
	Long:  `Validate and start a new instance of the specified lab.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runLab(args[0])
	},
}

func runLab(labID string) {
	if lifecycleMgr == nil {
		exitWithError("lifecycle manager not initialized", nil)
	}

	result, err := lifecycleMgr.Run(labID)
	if err != nil {
		exitWithError("Failed to run lab", err)
	}

	fmt.Printf("Environment: %s\n", result.LabID)
	fmt.Printf("Version:     %s\n", result.Version)
	fmt.Printf("Runtime:     Docker Compose\n")
	fmt.Println()

	fmt.Println("[check] Docker is available")
	fmt.Println("[check] Compose configuration is valid")
	fmt.Println("[check] Environment safety policies passed")
	fmt.Println("[info]  Starting services...")
	fmt.Println("[check]  App is ready")
	fmt.Println()

	for _, url := range result.URLs {
		fmt.Printf("Lab URL: %s\n", url)
	}
	fmt.Println()
	fmt.Printf("Use 'labctl logs %s' to inspect logs.\n", result.LabID)
	fmt.Printf("Use 'labctl stop %s' to stop the environment.\n", result.LabID)

	os.Stdout.Sync()
}
