package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags.
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the labctl version",
	Long:  `Display the version of the labctl CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("labctl version %s\n", version)
		fmt.Printf("built with %s\n", runtime.Version())
	},
}
