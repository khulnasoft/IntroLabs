package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CatalogVersion is the catalog schema version.
const CatalogVersion = "catalog/v1"

// Catalog represents the lab catalog index.
type Catalog struct {
	Version     string `yaml:"version"`
	LastUpdated string `yaml:"lastUpdated,omitempty"`
	Labs        []Lab  `yaml:"labs"`
}

// Lab describes a lab entry in the catalog.
type Lab struct {
	ID               string   `yaml:"id"`
	Name             string   `yaml:"name"`
	Version          string   `yaml:"version"`
	Category         string   `yaml:"category"`
	Difficulty       string   `yaml:"difficulty"`
	Path             string   `yaml:"path"`
	Description      string   `yaml:"description,omitempty"`
	Tags             []string `yaml:"tags,omitempty"`
	Architectures    []string `yaml:"architectures,omitempty"`
	License          string   `yaml:"license,omitempty"`
	Author           string   `yaml:"author,omitempty"`
	Deprecated       bool     `yaml:"deprecated,omitempty"`
	DeprecatedReason string   `yaml:"deprecatedReason,omitempty"`
}

// Resolver loads and queries the catalog.
type Resolver struct {
	catalog Catalog
	rootDir string
}

// NewResolver loads the catalog from index.yaml in the given root directory.
func NewResolver(rootDir string) (*Resolver, error) {
	indexPath := filepath.Join(rootDir, "catalog", "index.yaml")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("reading catalog index: %w", err)
	}

	var cat Catalog
	if err := yaml.Unmarshal(data, &cat); err != nil {
		return nil, fmt.Errorf("parsing catalog YAML: %w", err)
	}

	if cat.Version != CatalogVersion {
		return nil, fmt.Errorf("unsupported catalog version: %s", cat.Version)
	}

	return &Resolver{catalog: cat, rootDir: rootDir}, nil
}

// List returns all labs in the catalog.
func (r *Resolver) List() []Lab {
	return r.catalog.Labs
}

// Find returns the lab with the given ID, or nil if not found.
func (r *Resolver) Find(id string) *Lab {
	for i := range r.catalog.Labs {
		if r.catalog.Labs[i].ID == id {
			return &r.catalog.Labs[i]
		}
	}
	return nil
}

// Search returns labs matching the query (case-insensitive substring match on id, name, category, tags).
func (r *Resolver) Search(query string) []Lab {
	q := strings.ToLower(query)
	var results []Lab
	for _, lab := range r.catalog.Labs {
		if matches(lab, q) {
			results = append(results, lab)
		}
	}
	return results
}

// TemplatePath returns the absolute path to a lab's template directory.
func (r *Resolver) TemplatePath(labID string) (string, error) {
	lab := r.Find(labID)
	if lab == nil {
		return "", fmt.Errorf("lab not found: %s", labID)
	}
	return filepath.Join(r.rootDir, lab.Path), nil
}

func matches(lab Lab, q string) bool {
	if strings.Contains(strings.ToLower(lab.ID), q) {
		return true
	}
	if strings.Contains(strings.ToLower(lab.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(lab.Category), q) {
		return true
	}
	if strings.Contains(strings.ToLower(lab.Difficulty), q) {
		return true
	}
	for _, tag := range lab.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}
