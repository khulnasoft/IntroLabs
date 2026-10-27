package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// CatalogEntry is a catalog index entry for a published template.
type CatalogEntry struct {
	CVEID              string    `yaml:"cve_id"`
	TemplateID         string    `yaml:"template_id"`
	Version            string    `yaml:"version"`
	Technology         string    `yaml:"technology"`
	AffectedComponent  string    `yaml:"affected_component"`
	Category           string    `yaml:"category"`
	Difficulty         string    `yaml:"difficulty"`
	VerificationStatus string    `yaml:"verification_status"`
	PublishedAt        time.Time `yaml:"published_at"`
	Checksum           string    `yaml:"checksum"`
	Archive            string    `yaml:"archive"`
}

// Catalog manages the catalog/index.yaml file for generated labs.
type Catalog struct {
	path string
}

// NewCatalog creates a Catalog manager for the given catalog index path.
func NewCatalog(catalogPath string) *Catalog {
	return &Catalog{path: catalogPath}
}

// WriteCatalogEntry appends or replaces an entry in the catalog index.
func (c *Catalog) WriteCatalogEntry(entry CatalogEntry) error {
	data, err := os.ReadFile(c.path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading catalog: %w", err)
	}

	var existing []CatalogEntry
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parsing catalog: %w", err)
		}
	}

	// Replace existing entry with same template_id
	found := false
	for i, e := range existing {
		if e.TemplateID == entry.TemplateID {
			existing[i] = entry
			found = true
			break
		}
	}
	if !found {
		existing = append(existing, entry)
	}

	out, err := yaml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("marshaling catalog: %w", err)
	}

	if err := os.WriteFile(c.path, out, 0o644); err != nil {
		return fmt.Errorf("writing catalog: %w", err)
	}
	return nil
}

// ResolveCatalogPath returns the path to the catalog index file.
func ResolveCatalogPath(rootDir string) string {
	return filepath.Join(rootDir, "catalog", "index.yaml")
}

// NormalizeTemplateID converts a filesystem path to a catalog template ID.
func NormalizeTemplateID(path string) string {
	base := filepath.Base(path)
	if !filepath.IsAbs(base) && !containsSlash(base) {
		return "web/" + base
	}
	return base
}

func containsSlash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return true
		}
	}
	return false
}

// _ = time.Second to ensure time package is used if not already
var _ = time.Second
