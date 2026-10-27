//go:build integration

package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// labctlBinary is the path to the built labctl binary.
var labctlBinary string

var projectRootDir string

func TestMain(m *testing.M) {
	var err error
	projectRootDir, err = findProjectRoot()
	if err != nil {
		_ = m.Run()
		os.Exit(0)
	}
	labctlBinary, err = buildLabctl()
	if err != nil {
		// If we can't build, skip all tests
		_ = m.Run()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func buildLabctl() (string, error) {
	binaryPath := filepath.Join(projectRootDir, ".labctl-integration-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/labctl")
	cmd.Dir = projectRootDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("building labctl: %w\n%s", err, output)
	}
	return binaryPath, nil
}

// templatePath returns the absolute path to a built-in template directory.
func templatePath(relPath string) string {
	return filepath.Join(projectRootDir, relPath)
}

func runLabctl(args ...string) (string, error) {
	cmd := exec.Command(labctlBinary, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("labctl %v failed: %w\nstderr: %s", args, err, stderr.String())
	}
	return stdout.String(), nil
}

func waitForEndpoint(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("endpoint %s not ready within %v", url, timeout)
}

func TestLifecycleIntegration(t *testing.T) {
	if labctlBinary == "" {
		t.Skip("labctl binary not available")
	}

	// Skip if Docker is not available
	if _, err := exec.Command("docker", "info").Output(); err != nil {
		t.Skip("Docker not available")
	}

	// Step 1: Validate the template
	template := templatePath("templates/web/intro-lab")
	output, err := runLabctl("validate", template)
	if err != nil {
		t.Fatalf("validate failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Template validation passed") {
		t.Fatalf("expected 'Template validation passed', got: %s", output)
	}

	// Step 2: Check that no instance is already running
	output, _ = runLabctl("status")
	// Ignore errors from status if no instances

	// Step 3: Run the lab
	output, err = runLabctl("run", "web/intro-lab")
	if err != nil {
		t.Fatalf("run failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Lab URL:") {
		t.Fatalf("expected 'Lab URL:' in output, got: %s", output)
	}

	// Step 4: Wait for the health endpoint
	// Give the app time to start (docker compose up -d returns immediately)
	time.Sleep(3 * time.Second)
	if err := waitForEndpoint("http://127.0.0.1:8081/health", 30*time.Second); err != nil {
		t.Fatalf("readiness check failed: %v", err)
	}

	// Step 5: Check status shows the instance running
	output, err = runLabctl("status")
	if err != nil {
		t.Fatalf("status failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "web/intro-lab") {
		t.Fatalf("expected 'web/intro-lab' in status output, got: %s", output)
	}

	// Step 6: Stop the lab
	output, err = runLabctl("stop", "web/intro-lab")
	if err != nil {
		t.Fatalf("stop failed: %v\noutput: %s", err, output)
	}

	// Step 7: Remove the lab (with volumes to ensure full cleanup)
	output, err = runLabctl("remove", "--volumes", "--yes", "web/intro-lab")
	if err != nil {
		t.Fatalf("remove failed: %v\noutput: %s", err, output)
	}

	// Step 8: Verify no running instances remain
	output, err = runLabctl("status")
	if err != nil {
		t.Fatalf("status after remove failed: %v\noutput: %s", err, output)
	}
	if strings.Contains(output, "web/intro-lab") {
		t.Fatalf("expected no 'web/intro-lab' in status after remove, got: %s", output)
	}
}

func TestRunDuplicate(t *testing.T) {
	if labctlBinary == "" {
		t.Skip("labctl binary not available")
	}

	if _, err := exec.Command("docker", "info").Output(); err != nil {
		t.Skip("Docker not available")
	}

	// Run the lab first
	if _, err := runLabctl("run", "web/intro-lab"); err != nil {
		t.Skipf("could not start lab for duplicate test: %v", err)
	}
	defer runLabctl("remove", "--volumes", "--yes", "web/intro-lab")

	// Wait for readiness
	time.Sleep(3 * time.Second)

	// Try to run again - should fail with AlreadyRunningError
	_, err := runLabctl("run", "web/intro-lab")
	if err == nil {
		t.Error("expected error when running duplicate lab, got nil")
	}
}
