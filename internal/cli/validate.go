package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
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

	// Check required files. Supports both templates/<lab>/ and catalog/<vendor>/<product>/<cve>/ structures.
	requiredFiles := []string{
		"lab.yaml",
		"compose.yaml",
		"README.md",
	}

	// CVE catalog structure uses checks/smoke.sh, template structure uses tests/smoke.sh
	smokeTestPaths := []string{
		"checks/smoke.sh",
		"tests/smoke.sh",
	}

	// Template structure requires lessons/
	lessonFiles := []string{
		"lessons/objectives.md",
		"lessons/concepts.md",
		"lessons/remediation.md",
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

	// Check smoke test exists in either location
	smokeFound := false
	for _, smokePath := range smokeTestPaths {
		if _, err := os.Stat(filepath.Join(templatePath, smokePath)); err == nil {
			fmt.Printf("  ✓ %s\n", smokePath)
			smokeFound = true
			break
		}
	}
	if !smokeFound {
		fmt.Printf("  ✗ smoke test not found (checked: %v)\n", smokeTestPaths)
		allPassed = false
	}

	// Check for CVE-specific references.yaml
	refsPath := filepath.Join(templatePath, "references.yaml")
	if _, err := os.Stat(refsPath); os.IsNotExist(err) {
		// references.yaml is optional for non-CVE labs
	} else {
		fmt.Printf("  ✓ references.yaml\n")
	}

	// Check lessons only if they would be expected (template structure)
	hasLessons := false
	for _, lessonFile := range lessonFiles {
		if _, err := os.Stat(filepath.Join(templatePath, lessonFile)); err == nil {
			hasLessons = true
			break
		}
	}
	if hasLessons {
		for _, lessonFile := range lessonFiles {
			filePath := filepath.Join(templatePath, lessonFile)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				fmt.Printf("  ✗ %s (missing)\n", lessonFile)
				allPassed = false
			} else {
				fmt.Printf("  ✓ %s\n", lessonFile)
			}
		}
	}

	// Validate lab.yaml against JSON schema
	fmt.Println("\nValidating lab.yaml manifest...")
	manifestPath := filepath.Join(templatePath, "lab.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ lab.yaml not found\n")
		allPassed = false
	} else {
		labManifest, err := manifestLoader.Load(manifestPath)
		if err != nil {
			fmt.Printf("  ✗ Manifest validation failed: %v\n", err)
			allPassed = false
		} else {
			fmt.Println("  ✓ lab.yaml is valid against schema")

			// CVE-specific validation
			if labManifest.Metadata.CVE != "" {
				fmt.Println("\n  Validating CVE catalog fields...")
				if labManifest.Metadata.AffectedVersions == nil || len(labManifest.Metadata.AffectedVersions) == 0 {
					fmt.Printf("    ✗ CVE specified but affected_versions is empty\n")
					allPassed = false
				} else {
					fmt.Printf("    ✓ affected_versions declared: %v\n", labManifest.Metadata.AffectedVersions)
				}
				if labManifest.Metadata.FixedVersions == nil || len(labManifest.Metadata.FixedVersions) == 0 {
					fmt.Printf("    ✗ CVE specified but fixed_versions is empty\n")
					allPassed = false
				} else {
					fmt.Printf("    ✓ fixed_versions declared: %v\n", labManifest.Metadata.FixedVersions)
				}
				if labManifest.CoverageState == "" {
					fmt.Printf("    ✗ coverage_state is not set\n")
					allPassed = false
				} else {
					fmt.Printf("    ✓ coverage_state: %s\n", labManifest.CoverageState)
				}
				if labManifest.CoverageState == manifest.CoveragePublished && labManifest.Vulnerability == nil {
					fmt.Printf("    ✗ coverage_state is 'published' but vulnerability section is missing\n")
					allPassed = false
				}
				if labManifest.Vulnerability != nil && labManifest.Vulnerability.Provenance == nil {
					fmt.Printf("    ✗ vulnerability section missing provenance\n")
					allPassed = false
				}

				// Check references.yaml exists for CVE labs
				refsPath := filepath.Join(templatePath, "references.yaml")
				if _, err := os.Stat(refsPath); os.IsNotExist(err) {
					fmt.Printf("    ✗ references.yaml missing for CVE lab\n")
					allPassed = false
				} else {
					fmt.Printf("    ✓ references.yaml present\n")
				}
			}
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
				projectName := filepath.Base(filepath.Dir(composePath))
				projectName = strings.Map(func(r rune) rune {
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
						return r
					}
					return '-'
				}, projectName)
				if len(projectName) > 0 && projectName[0] < '0' {
					projectName = "lab" + projectName
				}
				o.SetProjectName(projectName, true)
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

	// Check smoke test is executable (either checks/smoke.sh or tests/smoke.sh)
	smokeFound = false
	for _, smokePath := range smokeTestPaths {
		smokeTestPath := filepath.Join(templatePath, smokePath)
		if info, err := os.Stat(smokeTestPath); err == nil {
			if info.Mode().Perm()&0111 != 0 {
				fmt.Printf("\n✓ %s is executable\n", smokePath)
			} else {
				fmt.Printf("\n⚠ %s is not executable (run: chmod +x %s)\n", smokePath, smokePath)
			}
			smokeFound = true
			break
		}
	}
	if !smokeFound {
		fmt.Printf("\n⚠ smoke test not found (checked: %v)\n", smokeTestPaths)
	}

	fmt.Println()
	if allPassed {
		fmt.Println("✓ Template validation passed")
	} else {
		fmt.Println("✗ Template validation failed")
		os.Exit(1)
	}
}
