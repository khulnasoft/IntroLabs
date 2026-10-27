package packaging

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackagerPackage(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "my-lab")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create a complete minimal template
	files := map[string]string{
		"lab.yaml": `schemaVersion: "labkit/v1"
kind: Environment
metadata:
  id: test/my-lab
  name: "Test Lab"
  version: "1.0.0"
  description: "A test lab"
  category: web
  difficulty: beginner
  license: Apache-2.0
learning:
  objectives:
    - Test objective
runtime:
  engine: docker-compose
  composeFile: compose.yaml
  architectures:
    - amd64
  resources:
    memoryMiB: 256
  network:
    exposure: loopback
endpoints:
  - id: web-ui
    service: app
    protocol: http
    containerPort: 8080
    hostBinding: 127.0.0.1
checks:
  smokeTest: tests/smoke.sh
`,
		"compose.yaml": `services:
  app:
    image: alpine:3.18
    ports:
      - "127.0.0.1:8080:8080"
`,
		"README.md":              `# Test Lab`,
		"lessons/objectives.md":  `# Objectives`,
		"lessons/concepts.md":    `# Concepts`,
		"lessons/remediation.md": `# Remediation`,
		"tests/smoke.sh":         `#!/bin/sh\nexit 0\n`,
		".env.example":           `# Example`,
		"LICENSE":                `Apache-2.0`,
	}

	for name, content := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(templateDir, name)), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(templateDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	packager := NewPackager()
	result, err := packager.Package(templateDir)
	if err != nil {
		t.Fatalf("failed to package: %v", err)
	}

	if result.Archive == "" {
		t.Error("expected archive path")
	}
	if result.Checksum == "" {
		t.Error("expected checksum")
	}
	if result.ChecksumFile == "" {
		t.Error("expected checksum file path")
	}
	if result.MetadataFile == "" {
		t.Error("expected metadata file path")
	}
	if result.Size == 0 {
		t.Error("expected non-zero archive size")
	}
	if result.Metadata.TemplateID != "test/my-lab" {
		t.Errorf("expected templateID test/my-lab, got %s", result.Metadata.TemplateID)
	}
	if result.Metadata.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", result.Metadata.Version)
	}
	if result.Metadata.ChecksumSHA256 != result.Checksum {
		t.Errorf("metadata checksum %s does not match archive checksum %s", result.Metadata.ChecksumSHA256, result.Checksum)
	}

	// Verify files exist
	for _, f := range []string{result.Archive, result.ChecksumFile, result.MetadataFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("file does not exist: %s", f)
		}
	}

	// Verify checksum content
	checksumData, err := os.ReadFile(result.ChecksumFile)
	if err != nil {
		t.Fatalf("failed to read checksum file: %v", err)
	}
	if string(checksumData) != result.Checksum+"\n" {
		t.Errorf("checksum file content mismatch")
	}

	// Verify metadata is valid JSON
	metadataData, err := os.ReadFile(result.MetadataFile)
	if err != nil {
		t.Fatalf("failed to read metadata file: %v", err)
	}
	var meta PackageMetadata
	if err := json.Unmarshal(metadataData, &meta); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	if meta.TemplateID != "test/my-lab" {
		t.Errorf("metadata templateID mismatch: %s", meta.TemplateID)
	}
}

func TestPackagerExcludesSecrets(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "secret-lab")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create allowed files
	os.WriteFile(filepath.Join(templateDir, "lab.yaml"), []byte("schemaVersion: labkit/v1\nmetadata:\n  id: test/secret-lab\n  name: Secret\n  version: 1.0.0\n  category: web\n  difficulty: beginner\n  license: MIT\nlearning:\n  objectives:\n    - test\nruntime:\n  engine: docker-compose\n  composeFile: compose.yaml\n  architectures:\n    - amd64\n  resources:\n    memoryMiB: 256\n  network:\n    exposure: loopback\nendpoints:\n  - id: web\n    service: app\n    protocol: http\n    containerPort: 8080\n    hostBinding: 127.0.0.1\nchecks:\n  smokeTest: tests/smoke.sh\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "compose.yaml"), []byte("services:\n  app:\n    image: alpine:3.18\n    ports:\n      - \"127.0.0.1:8080:8080\"\n"), 0644)
	os.MkdirAll(filepath.Join(templateDir, "tests"), 0755)
	os.WriteFile(filepath.Join(templateDir, "tests/smoke.sh"), []byte("#!/bin/sh\nexit 0\n"), 0644)

	// Create excluded files
	os.WriteFile(filepath.Join(templateDir, ".env"), []byte("SECRET=abc123\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "cache.tmp"), []byte("cache\n"), 0644)
	os.MkdirAll(filepath.Join(templateDir, ".git"), 0755)
	os.WriteFile(filepath.Join(templateDir, ".git", "config"), []byte("[core]\n"), 0644)

	packager := NewPackager()
	result, err := packager.Package(templateDir)
	if err != nil {
		t.Fatalf("failed to package: %v", err)
	}

	// Open archive and verify excluded files are not present
	f, err := os.Open(result.Archive)
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	foundExcluded := false
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar: %v", err)
		}

		if strings.Contains(header.Name, ".env") ||
			strings.Contains(header.Name, ".git") ||
			strings.Contains(header.Name, "cache.tmp") {
			foundExcluded = true
			t.Errorf("excluded file found in archive: %s", header.Name)
		}
	}

	if foundExcluded {
		t.Error("archive contains excluded files")
	}
}

