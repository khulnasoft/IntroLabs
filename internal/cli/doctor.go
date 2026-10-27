package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system prerequisites for running labs",
	Long:  `Check that Docker, Docker Compose, and other prerequisites are properly configured.`,
	Run: func(cmd *cobra.Command, args []string) {
		checkPrerequisites()
	},
}

func checkPrerequisites() {
	fmt.Println("LabKit Prerequisites Check")
	fmt.Println("============================")
	fmt.Println()

	allPassed := true

	// Check Docker
	fmt.Print("Docker: ")
	if checkCommand("docker", "--version") {
		fmt.Println("✓ Installed")
	} else {
		fmt.Println("✗ Not found")
		allPassed = false
	}

	// Check Docker Compose
	fmt.Print("Docker Compose: ")
	if checkCommand("docker", "compose", "version") {
		fmt.Println("✓ Installed")
	} else {
		fmt.Println("✗ Not found")
		allPassed = false
	}

	// Check architecture
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)

	// Check Docker daemon is running
	fmt.Print("Docker Daemon: ")
	if isDockerRunning() {
		fmt.Println("✓ Running")
	} else {
		fmt.Println("✗ Not running")
		allPassed = false
	}

	fmt.Println()
	if allPassed {
		fmt.Println("✓ All prerequisites met")
	} else {
		fmt.Println("✗ Some prerequisites are missing")
	}
}

func checkCommand(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	err := cmd.Run()
	return err == nil
}

func isDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}
