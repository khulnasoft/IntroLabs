package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidManifest(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	manifestPath := filepath.Join(projectRoot, "templates", "web", "intro-lab", "lab.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skipf("manifest not found at %s", manifestPath)
	}

	loader, err := NewLoader(filepath.Join(projectRoot, "schemas", "lab.schema.json"))
	if err != nil {
		t.Fatalf("failed to create loader: %v", err)
	}

	m, err := loader.Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if m.Metadata.ID != "web/intro-lab" {
		t.Errorf("expected id web/intro-lab, got %s", m.Metadata.ID)
	}
	if m.Runtime.Engine != "docker-compose" {
		t.Errorf("expected engine docker-compose, got %s", m.Runtime.Engine)
	}
	if len(m.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(m.Endpoints))
	}
	if m.Endpoints[0].HostBinding != "127.0.0.1" {
		t.Errorf("expected hostBinding 127.0.0.1, got %s", m.Endpoints[0].HostBinding)
	}
}

func TestLoadInvalidManifest(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	loader, err := NewLoader(filepath.Join(projectRoot, "schemas", "lab.schema.json"))
	if err != nil {
		t.Fatalf("failed to create loader: %v", err)
	}

	// Create a temp invalid manifest (missing required fields)
	tmpDir := t.TempDir()
	invalidYAML := []byte("schemaVersion: labkit/v1\nkind: Environment\n")
	invalidPath := filepath.Join(tmpDir, "lab.yaml")
	if err := os.WriteFile(invalidPath, invalidYAML, 0644); err != nil {
		t.Fatalf("failed to write temp manifest: %v", err)
	}

	_, err = loader.Load(invalidPath)
	if err == nil {
		t.Error("expected validation to fail for invalid manifest")
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
