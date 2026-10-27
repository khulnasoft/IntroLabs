package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Template management commands",
	Long:  `Commands for creating and managing lab templates.`,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new lab template from the scaffold",
	Long:  `Generate a new lab template by copying the scaffold directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createTemplate(args[0])
	},
}

func init() {
	templateCmd.AddCommand(templateCreateCmd)
}

func createTemplate(name string) {
	scaffoldPath := "templates/_scaffold"
	newTemplatePath := filepath.Join("templates", name)

	// Check if scaffold exists
	if _, err := os.Stat(scaffoldPath); os.IsNotExist(err) {
		exitWithError("Scaffold not found", err)
	}

	// Check if template already exists
	if _, err := os.Stat(newTemplatePath); !os.IsNotExist(err) {
		exitWithError(fmt.Sprintf("Template already exists: %s", newTemplatePath), nil)
	}

	fmt.Printf("Creating new template: %s\n", name)
	fmt.Println("Copying scaffold...")

	// Copy scaffold to new location
	// For simplicity, we'll use a shell command (in production, use proper copy)
	cmd := []string{"cp", "-r", scaffoldPath, newTemplatePath}
	if err := runCommand(cmd[0], cmd[1:]...); err != nil {
		exitWithError("Failed to copy scaffold", err)
	}

	fmt.Println("✓ Template created successfully")
	fmt.Printf("Location: %s\n", newTemplatePath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("1. Edit %s/lab.yaml to update metadata\n", newTemplatePath)
	fmt.Printf("2. Modify %s/compose.yaml for your services\n", newTemplatePath)
	fmt.Printf("3. Update documentation in %s/lessons/\n", newTemplatePath)
	fmt.Printf("4. Run: labctl validate %s\n", newTemplatePath)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
