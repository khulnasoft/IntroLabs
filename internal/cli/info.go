package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <lab-id>",
	Short: "Show detailed information about a lab",
	Long:  `Display requirements, safety details, and metadata for a specific lab.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showLabInfo(args[0])
	},
}

func showLabInfo(labID string) {
	if catResolver == nil || manifestLoader == nil {
		exitWithError("dependencies not initialized", nil)
	}

	lab := catResolver.Find(labID)
	if lab == nil {
		exitWithError(fmt.Sprintf("lab not found: %s", labID), nil)
	}

	templatePath, err := catResolver.TemplatePath(labID)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to resolve template path: %v", err), nil)
	}

	m, err := manifestLoader.LoadFromTemplate(templatePath)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to load manifest: %v", err), nil)
	}

	fmt.Printf("Lab: %s\n", lab.ID)
	fmt.Println()
	fmt.Printf("Name: %s\n", m.Metadata.Name)
	fmt.Printf("Version: %s\n", m.Metadata.Version)
	fmt.Printf("Category: %s\n", m.Metadata.Category)
	fmt.Printf("Difficulty: %s\n", m.Metadata.Difficulty)
	fmt.Printf("License: %s\n", m.Metadata.License)
	if m.Metadata.Author != "" {
		fmt.Printf("Author: %s\n", m.Metadata.Author)
	}
	fmt.Println()
	fmt.Printf("Description:\n  %s\n", m.Metadata.Description)
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Printf("  - Docker Engine %s\n", m.Runtime.DockerVersion)
	fmt.Printf("  - Docker Compose %s\n", m.Runtime.ComposeVersion)
	fmt.Printf("  - %d MiB RAM\n", m.Runtime.Resources.MemoryMiB)
	if m.Runtime.Resources.CPU > 0 {
		fmt.Printf("  - %.2f CPU cores\n", m.Runtime.Resources.CPU)
	}
	fmt.Println()
	fmt.Println("Safety:")
	fmt.Printf("  - Network exposure: %s\n", m.Runtime.Network.Exposure)
	fmt.Printf("  - Outbound access: %s\n", m.Runtime.Network.Outbound)
	fmt.Println()
	fmt.Println("Endpoints:")
	for _, ep := range m.Endpoints {
		fmt.Printf("  - %s: %s://%s:%d (HTTP)\n", ep.ID, ep.Protocol, ep.HostBinding, ep.ContainerPort)
		if ep.Path != "" {
			fmt.Printf("    Path: %s\n", ep.Path)
		}
	}
	fmt.Println()
	fmt.Println("Learning Objectives:")
	for _, obj := range m.Learning.Objectives {
		fmt.Printf("  - %s\n", obj)
	}
}
