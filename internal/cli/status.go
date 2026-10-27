package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/labkit-project/vulnerable-lab-platform/internal/generator"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [template-id]",
	Short: "Show running lab instances or template verification status",
	Long:  `Display the status of all currently running lab environments. If a template-id is provided, show the template's verification status.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			showTemplateStatus(args[0])
			return
		}
		showStatus()
	},
}

func showTemplateStatus(templateID string) {
	manifestPath := resolveManifestPath(templateID)
	if manifestPath == "" {
		exitWithError(fmt.Sprintf("template not found: %s", templateID), nil)
	}

	ml, err := manifestLoader.Load(manifestPath)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to load manifest: %v", err), nil)
	}

	// Determine status from verification.status or coverage_state
	stateStr := string(ml.CoverageState)
	if ml.Metadata.Verification != nil && ml.Metadata.Verification.Status != "" {
		stateStr = ml.Metadata.Verification.Status
	}
	state := generator.TemplateState(stateStr)

	fmt.Printf("Status: %s\n", state)
	if ml.Metadata.CVE != "" {
		fmt.Printf("Source CVE: %s\n", ml.Metadata.CVE)
	}
	if ml.Metadata.Verification != nil {
		fmt.Printf("Verification: %s\n", ml.Metadata.Verification.Status)
		if len(ml.Metadata.Verification.Checks) > 0 {
			fmt.Printf("Checks: %v\n", ml.Metadata.Verification.Checks)
		}
	}
	if ml.Metadata.Review != nil {
		fmt.Printf("Review: %v\n", ml.Metadata.Review.Approved)
		if ml.Metadata.Review.Reviewer != "" {
			fmt.Printf("Reviewer: %s\n", ml.Metadata.Review.Reviewer)
		}
	}
	fmt.Printf("Publishable: %v\n", state.IsPublishable())
	fmt.Printf("Runnable: %v\n", state.IsRunnable())
}

func resolveManifestPath(templateID string) string {
	candidates := []string{
		templateID,
		filepath.Join("templates", "web", templateID),
		filepath.Join("templates", "api", templateID),
		filepath.Join("templates", templateID),
	}
	for _, c := range candidates {
		if fileExistsShared(c + "/lab.yaml") {
			return c + "/lab.yaml"
		}
	}
	return ""
}

func fileExistsShared(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func showStatus() {
	if lifecycleMgr == nil {
		exitWithError("lifecycle manager not initialized", nil)
	}

	instances := lifecycleMgr.Running()
	if len(instances) == 0 {
		fmt.Println("No running lab instances found.")
		return
	}

	fmt.Println("Running Lab Instances")
	fmt.Println("======================")
	fmt.Println()
	fmt.Printf("%-20s %-10s %-15s %s\n", "Instance ID", "Status", "Ports", "URLs")
	fmt.Printf("%-20s %-10s %-15s %s\n", "-----------", "------", "-----", "---")
	for _, inst := range instances {
		ports := ""
		for i, p := range inst.Ports {
			if i > 0 {
				ports += ", "
			}
			ports += fmt.Sprintf("%d", p)
		}
		urls := ""
		for i, u := range inst.URLs {
			if i > 0 {
				urls += ", "
			}
			urls += u
		}
		fmt.Printf("%-20s %-10s %-15s %s\n", inst.LabID, inst.Status, ports, urls)
	}
}
