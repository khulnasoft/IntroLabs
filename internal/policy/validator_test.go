package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(findProjectRoot(t), "internal", "policy", "testdata", name)
}

func TestValidateComposeValid(t *testing.T) {
	validator := NewValidator()

	composePath := filepath.Join(findProjectRoot(t), "templates", "web", "intro-lab", "compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Skipf("Compose file not found at %s", composePath)
	}

	result := validator.ValidateCompose(composePath)
	if !result.Passed {
		t.Errorf("Expected validation to pass for valid compose file, got errors: %v", result.Errors)
	}
}

func TestValidateComposeNonExistent(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose("nonexistent.yaml")
	if result.Passed {
		t.Error("Expected validation to fail for non-existent file")
	}
}

func TestValidateComposePrivileged(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "privileged.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for privileged container")
	}
	if !containsError(result.Errors, "privileged mode enabled") {
		t.Errorf("Expected 'privileged mode enabled' error, got: %v", result.Errors)
	}
}

func TestValidateComposeDockerSocket(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "docker-socket.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for Docker socket mount")
	}
	if !containsError(result.Errors, "mounts Docker socket") {
		t.Errorf("Expected 'mounts Docker socket' error, got: %v", result.Errors)
	}
}

func TestValidateComposeHostNetwork(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "host-network.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for host networking")
	}
	if !containsError(result.Errors, "uses host networking") {
		t.Errorf("Expected 'uses host networking' error, got: %v", result.Errors)
	}
}

func TestValidateComposeHostPID(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "host-pid.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for host PID namespace")
	}
	if !containsError(result.Errors, "uses host PID namespace") {
		t.Errorf("Expected 'uses host PID namespace' error, got: %v", result.Errors)
	}
}

func TestValidateComposeHostIPC(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "host-ipc.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for host IPC namespace")
	}
	if !containsError(result.Errors, "uses host IPC namespace") {
		t.Errorf("Expected 'uses host IPC namespace' error, got: %v", result.Errors)
	}
}

func TestValidateComposeNonLoopbackPort(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "non-loopback-port.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for non-loopback port binding")
	}
	if !containsError(result.Errors, "must bind to 127.0.0.1") {
		t.Errorf("Expected 'must bind to 127.0.0.1' error, got: %v", result.Errors)
	}
}

func TestValidateComposeUnpinnedImage(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "unpinned-image.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for unpinned image")
	}
	if !containsError(result.Errors, "unpinned image reference") {
		t.Errorf("Expected 'unpinned image reference' error, got: %v", result.Errors)
	}
}

func TestValidateComposeBroadHostMount(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "broad-host-mount.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for broad host filesystem mount")
	}
	if !containsError(result.Errors, "broad host filesystem mount") {
		t.Errorf("Expected 'broad host filesystem mount' error, got: %v", result.Errors)
	}
}

func TestValidateComposeSecurityOpt(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "security-opt.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for reduced isolation security opt")
	}
	if !containsError(result.Errors, "reduces isolation") {
		t.Errorf("Expected 'reduces isolation' error, got: %v", result.Errors)
	}
}

func TestValidateComposeCapAddAll(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "cap-add-all.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for cap_add ALL")
	}
	if !containsError(result.Errors, "adds ALL capabilities") {
		t.Errorf("Expected 'adds ALL capabilities' error, got: %v", result.Errors)
	}
}

func TestValidateComposeExternalNetwork(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "external-network.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for external network")
	}
	if !containsError(result.Errors, "is external") {
		t.Errorf("Expected 'is external' error, got: %v", result.Errors)
	}
}

func TestValidateComposeEnvSecret(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateCompose(testdataPath(t, "env-secret.yaml"))
	if result.Passed {
		t.Error("Expected validation to fail for committed secret in environment")
	}
	if !containsError(result.Errors, "committed secret") {
		t.Errorf("Expected 'committed secret' error, got: %v", result.Errors)
	}
}

func containsError(errors []string, substring string) bool {
	for _, e := range errors {
		if strings.Contains(e, substring) {
			return true
		}
	}
	return false
}

func TestValidateManifestValid(t *testing.T) {
	validator := NewValidator()

	manifestPath := filepath.Join(findProjectRoot(t), "templates", "web", "intro-lab", "lab.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skipf("Manifest file not found at %s", manifestPath)
	}

	result := validator.ValidateManifest(manifestPath)
	if !result.Passed {
		t.Errorf("Expected validation to pass for valid manifest, got errors: %v", result.Errors)
	}
}

func TestValidateManifestNonExistent(t *testing.T) {
	validator := NewValidator()
	result := validator.ValidateManifest("nonexistent.yaml")
	if result.Passed {
		t.Error("Expected validation to fail for non-existent file")
	}
}
