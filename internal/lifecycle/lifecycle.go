package lifecycle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/labkit-project/vulnerable-lab-platform/internal/catalog"
	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
	"github.com/labkit-project/vulnerable-lab-platform/internal/policy"
	"github.com/labkit-project/vulnerable-lab-platform/internal/state"
)

// Manager orchestrates the lab lifecycle (run, stop, reset, remove).
type Manager struct {
	catalog  *catalog.Resolver
	manifest *manifest.Loader
	policy   *policy.Validator
	store    *state.Store
	rootDir  string
}

// NewManager creates a lifecycle manager.
func NewManager(rootDir string, store *state.Store) (*Manager, error) {
	catResolver, err := catalog.NewResolver(rootDir)
	if err != nil {
		return nil, fmt.Errorf("creating catalog resolver: %w", err)
	}

	schemaPath := filepath.Join(rootDir, "schemas", "lab.schema.json")
	manifestLoader, err := manifest.NewLoader(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("creating manifest loader: %w", err)
	}

	return &Manager{
		catalog:  catResolver,
		manifest: manifestLoader,
		policy:   policy.NewValidator(),
		store:    store,
		rootDir:  rootDir,
	}, nil
}

// Run validates and starts a lab instance.
func (m *Manager) Run(labID string) (*RunResult, error) {
	templatePath, err := m.catalog.TemplatePath(labID)
	if err != nil {
		return nil, err
	}

	// Load and validate manifest
	labManifest, err := m.manifest.LoadFromTemplate(templatePath)
	if err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	// Load and validate compose file
	composePath := filepath.Join(templatePath, "compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("compose.yaml not found in template: %s", templatePath)
	}

	composeProject, err := loadComposeProject(composePath)
	if err != nil {
		return nil, fmt.Errorf("compose validation failed: %w", err)
	}

	// Policy validation on compose (against the resolved configuration)
	policyResult := m.policy.ValidateComposeProject(composeProject, composePath)
	if !policyResult.Passed {
		return nil, &PolicyError{Errors: policyResult.Errors}
	}

	// Check for existing instance
	if existing := m.store.Get(labID); existing != nil && existing.Status == "running" {
		return nil, &AlreadyRunningError{LabID: labID}
	}

	// Start services with docker compose
	// Note: port preflight checking is an early warning only. Docker's actual
	// bind failure during `up` is the authoritative result.
	if err := runDockerComposeCapture(templatePath, "up", "-d"); err != nil {
		// Cleanup on failure: stop and remove containers
		_ = runDockerCompose(templatePath, "down")
		if strings.Contains(err.Error(), "Bind for 0.0.0.0") || strings.Contains(err.Error(), "bind: address already in use") {
			return nil, &BindFailureError{Err: err}
		}
		return nil, fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for readiness if a health check is declared
	if labManifest.Checks.HealthCheck != nil && labManifest.Checks.HealthCheck.Endpoint != "" {
		healthURLs := buildHealthURLs(labManifest.Checks.HealthCheck.Endpoint, allocatedHostPorts(composeProject))
		ready, err := waitForReadiness(context.Background(), healthURLs, labManifest.Checks.HealthCheck)
		if err != nil || !ready {
			// Cleanup on readiness failure
			_ = runDockerCompose(templatePath, "down")
			if err != nil {
				return nil, &ReadinessError{Err: err, URLs: healthURLs}
			}
			return nil, &ReadinessError{Err: fmt.Errorf("services did not become ready within timeout"), URLs: healthURLs}
		}
	}

	// Determine final URLs (use actual ports from compose)
	ports := allocatedHostPorts(composeProject)
	urls := buildURLs(labManifest.Endpoints, ports)

	// Record instance state
	projectName := composeProjectName(labID)
	inst := m.store.Create(labID, labManifest.Metadata.Version, templatePath, composePath, projectName, ports, urls)

	return &RunResult{
		LabID:     labID,
		Version:   labManifest.Metadata.Version,
		Status:    inst.Status,
		Ports:     inst.Ports,
		URLs:      inst.URLs,
		StartedAt: inst.StartedAt,
	}, nil
}

// buildHealthURLs builds health check URLs from the endpoint and actual allocated ports.
func buildHealthURLs(endpoint string, usedPorts []int) []string {
	if len(usedPorts) == 0 {
		return []string{endpoint}
	}

	// Try to find a port placeholder in the endpoint and substitute actual ports
	var urls []string
	for _, port := range usedPorts {
		candidate := endpoint
		replaced := false
		for _, old := range []string{"${PORT}", "{port}", ":8080", ":8081"} {
			if strings.Contains(candidate, old) {
				candidate = strings.ReplaceAll(candidate, old, fmt.Sprintf(":%d", port))
				replaced = true
			}
		}
		if replaced {
			urls = append(urls, candidate)
		}
	}

	if len(urls) == 0 {
		urls = []string{endpoint}
	}
	return urls
}

// waitForReadiness polls the health check endpoint until it returns success or
// the context/deadline is cancelled.
//
// Readiness contract:
//   - Connection refused: continue polling until the deadline.
//   - HTTP 3xx: handled per CheckRedirect (ErrUseLastResponse) — not accepted
//     as success unless the policy allows it.
//   - HTTP 4xx/5xx: not treated as ready; polling continues (service may still
//     be starting). They are reported as diagnostics on timeout.
//   - Request timeout: retried within the overall readiness deadline.
//   - Context cancellation: terminates promptly and returns an error so the
//     caller can initiate cleanup.
//   - Successful response: marks ready only when the status matches expectedStatus.
func waitForReadiness(ctx context.Context, urls []string, hc *manifest.HealthCheck) (bool, error) {
	timeout := hc.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	expectedStatus := hc.ExpectedStatus
	if expectedStatus <= 0 {
		expectedStatus = 200
	}

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	interval := 2 * time.Second

	client := &http.Client{
		Timeout: 5 * time.Second,
		// Follow redirects so that 3xx responses don't block readiness.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	lastStatus := 0
	lastURL := ""

	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, &ReadinessProbeError{URL: lastURL, Status: lastStatus, Expected: expectedStatus}
		default:
		}

		if time.Now().After(deadline) {
			return false, &ReadinessProbeError{URL: lastURL, Status: lastStatus, Expected: expectedStatus}
		}

		for _, url := range urls {
			resp, err := client.Get(url)
			if err != nil {
				lastURL = url
				continue
			}
			lastStatus = resp.StatusCode
			lastURL = url
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode == expectedStatus {
				return true, nil
			}
			// Non-success status codes are reported but we keep polling
			// because the service may still be starting up.
		}

		// Wait for the next polling interval, but remain responsive to
		// cancellation and the deadline.
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			return false, &ReadinessProbeError{URL: lastURL, Status: lastStatus, Expected: expectedStatus}
		case <-deadlineCh(deadline):
			return false, &ReadinessProbeError{URL: lastURL, Status: lastStatus, Expected: expectedStatus}
		case <-timer.C:
		}
	}
}

