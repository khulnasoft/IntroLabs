package compose

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderLoad(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	composePath := filepath.Join(projectRoot, "templates", "web", "intro-lab", "compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Skipf("compose file not found at %s", composePath)
	}

	loader := NewLoader()
	project, err := loader.Load(composePath)
	if err != nil {
		t.Fatalf("failed to load compose file: %v", err)
	}

	if len(project.Services) == 0 {
		t.Error("expected at least one service")
	}

	app := project.Services[0]
	if app.Name != "app" {
		t.Errorf("expected service name 'app', got %s", app.Name)
	}

	if len(app.PortBindings) == 0 {
		t.Error("expected at least one port binding")
	}

	// Check for loopback violation
	violations := project.LoopbackViolations()
	if len(violations) > 0 {
		t.Errorf("expected no loopback violations, got %d: %v", len(violations), violations)
	}
}

func TestPublishedPorts(t *testing.T) {
	project := &Project{
		Services: []ServiceInfo{
			{
				Name: "app",
				PortBindings: []PortBinding{
					{Service: "app", HostIP: "127.0.0.1", HostPort: 8081, ContainerPort: 8080},
				},
			},
		},
	}

	ports := project.PublishedPorts()
	if len(ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(ports))
	}
	if ports[0].HostPort != 8081 {
		t.Errorf("expected host port 8081, got %d", ports[0].HostPort)
	}
}

func TestLoopbackViolations(t *testing.T) {
	project := &Project{
		Services: []ServiceInfo{
			{
				Name: "app",
				PortBindings: []PortBinding{
					{Service: "app", HostIP: "127.0.0.1", HostPort: 8081, ContainerPort: 8080},
					{Service: "app", HostIP: "0.0.0.0", HostPort: 9090, ContainerPort: 9090},
				},
			},
		},
	}

	violations := project.LoopbackViolations()
	if len(violations) != 1 {
		t.Errorf("expected 1 loopback violation, got %d", len(violations))
	}
	if violations[0].HostPort != 9090 {
		t.Errorf("expected violation on port 9090, got %d", violations[0].HostPort)
	}
}

func TestFindAvailablePort(t *testing.T) {
	// Port 1 is very likely free on any system
	port, err := FindAvailablePort(1)
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}
	if port < 1 {
		t.Errorf("expected positive port, got %d", port)
	}
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
