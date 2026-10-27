package generator

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Packager builds versioned archives and catalog entries for generated templates.
type Packager struct{}

// NewPackager creates a Packager.
func NewPackager() *Packager {
	return &Packager{}
}

// PackageResult holds the result of packaging a template.
type PackageResult struct {
	Archive      string
	ChecksumFile string
	Checksum     string
	Size         int64
}

// Package creates a versioned tar.gz archive with SHA-256 checksum.
func (p *Packager) Package(templatePath string, version string) (*PackageResult, error) {
	info, err := os.Stat(templatePath)
	if err != nil {
		return nil, fmt.Errorf("template path not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", templatePath)
	}

	if version == "" {
		version = "1.0.0"
	}

	templateName := filepath.Base(templatePath)
	archiveName := templateName + "-" + version + ".tar.gz"
	archivePath := filepath.Join(os.TempDir(), archiveName)

	if err := createTarGz(templatePath, archivePath); err != nil {
		return nil, fmt.Errorf("creating archive: %w", err)
	}

	checksum, err := sha256File(archivePath)
	if err != nil {
		return nil, fmt.Errorf("computing checksum: %w", err)
	}

	archiveInfo, err := os.Stat(archivePath)
	if err != nil {
		return nil, fmt.Errorf("stating archive: %w", err)
	}

	checksumFile := archivePath + ".sha256"
	if err := os.WriteFile(checksumFile, []byte(checksum+"  "+archiveName+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("writing checksum file: %w", err)
	}

	return &PackageResult{
		Archive:      archivePath,
		ChecksumFile: checksumFile,
		Checksum:     checksum,
		Size:         archiveInfo.Size(),
	}, nil
}

// createTarGz creates a gzipped tar archive of srcDir at dstPath.
func createTarGz(srcDir, dstPath string) error {
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gz := gzip.NewWriter(dstFile)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		header.Name = rel

		if d.IsDir() {
			header.Name += "/"
			return tw.WriteHeader(header)
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

// sha256File computes the SHA-256 hex digest of a file.
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
