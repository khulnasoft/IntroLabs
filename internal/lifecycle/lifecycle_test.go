package lifecycle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
)

func TestPortConflictError(t *testing.T) {
	err := &PortConflictError{
		Port:        8081,
		Alternative: 8082,
		Suggestion:  "port 8081 is already in use",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	if err.Port != 8081 {
		t.Errorf("expected port 8081, got %d", err.Port)
	}
	if err.Alternative != 8082 {
		t.Errorf("expected alternative port 8082, got %d", err.Alternative)
	}
}

func TestPolicyError(t *testing.T) {
	err := &PolicyError{
		Errors: []string{"privileged mode not allowed", "host networking not allowed"},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestAlreadyRunningError(t *testing.T) {
	err := &AlreadyRunningError{LabID: "web/intro-lab"}
	if err.Error() != "lab web/intro-lab is already running" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{LabID: "web/intro-lab"}
	if err.Error() != "lab instance not found: web/intro-lab" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNotRunningError(t *testing.T) {
	err := &NotRunningError{LabID: "web/intro-lab"}
	if err.Error() != "lab web/intro-lab is not running" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestBindFailureError(t *testing.T) {
	underlying := fmt.Errorf("Bind for 0.0.0.0:8080 failed: port is already allocated")
	err := &BindFailureError{Err: underlying}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	if !strings.Contains(msg, "port bind failure") {
		t.Errorf("expected 'port bind failure' in message, got %s", msg)
	}
}

func TestReadinessError(t *testing.T) {
	err := &ReadinessError{
		Err:  fmt.Errorf("timeout"),
		URLs: []string{"http://127.0.0.1:8081/health"},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	if len(err.URLs) != 1 {
		t.Errorf("expected 1 URL, got %d", len(err.URLs))
	}
}

func TestReadinessProbeError(t *testing.T) {
	err := &ReadinessProbeError{
		URL:      "http://127.0.0.1:8081/health",
		Status:   500,
		Expected: 200,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	if !strings.Contains(msg, "HTTP 500") {
		t.Errorf("expected 'HTTP 500' in message, got %s", msg)
	}
	if !strings.Contains(msg, "expected 200") {
		t.Errorf("expected 'expected 200' in message, got %s", msg)
	}
}

func TestComposeProjectName(t *testing.T) {
	name := composeProjectName("web/intro-lab")
	if name != "labkit-web-intro-lab" {
		t.Errorf("expected labkit-web-intro-lab, got %s", name)
	}

	name = composeProjectName("api/auth-lab")
	if name != "labkit-api-auth-lab" {
		t.Errorf("expected labkit-api-auth-lab, got %s", name)
	}
}

func TestBuildURLs(t *testing.T) {
	endpoints := []manifest.Endpoint{
		{ID: "web-ui", ContainerPort: 8080, HostBinding: "127.0.0.1"},
	}
	ports := []int{8081}

	urls := buildURLs(endpoints, ports)
	if len(urls) != 1 {
		t.Errorf("expected 1 URL, got %d", len(urls))
	}
	if urls[0] != "http://127.0.0.1:8081" {
		t.Errorf("expected http://127.0.0.1:8081, got %s", urls[0])
	}
}

func TestBuildURLsUnknownPort(t *testing.T) {
	endpoints := []manifest.Endpoint{
		{ID: "web-ui", ContainerPort: 8080, HostBinding: "127.0.0.1"},
	}
	ports := []int{9999} // port not in endpoint map

	urls := buildURLs(endpoints, ports)
	if len(urls) != 1 {
		t.Errorf("expected 1 URL, got %d", len(urls))
	}
	if urls[0] != "http://127.0.0.1:9999" {
		t.Errorf("expected http://127.0.0.1:9999 (default host binding), got %s", urls[0])
	}
}

func TestBuildHealthURLs(t *testing.T) {
	// Endpoint with no port placeholder returns as-is
	urls := buildHealthURLs("http://127.0.0.1:8081/health", []int{8081})
	if len(urls) != 1 {
		t.Errorf("expected 1 URL, got %d", len(urls))
	}
	if urls[0] != "http://127.0.0.1:8081/health" {
		t.Errorf("expected unchanged endpoint, got %s", urls[0])
	}

	// Empty ports returns endpoint as-is
	urls = buildHealthURLs("http://127.0.0.1:8081/health", []int{})
	if len(urls) != 1 || urls[0] != "http://127.0.0.1:8081/health" {
		t.Errorf("expected unchanged endpoint for empty ports, got %v", urls)
	}
}

func TestPortIsAvailable(t *testing.T) {
	// Port 1 is very likely available on any system
	if !portIsAvailable(1) {
		t.Error("expected port 1 to be available")
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	alreadyRunning := &AlreadyRunningError{LabID: "test"}
	notFound := &NotFoundError{LabID: "test"}
	notRunning := &NotRunningError{LabID: "test"}
	portConflict := &PortConflictError{Port: 8080, Alternative: 8081, Suggestion: "conflict"}
	policyErr := &PolicyError{Errors: []string{"violation"}}
	bindErr := &BindFailureError{Err: fmt.Errorf("bind failed")}
	readinessErr := &ReadinessError{Err: fmt.Errorf("timeout"), URLs: []string{"http://x"}}
	probeErr := &ReadinessProbeError{URL: "http://x", Status: 500, Expected: 200}

	messages := map[string]bool{
		alreadyRunning.Error(): true,
		notFound.Error():       true,
		notRunning.Error():     true,
		portConflict.Error():   true,
		policyErr.Error():      true,
		bindErr.Error():        true,
		readinessErr.Error():   true,
		probeErr.Error():       true,
	}

	if len(messages) != 8 {
		t.Errorf("expected 8 distinct error messages, got %d", len(messages))
	}
}
