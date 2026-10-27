package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// SchemaVersion is the lab manifest schema version.
const SchemaVersion = "labkit/v1"

// Kind is the manifest resource kind.
const Kind = "Environment"

// Manifest represents a parsed lab.yaml manifest.
type Manifest struct {
	SchemaVersion string            `yaml:"schemaVersion"`
	Kind          string            `yaml:"kind"`
	Metadata      Metadata          `yaml:"metadata"`
	Learning      Learning          `yaml:"learning"`
	Runtime       Runtime           `yaml:"runtime"`
	Endpoints     []Endpoint        `yaml:"endpoints"`
	Checks        Checks            `yaml:"checks"`
	CoverageState CoverageState     `yaml:"coverage_state,omitempty"`
	Vulnerability *Vulnerability    `yaml:"vulnerability,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Secrets       []Secret          `yaml:"secrets,omitempty"`
	References    []Reference       `yaml:"references,omitempty"`
}

// Metadata describes the lab.
type Metadata struct {
	ID               string   `yaml:"id"`
	Name             string   `yaml:"name"`
	Version          string   `yaml:"version"`
	Description      string   `yaml:"description"`
	Category         string   `yaml:"category"`
	Difficulty       string   `yaml:"difficulty"`
	License          string   `yaml:"license"`
	Author           string   `yaml:"author,omitempty"`
	Tags             []string `yaml:"tags,omitempty"`
	CVE              string   `yaml:"cve,omitempty"`
	Vendor           string   `yaml:"vendor,omitempty"`
	Product          string   `yaml:"product,omitempty"`
	AffectedVersions []string `yaml:"affected_versions,omitempty"`
	FixedVersions    []string `yaml:"fixed_versions,omitempty"`
}

// CoverageState tracks catalog coverage for a lab environment.
type CoverageState string

const (
	CoverageDiscovered   CoverageState = "discovered"
	CoverageTriaged      CoverageState = "triaged"
	CoverageReproducible CoverageState = "reproducible"
	CoveragePublished    CoverageState = "published"
	CoverageRetired      CoverageState = "retired"
)

// Vulnerability describes vulnerability-specific information for CVE-driven labs.
type Vulnerability struct {
	Severity       string          `yaml:"severity,omitempty"`
	SeveritySource string          `yaml:"severity_source,omitempty"`
	References     []VulnReference `yaml:"references,omitempty"`
	Provenance     *Provenance     `yaml:"provenance,omitempty"`
	Verification   *Verification   `yaml:"verification,omitempty"`
	Safety         *Safety         `yaml:"safety,omitempty"`
}

// VulnReference describes an advisory or database reference.
type VulnReference struct {
	Title       string `yaml:"title"`
	URL         string `yaml:"url"`
	Description string `yaml:"description,omitempty"`
}

// Provenance describes provenance information for the lab environment.
type Provenance struct {
	ImageDigest      string `yaml:"image_digest,omitempty"`
	SourceRevision   string `yaml:"source_revision,omitempty"`
	VerificationDate string `yaml:"verification_date,omitempty"`
	VerifiedBy       string `yaml:"verified_by,omitempty"`
}

// Verification describes checks that confirm the documented lab state.
type Verification struct {
	AffectedVersionConfirmed    bool   `yaml:"affected_version_confirmed,omitempty"`
	FixedVersionConfirmed       bool   `yaml:"fixed_version_confirmed,omitempty"`
	ReproductionStepsDocumented bool   `yaml:"reproduction_steps_documented,omitempty"`
	Notes                       string `yaml:"notes,omitempty"`
}

// Safety describes network restrictions, isolation, and resource limits.
type Safety struct {
	NetworkIsolation bool   `yaml:"network_isolation,omitempty"`
	NoPublicExposure bool   `yaml:"no_public_exposure,omitempty"`
	ResourceLimits   bool   `yaml:"resource_limits,omitempty"`
	CleanupBehavior  string `yaml:"cleanup_behavior,omitempty"`
}

// Learning describes educational content.
type Learning struct {
	Objectives    []string `yaml:"objectives"`
	Prerequisites []string `yaml:"prerequisites,omitempty"`
	EstimatedTime string   `yaml:"estimatedTime,omitempty"`
}

// Runtime describes container orchestration requirements.
type Runtime struct {
	Engine         string    `yaml:"engine"`
	ComposeFile    string    `yaml:"composeFile"`
	Architectures  []string  `yaml:"architectures"`
	Resources      Resources `yaml:"resources"`
	Network        Network   `yaml:"network"`
	DockerVersion  string    `yaml:"dockerVersion,omitempty"`
	ComposeVersion string    `yaml:"composeVersion,omitempty"`
}

// Resources describes minimum resource requirements.
type Resources struct {
	MemoryMiB int     `yaml:"memoryMiB"`
	CPU       float64 `yaml:"cpu,omitempty"`
	DiskMiB   int     `yaml:"diskMiB,omitempty"`
}

// Network describes network exposure policy.
type Network struct {
	Exposure string `yaml:"exposure"`
	Outbound string `yaml:"outbound,omitempty"`
}

// Endpoint describes a declared service endpoint.
type Endpoint struct {
	ID            string `yaml:"id"`
	Service       string `yaml:"service"`
	Protocol      string `yaml:"protocol"`
	ContainerPort int    `yaml:"containerPort"`
	HostBinding   string `yaml:"hostBinding"`
	Path          string `yaml:"path,omitempty"`
	Description   string `yaml:"description,omitempty"`
}

// Checks describes validation and health checks.
type Checks struct {
	SmokeTest      string       `yaml:"smokeTest"`
	ReadinessCheck string       `yaml:"readinessCheck,omitempty"`
	HealthCheck    *HealthCheck `yaml:"healthCheck,omitempty"`
}

// HealthCheck describes an HTTP health check.
type HealthCheck struct {
	Endpoint       string `yaml:"endpoint"`
	ExpectedStatus int    `yaml:"expectedStatus,omitempty"`
	Timeout        int    `yaml:"timeout,omitempty"`
}

// Secret describes a required secret (name only, no value).
type Secret struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Reference describes an external reference.
type Reference struct {
	Title       string `yaml:"title"`
	URL         string `yaml:"url"`
	Description string `yaml:"description,omitempty"`
}

// Loader loads and validates manifests.
type Loader struct {
	schemaLoader gojsonschema.JSONLoader
}

// NewLoader creates a manifest loader using the given schema file.
func NewLoader(schemaPath string) (*Loader, error) {
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	return &Loader{schemaLoader: schemaLoader}, nil
}

// Load parses a lab.yaml file and validates it against the schema.
func (l *Loader) Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest YAML: %w", err)
	}

	if err := l.validate(data); err != nil {
		return nil, fmt.Errorf("schema validation: %w", err)
	}

	return &m, nil
}

// validate checks the raw YAML bytes against the JSON schema.
func (l *Loader) validate(data []byte) error {
	// Convert YAML to JSON for JSON Schema validation
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshaling to JSON: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	result, err := gojsonschema.Validate(l.schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("running schema validation: %w", err)
	}

	if !result.Valid() {
		errs := make([]string, 0, len(result.Errors()))
		for _, desc := range result.Errors() {
			errs = append(errs, desc.String())
		}
		return &ValidationError{Errors: errs}
	}
	return nil
}

// ValidationError contains schema validation errors.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("manifest validation failed (%d errors)", len(e.Errors))
}

// LoadFromTemplate loads a manifest from a template directory.
func (l *Loader) LoadFromTemplate(templatePath string) (*Manifest, error) {
	return l.Load(filepath.Join(templatePath, "lab.yaml"))
}

// MarshalJSON returns JSON representation of the manifest.
func (m *Manifest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SchemaVersion string            `json:"schemaVersion"`
		Kind          string            `json:"kind"`
		Metadata      Metadata          `json:"metadata"`
		Learning      Learning          `json:"learning"`
		Runtime       Runtime           `json:"runtime"`
		Endpoints     []Endpoint        `json:"endpoints"`
		Checks        Checks            `json:"checks"`
		CoverageState CoverageState     `json:"coverage_state,omitempty"`
		Vulnerability *Vulnerability    `json:"vulnerability,omitempty"`
		Environment   map[string]string `json:"environment,omitempty"`
		Secrets       []Secret          `json:"secrets,omitempty"`
		References    []Reference       `json:"references,omitempty"`
	}{
		SchemaVersion: m.SchemaVersion,
		Kind:          m.Kind,
		Metadata:      m.Metadata,
		Learning:      m.Learning,
		Runtime:       m.Runtime,
		Endpoints:     m.Endpoints,
		Checks:        m.Checks,
		CoverageState: m.CoverageState,
		Vulnerability: m.Vulnerability,
		Environment:   m.Environment,
		Secrets:       m.Secrets,
		References:    m.References,
	})
}
