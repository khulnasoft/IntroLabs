package compose

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

// PortBinding represents a host port binding for a service.
type PortBinding struct {
	Service       string
	HostIP        string
	HostPort      int
	ContainerPort int
	Protocol      string
}

// ServiceInfo summarizes a compose service.
type ServiceInfo struct {
	Name         string
	Image        string
	PortBindings []PortBinding
	Networks     []string
	Privileged   bool
}

// Project represents a loaded compose project.
type Project struct {
	Services []ServiceInfo
	Networks map[string]NetworkInfo
	Volumes  map[string]VolumeInfo
}

// NetworkInfo describes a compose network.
type NetworkInfo struct {
	Name     string
	Driver   string
	Internal bool
}

// VolumeInfo describes a compose volume.
type VolumeInfo struct {
	Name   string
	Driver string
}

// Loader loads and inspects compose files.
type Loader struct{}

// NewLoader creates a compose loader.
func NewLoader() *Loader {
	return &Loader{}
}

// Load parses a compose file and returns project info.
func (l *Loader) Load(composePath string) (*Project, error) {
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
		o.SetProjectName(filepath.Base(filepath.Dir(composePath)), true)
	})
	if err != nil {
		return nil, fmt.Errorf("parsing compose file: %w", err)
	}

	project := &Project{
		Services: make([]ServiceInfo, 0, len(doc.Services)),
		Networks: make(map[string]NetworkInfo),
		Volumes:  make(map[string]VolumeInfo),
	}

	for _, service := range doc.Services {
		si := ServiceInfo{
			Name:       service.Name,
			Image:      service.Image,
			Privileged: service.Privileged,
		}

		// Extract network names from service.Networks map keys
		for netName := range service.Networks {
			si.Networks = append(si.Networks, netName)
		}

		// Extract port bindings
		for _, port := range service.Ports {
			pb := PortBinding{
				Service:       service.Name,
				ContainerPort: int(port.Target),
				Protocol:      port.Protocol,
			}
			if port.Published != "" {
				if parsed, err := strconv.Atoi(port.Published); err == nil {
					pb.HostPort = parsed
				}
			}
			if port.HostIP != "" {
				pb.HostIP = port.HostIP
			} else {
				pb.HostIP = "0.0.0.0"
			}
			si.PortBindings = append(si.PortBindings, pb)
		}

		project.Services = append(project.Services, si)
	}

	for name, net := range doc.Networks {
		project.Networks[name] = NetworkInfo{
			Name:     name,
			Driver:   net.Driver,
			Internal: net.Internal,
		}
	}

	for name, vol := range doc.Volumes {
		project.Volumes[name] = VolumeInfo{
			Name:   name,
			Driver: vol.Driver,
		}
	}

	return project, nil
}

// PublishedPorts returns all published port bindings across services.
func (p *Project) PublishedPorts() []PortBinding {
	var bindings []PortBinding
	for _, svc := range p.Services {
		bindings = append(bindings, svc.PortBindings...)
	}
	return bindings
}

// LoopbackViolations returns port bindings not bound to loopback.
func (p *Project) LoopbackViolations() []PortBinding {
	var violations []PortBinding
	for _, pb := range p.PublishedPorts() {
		if pb.HostIP != "127.0.0.1" && pb.HostIP != "::1" && pb.HostIP != "localhost" {
			violations = append(violations, pb)
		}
	}
	return violations
}

// AllocatedHostPorts returns the set of host ports in use.
func (p *Project) AllocatedHostPorts() []int {
	var ports []int
	seen := make(map[int]bool)
	for _, pb := range p.PublishedPorts() {
		if pb.HostPort > 0 && !seen[pb.HostPort] {
			seen[pb.HostPort] = true
			ports = append(ports, pb.HostPort)
		}
	}
	return ports
}

// FindAvailablePort finds a free TCP port starting from startPort.
func FindAvailablePort(startPort int) (int, error) {
	for port := startPort; port < 65535; port++ {
		if portIsFree(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found starting from %d", startPort)
}

func portIsFree(port int) bool {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// EnsureComposeFile checks that a compose file exists at the given path.
func EnsureComposeFile(templatePath string) (string, error) {
	composePath := filepath.Join(templatePath, "compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return "", fmt.Errorf("compose.yaml not found in template: %s", templatePath)
	}
	return composePath, nil
}
