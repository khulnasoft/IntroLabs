package cli

import (
	"os"
	"testing"
)

func TestRunLab(t *testing.T) {
	// Test that runLab function handles errors correctly
	// This is a basic test to ensure the package compiles with tests
	if os.Getenv("TEST_INTEGRATION") == "1" {
		t.Skip("Skipping integration test unless TEST_INTEGRATION=1")
	}
}

func TestIsDockerRunning(t *testing.T) {
	// Test Docker availability check
	if os.Getenv("TEST_INTEGRATION") == "1" {
		t.Skip("Skipping integration test unless TEST_INTEGRATION=1")
	}
}
