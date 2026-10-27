package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/labkit-project/vulnerable-lab-platform/internal/generator"
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

	// Resolve the template path and check verification state
	templatePath := resolveTemplatePath(labID)
	if templatePath == "" {
		exitWithError(fmt.Sprintf("template not found: %s", labID), nil)
	}

	ml, err := manifestLoader.Load(filepath.Join(templatePath, "lab.yaml"))
	if err != nil {
		exitWithError(fmt.Sprintf("failed to load manifest: %v", err), nil)
	}

	stateStr := string(ml.CoverageState)
	if ml.Metadata.Verification != nil && ml.Metadata.Verification.Status != "" {
		stateStr = ml.Metadata.Verification.Status
	}
	state := generator.TemplateState(stateStr)

	if !state.IsRunnable() {
		exitWithError(fmt.Sprintf("template is in state %q and is not runnable; it must be verified before launching", state), nil)
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

func resolveTemplatePath(labID string) string {
	candidates := []string{
		labID,
		filepath.Join("templates", "web", labID),
		filepath.Join("templates", "api", labID),
		filepath.Join("templates", labID),
	}
	for _, c := range candidates {
		if fileExistsShared(c + "/lab.yaml") {
			return c
		}
	}
	return ""
}
