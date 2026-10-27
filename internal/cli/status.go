package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show running lab instances",
	Long:  `Display the status of all currently running lab environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
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
