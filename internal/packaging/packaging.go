package packaging

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PackageResult contains the result of packaging a template.
type PackageResult struct {
	Archive      string
	Checksum     string
	ChecksumFile string
	Size         int64
	Metadata     PackageMetadata
	MetadataFile string
}

// PackageMetadata describes a packaged template.
type PackageMetadata struct {
	TemplateID     string   `json:"templateID"`
	Version        string   `json:"version"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	CreatedAt      string   `json:"createdAt"`
	LabSchema      string   `json:"labSchema"`
	Files          []string `json:"files"`
	ChecksumSHA256 string   `json:"checksumSHA256"`
	Packager       string   `json:"packager"`
}

// Packager creates distributable archives of lab templates.
type Packager struct{}

// NewPackager creates a packager.
func NewPackager() *Packager {
	return &Packager{}
}

// allowedFiles lists the file patterns that may be included in a template package.
var allowedFiles = []string{
	"lab.yaml",
	"compose.yaml",
	"README.md",
	"lessons/",
	"tests/",
	".env.example",
	"LICENSE",
}

// excludedPatterns lists file patterns that must never be included in a package.
var excludedPatterns = []string{
	".env",
	".git/",
	".gitignore",
	"__pycache__/",
	"*.pyc",
	".DS_Store",
	"state/",
	".labkit/",
	"*.log",
	"*.tmp",
	"cache/",
	"node_modules/",
	"vendor/",
}

// Package creates a tar.gz archive of a template directory with SHA-256 checksum
// and package metadata. Only allow-listed files are included; secrets, local
// state, caches, and unrelated files are excluded.
func (p *Packager) Package(templatePath string) (*PackageResult, error) {
	info, err := os.Stat(templatePath)
	if err != nil {
		return nil, fmt.Errorf("stating template path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", templatePath)
	}

	// Validate the template before packaging and get metadata
	manifest, err := loadManifestForPackaging(templatePath)
	if err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	archiveName := fmt.Sprintf("labkit-%s-%s.tar.gz", filepath.Base(templatePath), timestamp)

	// Collect allow-listed files
	files, err := collectAllowedFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("collecting files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no allow-listed files found in template: %s", templatePath)
	}

	// Create tar.gz archive
	if err := createArchive(templatePath, archiveName, files); err != nil {
		return nil, fmt.Errorf("creating archive: %w", err)
	}

	archiveInfo, err := os.Stat(archiveName)
	if err != nil {
		return nil, fmt.Errorf("stating archive: %w", err)
	}

	checksum, err := calculateChecksum(archiveName)
	if err != nil {
		return nil, fmt.Errorf("calculating checksum: %w", err)
	}

	// Write checksum file
	checksumFile := archiveName + ".sha256"
	if err := os.WriteFile(checksumFile, []byte(checksum+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("writing checksum file: %w", err)
	}

	// Create package metadata
	metadata := PackageMetadata{
		TemplateID:     manifest.Metadata.ID,
		Version:        manifest.Metadata.Version,
		Name:           manifest.Metadata.Name,
		Category:       manifest.Metadata.Category,
		CreatedAt:      timestamp,
		LabSchema:      manifest.SchemaVersion,
		Files:          files,
		ChecksumSHA256: checksum,
		Packager:       "labkit-packager v1",
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling metadata: %w", err)
	}

	metadataFile := archiveName + ".meta.json"
	if err := os.WriteFile(metadataFile, metadataJSON, 0644); err != nil {
		return nil, fmt.Errorf("writing metadata file: %w", err)
	}

	return &PackageResult{
		Archive:      archiveName,
		Checksum:     checksum,
		ChecksumFile: checksumFile,
		Size:         archiveInfo.Size(),
		Metadata:     metadata,
		MetadataFile: metadataFile,
	}, nil
}

// loadManifestForPackaging reads the lab.yaml manifest for metadata.
func loadManifestForPackaging(templatePath string) (*manifestWrapper, error) {
	manifestPath := filepath.Join(templatePath, "lab.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	// Minimal parse for metadata fields
	raw := &manifestWrapper{}

	// Simple YAML parsing without external dependency
	if err := simpleYAMLUnmarshal(data, raw); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	return raw, nil
}

type manifestWrapper struct {
	SchemaVersion string
	Metadata      struct {
		ID       string
		Name     string
		Version  string
		Category string
	}
}

// collectAllowedFiles walks the template directory and returns relative paths
// of files matching the allow-list, excluding any excluded patterns.
func collectAllowedFiles(templatePath string) ([]string, error) {
	var files []string

	err := filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == templatePath {
			return nil
		}

		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return err
		}

		// Skip excluded patterns
		if isExcluded(relPath, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (we only add files)
		if info.IsDir() {
			return nil
		}

		// Only include allow-listed files
		if isAllowed(relPath) {
			files = append(files, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// isExcluded checks if a path matches any excluded pattern.
func isExcluded(relPath string, isDir bool) bool {
	for _, pattern := range excludedPatterns {
		pattern = strings.TrimSuffix(pattern, "/")
		dirPattern := pattern + "/"

		if isDir {
			if relPath == pattern || strings.HasPrefix(relPath, dirPattern) {
				return true
			}
		} else {
			matched, _ := filepath.Match(pattern, relPath)
			if matched {
				return true
			}
			// Also check if the file is inside an excluded directory
			parts := strings.Split(relPath, string(filepath.Separator))
			for _, part := range parts {
				if part == pattern {
					return true
				}
			}
		}
	}
	return false
}

// isAllowed checks if a relative path matches any allow-list pattern.
func isAllowed(relPath string) bool {
	for _, pattern := range allowedFiles {
		if strings.HasSuffix(pattern, "/") {
			// Directory pattern: allow all files under this directory
			dirPrefix := strings.TrimSuffix(pattern, "/") + string(filepath.Separator)
			if strings.HasPrefix(relPath, dirPrefix) {
				return true
			}
		} else {
			matched, _ := filepath.Match(pattern, relPath)
			if matched {
				return true
			}
			// Also match files in subdirectories with the same basename
			base := filepath.Base(relPath)
			if base == pattern {
				return true
			}
		}
	}
	return false
}

// createArchive creates a tar.gz archive containing only the specified files.
func createArchive(source, target string, files []string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for _, relPath := range files {
		fullPath := filepath.Join(source, relPath)
		info, err := os.Stat(fullPath)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			f, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, f); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

func calculateChecksum(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// VerifyChecksum verifies that a file's SHA-256 matches the expected checksum.
func VerifyChecksum(filename, expectedChecksum string) error {
	actual, err := calculateChecksum(filename)
	if err != nil {
		return fmt.Errorf("calculating checksum: %w", err)
	}

	if actual != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actual)
	}

	return nil
}

// VerifyPackage verifies a package archive and its metadata file.
func VerifyPackage(archivePath, metadataPath, expectedChecksum string) (*PackageMetadata, error) {
	// Verify archive checksum
	if err := VerifyChecksum(archivePath, expectedChecksum); err != nil {
		return nil, fmt.Errorf("archive checksum verification failed: %w", err)
	}

	// Read and parse metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("reading metadata: %w", err)
	}

	var metadata PackageMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("parsing metadata: %w", err)
	}

	// Verify metadata checksum matches archive checksum
	if metadata.ChecksumSHA256 != expectedChecksum {
		return nil, fmt.Errorf("metadata checksum %s does not match archive checksum %s", metadata.ChecksumSHA256, expectedChecksum)
	}

	return &metadata, nil
}

// Install extracts a package archive into the local templates directory.
func Install(archivePath, installDir string) error {
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("creating install directory: %w", err)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}

		targetPath := filepath.Join(installDir, header.Name)

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", targetPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", filepath.Dir(targetPath), err)
		}

		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("creating file %s: %w", targetPath, err)
		}

		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return fmt.Errorf("writing file %s: %w", targetPath, err)
		}
		outFile.Close()
	}

	return nil
}

// simpleYAMLUnmarshal performs a minimal YAML unmarshal for the manifest wrapper.
// This avoids adding a YAML dependency to the packaging package.
func simpleYAMLUnmarshal(data []byte, v interface{}) error {
	// This is a simplified parser for the specific fields we need.
	// In production, this should use a proper YAML library.
	lines := strings.Split(string(data), "\n")
	currentSection := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect top-level keys
		if !strings.Contains(trimmed, ":") {
			continue
		}

		key, value, _ := strings.Cut(trimmed, ":")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"'")

		switch {
		case key == "schemaVersion":
			if sv, ok := v.(*manifestWrapper); ok {
				sv.SchemaVersion = value
			}
		case key == "metadata":
			currentSection = "metadata"
		case currentSection == "metadata":
			switch key {
			case "id":
				if sv, ok := v.(*manifestWrapper); ok {
					sv.Metadata.ID = value
				}
			case "name":
				if sv, ok := v.(*manifestWrapper); ok {
					sv.Metadata.Name = value
				}
			case "version":
				if sv, ok := v.(*manifestWrapper); ok {
					sv.Metadata.Version = value
				}
			case "category":
				if sv, ok := v.(*manifestWrapper); ok {
					sv.Metadata.Category = value
				}
			}
		}
	}

	return nil
}
