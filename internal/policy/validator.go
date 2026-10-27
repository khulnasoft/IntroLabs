package policy

import (
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

// Validator enforces safety policies on lab configurations.
type Validator struct{}

// NewValidator creates a new policy validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidationResult represents the result of policy validation.
type ValidationResult struct {
	Passed bool
	Errors []string
}

// ValidateCompose loads a Compose file and validates the resolved configuration.
func (v *Validator) ValidateCompose(composePath string) *ValidationResult {
	project, err := loader.Load(types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{
			{Filename: composePath},
		},
	})
	if err != nil {
		return &ValidationResult{
			Passed: false,
			Errors: []string{fmt.Sprintf("Failed to load Compose file: %v", err)},
		}
	}

	return v.ValidateComposeProject(project, composePath)
}

// ValidateComposeProject validates a resolved Compose project against safety policies.
// This inspects the effective configuration after interpolation, overrides, and
// defaults have been applied by compose-go.
func (v *Validator) ValidateComposeProject(project *types.Project, composePath string) *ValidationResult {
	result := &ValidationResult{
		Passed: true,
		Errors: []string{},
	}

	for _, service := range project.Services {
		// Check for privileged mode
		if service.Privileged {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' has privileged mode enabled (not allowed)", service.Name))
		}

		// Check for Docker socket mounts
		for _, mount := range service.Volumes {
			if isDockerSocketMount(mount) {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' mounts Docker socket (not allowed)", service.Name))
			}
			if isBroadHostMount(mount) {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' has broad host filesystem mount %s -> %s (not allowed)", service.Name, mount.Source, mount.Target))
			}
		}

		// Check for host networking
		if service.NetworkMode == "host" {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' uses host networking (not allowed)", service.Name))
		}

		// Check for host PID namespace
		if service.Pid == "host" {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' uses host PID namespace (not allowed)", service.Name))
		}

		// Check for host IPC namespace
		if service.Ipc == "host" {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' uses host IPC namespace (not allowed)", service.Name))
		}

		// Check for cap_add ALL
		for _, cap := range service.CapAdd {
			if cap == "ALL" {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' adds ALL capabilities (not allowed)", service.Name))
			}
		}

		// Check for read-only root filesystem (recommended)
		if !service.ReadOnly {
			// Recommend but don't fail
			result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' does not use read-only root filesystem (recommended)", service.Name))
		}

		// Check port bindings for non-loopback bindings
		for _, port := range service.Ports {
			hostIP := port.HostIP
			if hostIP == "" {
				hostIP = "0.0.0.0"
			}
			if hostIP != "127.0.0.1" && hostIP != "::1" && hostIP != "localhost" {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' publishes port %s to %s (must bind to 127.0.0.1)", service.Name, port.Published, hostIP))
			}
		}

		// Check for unpinned image references (no digest or versioned tag)
		if service.Image != "" && !isPinnedImage(service.Image) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' uses unpinned image reference '%s' (pin to version or digest)", service.Name, service.Image))
		}

		// Check for secrets in environment variables
		for envKey, envVal := range service.Environment {
			val := ""
			if envVal != nil {
				val = *envVal
			}
			if isSecretEnvVar(envKey) || isSecretEnvVar(val) {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' may contain committed secret in environment: %s=***", service.Name, envKey))
			}
		}

		// Check for security_opt that could reduce isolation
		for _, opt := range service.SecurityOpt {
			if opt == "label=disable" || opt == "apparmor=unconfined" || opt == "seccomp=unconfined" {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' uses security opt '%s' that reduces isolation (not allowed)", service.Name, opt))
			}
		}

		// Check for ulimits that could be excessive
		for ulimitName, ul := range service.Ulimits {
			if ul != nil && ulimitName == "nofile" && ul.Hard > 100000 {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("Service '%s' has excessive nofile ulimit %d (not allowed)", service.Name, ul.Hard))
			}
		}
	}

	// Check for undeclared external networks
	for _, net := range project.Networks {
		if net.External.External {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Network '%s' is external (undeclared external networks not allowed)", net.Name))
		}
	}

	// Check for secrets in top-level secrets definitions
	for _, secret := range project.Secrets {
		if secret.File != "" {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Secret '%s' references a file on the host (secrets must not reference host files)", secret.Name))
		}
	}

	return result
}

// isDockerSocketMount checks if a volume mount references the Docker socket.
func isDockerSocketMount(mount types.ServiceVolumeConfig) bool {
	source := mount.Source
	target := mount.Target
	return strings.Contains(source, "/var/run/docker.sock") ||
		strings.Contains(target, "/var/run/docker.sock")
}

// isBroadHostMount checks if a volume mount is a broad host filesystem mount.
// Mounts under /app (relative template paths) are allowed; absolute host paths
// outside the template are flagged.
func isBroadHostMount(mount types.ServiceVolumeConfig) bool {
	// Skip named volumes (no source path)
	if mount.Source == "" {
		return false
	}
	// Allow bind mounts that look like relative template paths
	if !filepathIsAbs(mount.Source) {
		return false
	}
	// Flag absolute host paths that are sensitive
	sensitive := []string{
		"/etc", "/var", "/usr", "/bin", "/sbin", "/lib", "/lib64",
		"/root", "/home", "/boot", "/proc", "/sys", "/dev",
	}
	for _, s := range sensitive {
		if mount.Source == s || strings.HasPrefix(mount.Source, s+"/") {
			return true
		}
	}
	return false
}

// filepathIsAbs reports whether path is absolute.
func filepathIsAbs(path string) bool {
	return len(path) > 0 && path[0] == '/'
}

// isPinnedImage checks if an image reference is pinned to a version or digest.
func isPinnedImage(image string) bool {
	// Digest references are always pinned
	if strings.Contains(image, "@sha256:") {
		return true
	}
	// Check for versioned tag (contains :something that is not latest)
	parts := strings.SplitN(image, ":", 2)
	if len(parts) == 2 {
		tag := parts[1]
		// Ignore port numbers in host:port/image format
		if strings.Contains(tag, "/") {
			return false
		}
		if tag != "latest" && tag != "" {
			return true
		}
	}
	return false
}

// isSecretEnvVar checks if an environment variable name or value looks like a secret.
func isSecretEnvVar(s string) bool {
	lower := strings.ToLower(s)
	secretKeywords := []string{"password", "passwd", "secret", "api_key", "apikey", "private_key", "token", "credential"}
	for _, kw := range secretKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// ValidateManifest checks a manifest against safety policies.
func (v *Validator) ValidateManifest(manifestPath string) *ValidationResult {
	result := &ValidationResult{
		Passed: true,
		Errors: []string{},
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to read manifest: %v", err))
		return result
	}

	content := string(data)

	if strings.Contains(content, "exposure: public") {
		result.Passed = false
		result.Errors = append(result.Errors, "Manifest declares public exposure (not allowed without explicit approval)")
	}

	// Check for committed secrets in environment
	if strings.Contains(content, "password:") || strings.Contains(content, "secret:") || strings.Contains(content, "api_key:") {
		// Only warn, don't fail - could be legitimate documentation
		result.Errors = append(result.Errors, "Manifest may contain committed secrets (review environment variables)")
	}

	return result
}
