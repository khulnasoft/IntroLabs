package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <lab-id>",
	Short: "Display service logs for a lab",
	Long:  `Show logs from the services running in the specified lab environment.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showLogs(args[0])
	},
}

func showLogs(labID string) {
	if catResolver == nil {
		exitWithError("catalog resolver not initialized", nil)
	}

	templatePath, err := catResolver.TemplatePath(labID)
	if err != nil {
		exitWithError(fmt.Sprintf("Lab template not found: %s", labID), err)
	}

	composeFile := filepath.Join(templatePath, "compose.yaml")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		exitWithError(fmt.Sprintf("Compose file not found: %s", composeFile), err)
	}

	cmd := exec.Command("docker", "compose", "-f", composeFile, "logs", "-f")
	cmd.Dir = templatePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitWithError("Failed to retrieve logs", err)
	}
}
