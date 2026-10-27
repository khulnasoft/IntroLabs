package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/labkit-project/vulnerable-lab-platform/internal/policy"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <template-path>",
	Short: "Validate a lab template",
	Long:  `Validate a lab template against the schema and check required files.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateTemplate(args[0])
	},
}

func validateTemplate(templatePath string) {
	fmt.Printf("Validating template: %s\n", templatePath)
	fmt.Println()

	allPassed := true

	// Check required files
	requiredFiles := []string{
		"lab.yaml",
		"compose.yaml",
		"README.md",
		"lessons/objectives.md",
		"lessons/concepts.md",
		"lessons/remediation.md",
		"tests/smoke.sh",
	}

	fmt.Println("Checking required files...")
	for _, file := range requiredFiles {
		filePath := filepath.Join(templatePath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("  ✗ %s (missing)\n", file)
			allPassed = false
		} else {
			fmt.Printf("  ✓ %s\n", file)
		}
	}

	// Validate lab.yaml against JSON schema
	fmt.Println("\nValidating lab.yaml manifest...")
	manifestPath := filepath.Join(templatePath, "lab.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ lab.yaml not found\n")
		allPassed = false
	} else {
		_, err := manifestLoader.Load(manifestPath)
		if err != nil {
			fmt.Printf("  ✗ Manifest validation failed: %v\n", err)
			allPassed = false
		} else {
			fmt.Println("  ✓ lab.yaml is valid against schema")
		}
	}

	// Validate compose.yaml
	fmt.Println("\nValidating compose.yaml...")
	composePath := filepath.Join(templatePath, "compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		fmt.Printf("  ✗ compose.yaml not found\n")
		allPassed = false
	} else {
		fmt.Println("  ✓ compose.yaml exists")

		// Validate the resolved Compose configuration against safety policies.
		// This inspects the effective configuration after interpolation,
		// overrides, and defaults have been applied (equivalent to
		// `docker compose config --format json`).
		data, err := os.ReadFile(composePath)
		if err != nil {
			fmt.Printf("  ✗ compose.yaml unreadable: %v\n", err)
			allPassed = false
		} else {
			project, err := loader.Load(types.ConfigDetails{
				ConfigFiles: []types.ConfigFile{
					{Filename: composePath, Content: data},
				},
				WorkingDir: filepath.Dir(composePath),
			}, func(o *loader.Options) {
				o.SetProjectName(filepath.Base(filepath.Dir(composePath)), true)
			})
			if err != nil {
				fmt.Printf("  ✗ compose.yaml failed to resolve: %v\n", err)
				allPassed = false
			} else {
				policyResult := policy.NewValidator().ValidateComposeProject(project, composePath)
				if !policyResult.Passed {
					allPassed = false
					fmt.Println("\n  Security policy violations:")
					for _, errMsg := range policyResult.Errors {
						fmt.Printf("    ✗ %s\n", errMsg)
					}
				} else {
					fmt.Println("  ✓ compose.yaml passes safety policies")
				}
			}
		}
	}

	// Check smoke test is executable
	smokeTestPath := filepath.Join(templatePath, "tests/smoke.sh")
	if info, err := os.Stat(smokeTestPath); err == nil {
		if info.Mode().Perm()&0111 != 0 {
			fmt.Println("\n✓ smoke.sh is executable")
		} else {
			fmt.Println("\n⚠ smoke.sh is not executable (run: chmod +x tests/smoke.sh)")
		}
	}

	fmt.Println()
	if allPassed {
		fmt.Println("✓ Template validation passed")
	} else {
		fmt.Println("✗ Template validation failed")
		os.Exit(1)
	}
}
