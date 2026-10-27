package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/labkit-project/vulnerable-lab-platform/internal/catalog"
	"github.com/labkit-project/vulnerable-lab-platform/internal/lifecycle"
	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
	"github.com/labkit-project/vulnerable-lab-platform/internal/packaging"
	"github.com/labkit-project/vulnerable-lab-platform/internal/state"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "labctl",
	Short: "LabKit CLI - Manage vulnerable security lab environments",
	Long: `LabKit is a platform for launching pre-built, ready-to-use vulnerable
environments for security education, training, and authorized research.

Use labctl to discover, launch, inspect, and manage lab environments.`,
}

// Shared dependencies initialized by Setup.
var (
	catResolver    *catalog.Resolver
	manifestLoader *manifest.Loader
	lifecycleMgr   *lifecycle.Manager
	stateStore     *state.Store
	packager       *packaging.Packager
	rootDir        string
)

// Setup initializes shared dependencies for all CLI commands.
func Setup(rd string) error {
	var err error

	rootDir = rd

	stateStore, err = state.NewStore(state.DefaultStateDir())
	if err != nil {
		return fmt.Errorf("creating state store: %w", err)
	}

	lifecycleMgr, err = lifecycle.NewManager(rootDir, stateStore)
	if err != nil {
		return fmt.Errorf("creating lifecycle manager: %w", err)
	}

	catResolver, err = catalog.NewResolver(rootDir)
	if err != nil {
		return fmt.Errorf("creating catalog resolver: %w", err)
	}

	schemaPath := filepath.Join(rootDir, "schemas", "lab.schema.json")
	manifestLoader, err = manifest.NewLoader(schemaPath)
	if err != nil {
		return fmt.Errorf("creating manifest loader: %w", err)
	}

	packager = packaging.NewPackager()

	return nil
}

func Run() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(packageCmd)
	rootCmd.AddCommand(versionCmd)
}

func exitWithError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
	os.Exit(1)
}
