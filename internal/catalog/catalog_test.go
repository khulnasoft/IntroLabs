package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolverList(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	resolver, err := NewResolver(projectRoot)
	if err != nil {
		t.Fatalf("failed to create resolver: %v", err)
	}

	labs := resolver.List()
	if len(labs) == 0 {
		t.Error("expected at least one lab in catalog")
	}

	found := false
	for _, lab := range labs {
		if lab.ID == "web/intro-lab" {
			found = true
			if lab.Name != "Introduction to Web Security" {
				t.Errorf("expected name Introduction to Web Security, got %s", lab.Name)
			}
			if lab.Category != "web" {
				t.Errorf("expected category web, got %s", lab.Category)
			}
		}
	}
	if !found {
		t.Error("expected web/intro-lab in catalog")
	}
}

func TestResolverFind(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	resolver, err := NewResolver(projectRoot)
	if err != nil {
		t.Fatalf("failed to create resolver: %v", err)
	}

	lab := resolver.Find("web/intro-lab")
	if lab == nil {
		t.Fatal("expected to find web/intro-lab")
	}
	if lab.Path != "templates/web/intro-lab" {
		t.Errorf("expected path templates/web/intro-lab, got %s", lab.Path)
	}

	if resolver.Find("nonexistent/lab") != nil {
		t.Error("expected nil for nonexistent lab")
	}
}

func TestResolverSearch(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	resolver, err := NewResolver(projectRoot)
	if err != nil {
		t.Fatalf("failed to create resolver: %v", err)
	}

	results := resolver.Search("web")
	if len(results) == 0 {
		t.Error("expected search for 'web' to return results")
	}

	results = resolver.Search("nonexistent-query-xyz")
	if len(results) != 0 {
		t.Error("expected no results for nonexistent query")
	}
}

func TestResolverTemplatePath(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	resolver, err := NewResolver(projectRoot)
	if err != nil {
		t.Fatalf("failed to create resolver: %v", err)
	}

	path, err := resolver.TemplatePath("web/intro-lab")
	if err != nil {
		t.Fatalf("failed to get template path: %v", err)
	}

	expected := filepath.Join(projectRoot, "templates", "web", "intro-lab")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("template path does not exist: %s", path)
	}

	_, err = resolver.TemplatePath("nonexistent/lab")
	if err == nil {
		t.Error("expected error for nonexistent lab")
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
