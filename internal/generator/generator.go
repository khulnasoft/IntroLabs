package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
	"github.com/labkit-project/vulnerable-lab-platform/internal/policy"
)

// Generator orchestrates the CVE-to-lab template generation flow.
type Generator struct {
	ingestor   *Ingestor
	normalizer *Normalizer
	scaffolder *ScaffoldGenerator
	verifier   *Verifier
	packer     *Packager
	catalog    *Catalog
	rootDir    string
}

// NewGenerator creates a Generator wired with all sub-components.
func NewGenerator(rootDir string) *Generator {
	schemaPath := filepath.Join(rootDir, "schemas", "lab.schema.json")
	ml, err := manifest.NewLoader(schemaPath)
	if err != nil {
		// Should not happen in normal operation; schema is required
		ml = nil
	}
	pv := policy.NewValidator()

	return &Generator{
		ingestor:   NewIngestor(),
		normalizer: NewNormalizer(),
		scaffolder: NewScaffoldGenerator(rootDir),
		verifier:   NewVerifier(ml, pv),
		packer:     NewPackager(),
		catalog:    NewCatalog(ResolveCatalogPath(rootDir)),
		rootDir:    rootDir,
	}
}

// GenerateResult holds the outcome of a generation run.
type GenerateResult struct {
	TemplatePath string
	State        TemplateState
	Normalized   *NormalizedCVE
	GateResults  GateResultSlice
}

// GenerateCVE runs the full generate flow: ingest, normalize, scaffold.
func (g *Generator) GenerateCVE(cveID, outputPath, advisoryURL, appFamily string) (*GenerateResult, error) {
	if !IsValidCVE(cveID) {
		return nil, fmt.Errorf("invalid CVE identifier: %s", cveID)
	}

	// Step 1: Ingest CVE metadata
	cve, err := g.ingestor.Fetch(cveID)
	if err != nil {
		return nil, fmt.Errorf("ingesting CVE %s: %w", cveID, err)
	}

	// Step 2: Normalize
	nc := g.normalizer.Normalize(cve, advisoryURL)
	if !nc.Actionable {
		return nil, fmt.Errorf("CVE %s is not actionable: %s", cveID, nc.NonActionableReason)
	}

	// Step 3: Scaffold template
	if err := g.scaffolder.Generate(outputPath, nc); err != nil {
		return nil, fmt.Errorf("scaffolding template: %w", err)
	}

	return &GenerateResult{
		TemplatePath: outputPath,
		State:        StateDraft,
		Normalized:   nc,
	}, nil
}

// ValidateGate runs the verification gate on a template and returns results.
func (g *Generator) ValidateGate(templatePath string) GateResultSlice {
	return g.verifier.Verify(templatePath)
}

// CanTransitionTo checks whether a state transition is allowed.
func (g *Generator) CanTransitionTo(from, to TemplateState) bool {
	return CanTransition(from, to)
}

// PackageTemplate packages a template directory.
func (g *Generator) PackageTemplate(templatePath, version string) (*PackageResult, error) {
	return g.packer.Package(templatePath, version)
}

// WriteCatalogEntry adds a catalog entry for a published template.
func (g *Generator) WriteCatalogEntry(catalogPath string, entry CatalogEntry) error {
	return g.catalog.WriteCatalogEntry(entry)
}

// EnsureDirExists checks that a directory exists.
func EnsureDirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

var _ = manifest.SchemaVersion // ensure manifest import is used
