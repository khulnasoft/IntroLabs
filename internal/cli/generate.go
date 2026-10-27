package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labkit-project/vulnerable-lab-platform/internal/generator"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate lab templates from external sources",
	Long:  `Generate lab templates from external sources such as CVE advisories.`,
}

var generateCVECmd = &cobra.Command{
	Use:   "cve <CVE-ID>",
	Short: "Generate a lab template from a CVE advisory",
	Long: `Generate a scaffolded LabKit template from a CVE identifier.

The generator fetches CVE metadata from NVD, normalizes it, and scaffolds
a complete lab template directory. The template starts in draft state and
must pass verification checks before it can be marked verified or published.

Examples:
  labctl generate cve CVE-2026-34486 --output templates/web/cve-2026-34486 --advisory https://nvd.nist.gov/vuln/detail/CVE-2026-34486 --app-family tomcat
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cveID := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		advisoryURL, _ := cmd.Flags().GetString("advisory")
		appFamily, _ := cmd.Flags().GetString("app-family")

		if outputPath == "" {
			outputPath = filepath.Join("templates", "web", strings.ToLower(cveID))
		}

		gen := generator.NewGenerator(rootDir)
		result, err := gen.GenerateCVE(cveID, outputPath, advisoryURL, appFamily)
		if err != nil {
			exitWithError("Failed to generate template", err)
		}

		fmt.Printf("✓ Template generated successfully\n")
		fmt.Printf("  Location: %s\n", result.TemplatePath)
		fmt.Printf("  State: %s\n", result.State)
		fmt.Printf("  CVE: %s\n", result.Normalized.CVEID)
		fmt.Printf("  Component: %s\n", result.Normalized.Component)
		fmt.Printf("  Runtime: %s\n", result.Normalized.RequiredRuntime)
		fmt.Printf("  Affected versions: %v\n", result.Normalized.AffectedVersions)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("1. Review %s/lab.yaml and %s/compose.yaml\n", result.TemplatePath, result.TemplatePath)
		fmt.Printf("2. Run: labctl validate %s\n", result.TemplatePath)
		fmt.Printf("3. After verification passes: labctl package %s\n", result.TemplatePath)
		fmt.Printf("4. Publish (verified + approval): labctl publish %s --confirm\n", result.TemplatePath)
	},
}

var publishCmd = &cobra.Command{
	Use:   "publish <template-path>",
	Short: "Publish a verified template to the catalog",
	Long: `Publish a verified lab template to the catalog.

Only templates in the verified state can be published. Requires --confirm
for manual approval.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templatePath := args[0]
		confirm, _ := cmd.Flags().GetBool("confirm")

		if !confirm {
			exitWithError("publishing requires manual approval", nil)
		}

		if _, err := os.Stat(templatePath); err != nil {
			exitWithError("template path not found", err)
		}

		gen := generator.NewGenerator(rootDir)

		// Load manifest to check state and get metadata
		ml, err := manifestLoader.Load(filepath.Join(templatePath, "lab.yaml"))
		if err != nil {
			exitWithError("failed to load manifest", err)
		}

		state := generator.TemplateState(ml.CoverageState)
		if !state.IsPublishable() {
			exitWithError(fmt.Sprintf("template is in state %q; must be verified to publish", state), nil)
		}

		// Package
		pkgResult, err := gen.PackageTemplate(templatePath, ml.Metadata.Version)
		if err != nil {
			exitWithError("failed to package template", err)
		}

		// Catalog entry
		entry := generator.CatalogEntry{
			CVEID:              ml.Metadata.CVE,
			TemplateID:         generator.NormalizeTemplateID(templatePath),
			Version:            ml.Metadata.Version,
			Technology:         ml.Metadata.Category,
			AffectedComponent:  ml.Metadata.Product,
			Category:           ml.Metadata.Category,
			Difficulty:         ml.Metadata.Difficulty,
			VerificationStatus: string(state),
			PublishedAt:        nowFunc(),
			Checksum:           pkgResult.Checksum,
			Archive:            pkgResult.Archive,
		}

		catalogPath := filepath.Join(rootDir, "catalog", "index.yaml")
		if err := gen.WriteCatalogEntry(catalogPath, entry); err != nil {
			exitWithError("failed to write catalog entry", err)
		}

		fmt.Printf("✓ Template published\n")
		fmt.Printf("  Template: %s\n", templatePath)
		fmt.Printf("  Archive: %s\n", pkgResult.Archive)
		fmt.Printf("  Checksum: %s\n", pkgResult.Checksum)
		fmt.Printf("  Catalog: %s\n", catalogPath)
	},
}

func init() {
	generateCVECmd.Flags().String("output", "", "output directory for scaffolded template (default: templates/web/<cve-id>)")
	generateCVECmd.Flags().String("advisory", "", "advisory/NVD URL source")
	generateCVECmd.Flags().String("app-family", "", "application family hint (e.g., tomcat, nginx)")
	generateCmd.AddCommand(generateCVECmd)

	publishCmd.Flags().Bool("confirm", false, "confirm manual approval for publishing")
	rootCmd.AddCommand(publishCmd)

	rootCmd.AddCommand(generateCmd)
}

var nowFunc = func() time.Time { return time.Now() }