// deadlineCh returns a channel that fires when the deadline passes.
func deadlineCh(deadline time.Time) <-chan time.Time {
	return time.After(time.Until(deadline))
}

// portIsAvailable checks whether a TCP port is free on localhost.
// Used only as an early warning before `docker compose up`.
func portIsAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// Stop stops a running lab instance.
func (m *Manager) Stop(labID string) error {
	templatePath, err := m.catalog.TemplatePath(labID)
	if err != nil {
		return err
	}

	inst := m.store.Get(labID)
	if inst == nil {
		return &NotFoundError{LabID: labID}
	}

	if inst.Status != "running" {
		return &NotRunningError{LabID: labID}
	}

	if err := runDockerCompose(templatePath, "stop"); err != nil {
		return fmt.Errorf("failed to stop lab: %w", err)
	}

	return m.store.UpdateStatus(labID, "stopped")
}

// Reset stops and recreates a lab instance (preserves volumes).
func (m *Manager) Reset(labID string) error {
	templatePath, err := m.catalog.TemplatePath(labID)
	if err != nil {
		return err
	}

	inst := m.store.Get(labID)
	if inst == nil {
		return &NotFoundError{LabID: labID}
	}

	if err := runDockerCompose(templatePath, "down"); err != nil {
		return fmt.Errorf("failed to stop lab: %w", err)
	}

	if err := runDockerCompose(templatePath, "up", "-d"); err != nil {
		return fmt.Errorf("failed to start lab: %w", err)
	}

	return m.store.UpdateStatus(labID, "running")
}

// Remove removes a lab instance. If removeVolumes is true, volumes are also removed.
func (m *Manager) Remove(labID string, removeVolumes bool) error {
	templatePath, err := m.catalog.TemplatePath(labID)
	if err != nil {
		return err
	}

	inst := m.store.Get(labID)
	if inst == nil {
		return &NotFoundError{LabID: labID}
	}

	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}

	if err := runDockerComposeWithArgs(templatePath, args...); err != nil {
		return fmt.Errorf("failed to remove lab: %w", err)
	}

	return m.store.Remove(labID)
}

// Status returns the current status of all tracked instances.
func (m *Manager) Status() []*state.InstanceState {
	return m.store.List()
}

// Running returns all running instances.
func (m *Manager) Running() []*state.InstanceState {
	return m.store.Running()
}

// RunResult contains the result of a successful run.
type RunResult struct {
	LabID     string    `json:"labID"`
	Version   string    `json:"version"`
	Status    string    `json:"status"`
	Ports     []int     `json:"ports"`
	URLs      []string  `json:"urls"`
	StartedAt time.Time `json:"startedAt"`
}

// PolicyError represents policy validation failures.
type PolicyError struct {
	Errors []string
}

