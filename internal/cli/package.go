package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package <template-path>",
	Short: "Package a lab template for distribution",
	Long:  `Create a distributable archive of a lab template with checksums.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageTemplate(args[0])
	},
}

func packageTemplate(templatePath string) {
	if packager == nil {
		exitWithError("packager not initialized", nil)
	}

	fmt.Printf("Packaging template: %s\n", templatePath)
	fmt.Println()

	result, err := packager.Package(templatePath)
	if err != nil {
		exitWithError("Failed to package template", err)
	}

	fmt.Println()
	fmt.Println("✓ Package created successfully")
	fmt.Printf("Archive: %s\n", result.Archive)
	fmt.Printf("Checksum file: %s\n", result.ChecksumFile)
	fmt.Printf("SHA-256: %s\n", result.Checksum)
	fmt.Printf("Size: %d bytes\n", result.Size)
	os.Stdout.Sync()
}