func TestPackagerOnlyIncludesAllowedFiles(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "allowed-lab")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create allowed files
	os.WriteFile(filepath.Join(templateDir, "lab.yaml"), []byte("schemaVersion: labkit/v1\nmetadata:\n  id: test/allowed-lab\n  name: Allowed\n  version: 1.0.0\n  category: web\n  difficulty: beginner\n  license: MIT\nlearning:\n  objectives:\n    - test\nruntime:\n  engine: docker-compose\n  composeFile: compose.yaml\n  architectures:\n    - amd64\n  resources:\n    memoryMiB: 256\n  network:\n    exposure: loopback\nendpoints:\n  - id: web\n    service: app\n    protocol: http\n    containerPort: 8080\n    hostBinding: 127.0.0.1\nchecks:\n  smokeTest: tests/smoke.sh\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "compose.yaml"), []byte("services:\n  app:\n    image: alpine:3.18\n    ports:\n      - \"127.0.0.1:8080:8080\"\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "README.md"), []byte("# Allowed Lab"), 0644)
	os.MkdirAll(filepath.Join(templateDir, "lessons"), 0755)
	os.WriteFile(filepath.Join(templateDir, "lessons", "objectives.md"), []byte("# Objectives"), 0644)
	os.MkdirAll(filepath.Join(templateDir, "tests"), 0755)
	os.WriteFile(filepath.Join(templateDir, "tests", "smoke.sh"), []byte("#!/bin/sh\nexit 0\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "LICENSE"), []byte("MIT"), 0644)
	os.WriteFile(filepath.Join(templateDir, ".env.example"), []byte("# Example"), 0644)

	// Create unrelated file that should be excluded
	os.WriteFile(filepath.Join(templateDir, "random.txt"), []byte("random\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "notes.md"), []byte("notes\n"), 0644)

	packager := NewPackager()
	result, err := packager.Package(templateDir)
	if err != nil {
		t.Fatalf("failed to package: %v", err)
	}

	// Open archive and verify only allowed files are present
	f, err := os.Open(result.Archive)
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	allowed := map[string]bool{
		"lab.yaml":              true,
		"compose.yaml":          true,
		"README.md":             true,
		"lessons/objectives.md": true,
		"tests/smoke.sh":        true,
		"LICENSE":               true,
		".env.example":          true,
	}
	foundFiles := make(map[string]bool)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar: %v", err)
		}
		foundFiles[header.Name] = true
	}

	for f := range foundFiles {
		if !allowed[f] {
			t.Errorf("unexpected file in archive: %s", f)
		}
	}
}

func TestVerifyPackage(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "verify-lab")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	os.WriteFile(filepath.Join(templateDir, "lab.yaml"), []byte("schemaVersion: labkit/v1\nmetadata:\n  id: test/verify-lab\n  name: Verify\n  version: 1.0.0\n  category: web\n  difficulty: beginner\n  license: MIT\nlearning:\n  objectives:\n    - test\nruntime:\n  engine: docker-compose\n  composeFile: compose.yaml\n  architectures:\n    - amd64\n  resources:\n    memoryMiB: 256\n  network:\n    exposure: loopback\nendpoints:\n  - id: web\n    service: app\n    protocol: http\n    containerPort: 8080\n    hostBinding: 127.0.0.1\nchecks:\n  smokeTest: tests/smoke.sh\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "compose.yaml"), []byte("services:\n  app:\n    image: alpine:3.18\n    ports:\n      - \"127.0.0.1:8080:8080\"\n"), 0644)
	os.MkdirAll(filepath.Join(templateDir, "tests"), 0755)
	os.WriteFile(filepath.Join(templateDir, "tests", "smoke.sh"), []byte("#!/bin/sh\nexit 0\n"), 0644)

	packager := NewPackager()
	result, err := packager.Package(templateDir)
	if err != nil {
		t.Fatalf("failed to package: %v", err)
	}

	// Verify the package
	metadata, err := VerifyPackage(result.Archive, result.MetadataFile, result.Checksum)
	if err != nil {
		t.Fatalf("failed to verify package: %v", err)
	}
	if metadata.TemplateID != "test/verify-lab" {
		t.Errorf("expected templateID test/verify-lab, got %s", metadata.TemplateID)
	}

	// Verify with wrong checksum should fail
	if _, err := VerifyPackage(result.Archive, result.MetadataFile, "wrong-checksum"); err == nil {
		t.Error("expected checksum verification to fail for wrong checksum")
	}
}

func TestInstallPackage(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "install-lab")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	os.WriteFile(filepath.Join(templateDir, "lab.yaml"), []byte("schemaVersion: labkit/v1\nmetadata:\n  id: test/install-lab\n  name: Install\n  version: 1.0.0\n  category: web\n  difficulty: beginner\n  license: MIT\nlearning:\n  objectives:\n    - test\nruntime:\n  engine: docker-compose\n  composeFile: compose.yaml\n  architectures:\n    - amd64\n  resources:\n    memoryMiB: 256\n  network:\n    exposure: loopback\nendpoints:\n  - id: web\n    service: app\n    protocol: http\n    containerPort: 8080\n    hostBinding: 127.0.0.1\nchecks:\n  smokeTest: tests/smoke.sh\n"), 0644)
	os.WriteFile(filepath.Join(templateDir, "compose.yaml"), []byte("services:\n  app:\n    image: alpine:3.18\n    ports:\n      - \"127.0.0.1:8080:8080\"\n"), 0644)
	os.MkdirAll(filepath.Join(templateDir, "tests"), 0755)
	os.WriteFile(filepath.Join(templateDir, "tests", "smoke.sh"), []byte("#!/bin/sh\nexit 0\n"), 0644)

	packager := NewPackager()
	result, err := packager.Package(templateDir)
	if err != nil {
		t.Fatalf("failed to package: %v", err)
	}

	// Install into a new directory
	installDir := filepath.Join(dir, "installed")
	if err := Install(result.Archive, installDir); err != nil {
		t.Fatalf("failed to install: %v", err)
	}

	// Verify installed files exist
	installedFiles := []string{"lab.yaml", "compose.yaml", "tests/smoke.sh"}
	for _, f := range installedFiles {
		if _, err := os.Stat(filepath.Join(installDir, f)); os.IsNotExist(err) {
			t.Errorf("installed file missing: %s", f)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "my-lab")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}
	// Create a minimal lab.yaml for manifest validation
	os.WriteFile(filepath.Join(templateDir, "lab.yaml"), []byte("schemaVersion: labkit/v1\nmetadata:\n  id: test/checksum-lab\n  name: Checksum\n  version: 1.0.0\n  category: web\n  difficulty: beginner\n  license: MIT\nlearning:\n  objectives:\n    - test\nruntime:\n  engine: docker-compose\n  composeFile: compose.yaml\n  architectures:\n    - amd64\n  resources:\n    memoryMiB: 256\n  network:\n    exposure: loopback\nendpoints:\n  - id: web\n    service: app\n    protocol: http\n    containerPort: 8080\n    hostBinding: 127.0.0.1\nchecks:\n  smokeTest: tests/smoke.sh\n"), 0644)
	filePath := filepath.Join(templateDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	packager := NewPackager()
	result, err := packager.Package(templateDir)
	if err != nil {
		t.Fatalf("failed to package: %v", err)
	}

	// Verify the archive checksum
	if err := VerifyChecksum(result.Archive, result.Checksum); err != nil {
		t.Errorf("checksum verification failed: %v", err)
	}

	// Verify wrong checksum fails
	if err := VerifyChecksum(result.Archive, "invalid-checksum"); err == nil {
		t.Error("expected checksum verification to fail for invalid checksum")
	}
}

func TestPackageInvalidPath(t *testing.T) {
	packager := NewPackager()
	_, err := packager.Package("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}