func (e *PolicyError) Error() string {
	return fmt.Sprintf("policy validation failed: %s", strings.Join(e.Errors, "; "))
}

// AlreadyRunningError is returned when trying to run an already running lab.
type AlreadyRunningError struct {
	LabID string
}

func (e *AlreadyRunningError) Error() string {
	return fmt.Sprintf("lab %s is already running", e.LabID)
}

// NotFoundError is returned when a lab instance is not found.
type NotFoundError struct {
	LabID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("lab instance not found: %s", e.LabID)
}

// NotRunningError is returned when trying to stop a non-running lab.
type NotRunningError struct {
	LabID string
}

func (e *NotRunningError) Error() string {
	return fmt.Sprintf("lab %s is not running", e.LabID)
}

// BindFailureError is returned when Docker reports a port bind failure.
type BindFailureError struct {
	Err error
}

func (e *BindFailureError) Error() string {
	return fmt.Sprintf("port bind failure: %v", e.Err)
}

// ReadinessError is returned when services fail to become ready.
type ReadinessError struct {
	Err  error
	URLs []string
}

func (e *ReadinessError) Error() string {
	return fmt.Sprintf("readiness check failed: %v", e.Err)
}

// ReadinessProbeError reports the last observed HTTP status during readiness polling.
type ReadinessProbeError struct {
	URL      string
	Status   int
	Expected int
}

func (e *ReadinessProbeError) Error() string {
	return fmt.Sprintf("readiness probe to %s returned HTTP %d (expected %d)", e.URL, e.Status, e.Expected)
}

// PortConflictError is returned when a declared host port is already in use.
type PortConflictError struct {
	Port        int
	Alternative int
	Suggestion  string
}

func (e *PortConflictError) Error() string {
	return fmt.Sprintf("port conflict: %s (port %d; suggested alternative: %d)", e.Suggestion, e.Port, e.Alternative)
}

func buildURLs(endpoints []manifest.Endpoint, ports []int) []string {
	var urls []string
	portMap := make(map[int]string)
	for _, ep := range endpoints {
		if ep.ContainerPort > 0 {
			portMap[ep.ContainerPort] = ep.HostBinding
		}
	}

	for _, port := range ports {
		hostIP, ok := portMap[port]
		if !ok {
			hostIP = "127.0.0.1"
		}
		urls = append(urls, fmt.Sprintf("http://%s:%d", hostIP, port))
	}
	return urls
}

func composeProjectName(labID string) string {
	return "labkit-" + strings.ReplaceAll(labID, "/", "-")
}

func runDockerCompose(templatePath string, args ...string) error {
	cmd := exec.Command("docker", append([]string{"compose", "-f", "compose.yaml"}, args...)...)
	cmd.Dir = templatePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runDockerComposeWithArgs(templatePath string, args ...string) error {
	cmd := exec.Command("docker", append([]string{"compose", "-f", "compose.yaml"}, args...)...)
	cmd.Dir = templatePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runDockerComposeCapture runs docker compose and captures stderr for diagnostics
// when the command fails. The error message includes the compose stderr output.
func runDockerComposeCapture(templatePath string, args ...string) error {
	cmd := exec.Command("docker", append([]string{"compose", "-f", "compose.yaml"}, args...)...)
	cmd.Dir = templatePath
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// loadComposeProject loads and resolves a Compose file using compose-go.
func loadComposeProject(composePath string) (*types.Project, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, fmt.Errorf("reading compose file: %w", err)
	}

	doc, err := loader.Load(types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{
			{Filename: composePath, Content: data},
		},
		WorkingDir: filepath.Dir(composePath),
	}, func(o *loader.Options) {
		projectName := filepath.Base(filepath.Dir(composePath))
		// Docker Compose project names must be lowercase alphanumeric,
		// hyphens, and underscores, starting with a letter or number.
		// Replace invalid characters with hyphens.
		projectName = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, projectName)
		// Ensure it starts with a letter or number
		if len(projectName) > 0 && projectName[0] < '0' {
			projectName = "lab" + projectName
		}
		o.SetProjectName(projectName, true)
	})
	if err != nil {
		return nil, fmt.Errorf("parsing compose file: %w", err)
	}

	return doc, nil
}

// allocatedHostPorts extracts the set of published host ports from a resolved
// Compose project.
func allocatedHostPorts(project *types.Project) []int {
	var ports []int
	seen := make(map[int]bool)

	for _, service := range project.Services {
		for _, port := range service.Ports {
			if port.Published != "" {
				if p, err := strconvAtoi(port.Published); err == nil && p > 0 {
					if !seen[p] {
						seen[p] = true
						ports = append(ports, p)
					}
				}
			}
		}
	}

	return ports
}

// strconvAtoi is a local wrapper to avoid importing strconv in multiple places.
func strconvAtoi(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid port: %s", s)
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
