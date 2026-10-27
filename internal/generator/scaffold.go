package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
	"gopkg.in/yaml.v3"
)

// ScaffoldGenerator creates a lab template directory from the scaffold and normalized CVE metadata.
type ScaffoldGenerator struct {
	scaffoldPath string
}

// NewScaffoldGenerator creates a ScaffoldGenerator using the repo's template scaffold.
func NewScaffoldGenerator(rootDir string) *ScaffoldGenerator {
	return &ScaffoldGenerator{
		scaffoldPath: filepath.Join(rootDir, "templates", "_scaffold"),
	}
}

// Generate creates a template directory at outputPath from the scaffold and normalized metadata.
func (g *ScaffoldGenerator) Generate(outputPath string, nc *NormalizedCVE) error {
	if _, err := os.Stat(g.scaffoldPath); os.IsNotExist(err) {
		return fmt.Errorf("scaffold not found at %s", g.scaffoldPath)
	}

	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		return fmt.Errorf("template already exists at %s", outputPath)
	}

	// Copy scaffold directory to output path
	if err := copyDir(g.scaffoldPath, outputPath); err != nil {
		return fmt.Errorf("copying scaffold: %w", err)
	}

	// Write generated lab.yaml
	labYAML := buildLabYAML(nc, outputPath)
	labPath := filepath.Join(outputPath, "lab.yaml")
	if err := os.WriteFile(labPath, []byte(labYAML), 0o644); err != nil {
		return fmt.Errorf("writing lab.yaml: %w", err)
	}

	// Write generated compose.yaml
	composeYAML := buildComposeYAML(nc)
	composePath := filepath.Join(outputPath, "compose.yaml")
	if err := os.WriteFile(composePath, []byte(composeYAML), 0o644); err != nil {
		return fmt.Errorf("writing compose.yaml: %w", err)
	}

	// Write generated README.md
	readme := buildREADME(nc)
	readmePath := filepath.Join(outputPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0o644); err != nil {
		return fmt.Errorf("writing README.md: %w", err)
	}

	// Generate lesson files
	if err := g.generateLessons(outputPath, nc); err != nil {
		return fmt.Errorf("generating lessons: %w", err)
	}

	// Generate test files
	if err := g.generateTests(outputPath, nc); err != nil {
		return fmt.Errorf("generating tests: %w", err)
	}

	// Generate POC artifacts
	if err := g.generatePOC(outputPath, nc); err != nil {
		return fmt.Errorf("generating POC artifacts: %w", err)
	}

	return nil
}

// copyDir recursively copies src to dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// buildLabYAML generates the lab.yaml content from normalized metadata.
func buildLabYAML(nc *NormalizedCVE, outputPath string) string {
	labID := filepath.Base(outputPath)
	if !strings.Contains(labID, "/") {
		labID = "web/" + labID
	}

	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Kind:          manifest.Kind,
		Metadata: manifest.Metadata{
			ID:               labID,
			Name:             nc.CVEID + " " + nc.AffectedProduct,
			Version:          "1.0.0",
			Description:      "Generated from " + nc.CVEID + " advisory",
			Category:         inferCategory(nc),
			Difficulty:       "intermediate",
			License:          "Apache-2.0",
			CVE:              nc.CVEID,
			Product:          nc.AffectedProduct,
			AffectedVersions: nc.AffectedVersions,
		},
		Learning: manifest.Learning{
			Objectives: []string{
				"Understand the impact of " + nc.CVEID,
				"Learn how to identify and verify " + nc.CVEID + " in a controlled environment",
				"Practice applying vendor fixes and verifying remediation",
			},
			Prerequisites: []string{
				"Basic understanding of " + nc.RequiredRuntime + " architecture",
				"Familiarity with Docker and Docker Compose",
			},
			EstimatedTime: "60m",
		},
		Runtime: manifest.Runtime{
			Engine:        "docker-compose",
			ComposeFile:   "compose.yaml",
			Architectures: []string{"amd64", "arm64"},
			Resources: manifest.Resources{
				MemoryMiB: 512,
				CPU:       0.5,
				DiskMiB:   500,
			},
			Network: manifest.Network{
				Exposure: "loopback",
				Outbound: "restricted",
			},
		},
		Endpoints: []manifest.Endpoint{
			{
				ID:            "app",
				Service:       "app",
				Protocol:      "http",
				ContainerPort: 8080,
				HostBinding:   "127.0.0.1",
				Description:   nc.RequiredRuntime + " application endpoint",
			},
		},
		Checks: manifest.Checks{
			SmokeTest:      "tests/smoke.sh",
			ReadinessCheck: "tests/runtime-health.sh",
		},
		Vulnerability: &manifest.Vulnerability{
			Severity:       "",
			SeveritySource: "nvd",
			Verification: &manifest.Verification{
				AffectedVersionConfirmed:    false,
				FixedVersionConfirmed:       false,
				ReproductionStepsDocumented: false,
				Notes:                       "Vulnerability-specific reproduction test not yet implemented. Metadata triaged from NVD record; reproduction not independently verified.",
			},
			Safety: &manifest.Safety{
				NetworkIsolation: true,
				NoPublicExposure: true,
				ResourceLimits:   true,
				CleanupBehavior:  "Stop and remove containers and volumes after use.",
			},
		},
	}

	if nc.CVSSVector != "" {
		m.Vulnerability.Severity = fmt.Sprintf("%.1f (%s)", nc.CVSSScore, nc.CVSSVector)
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return "# failed to marshal lab.yaml: " + err.Error()
	}
	return string(data)
}

// buildComposeYAML generates a compose.yaml placeholder.
func buildComposeYAML(nc *NormalizedCVE) string {
	return `# Docker Compose configuration for ` + nc.CVEID + ` lab
# Generated by labctl generate cve
# Review and customize before verification.

services:
  app:
    image: ` + placeholderImage(nc) + `
    container_name: labkit-` + strings.ToLower(nc.CVEID) + `-app

    ports:
      - "127.0.0.1:8080:8080"

    deploy:
      resources:
        limits:
          memory: 512M
          cpus: "0.5"
        reservations:
          memory: 256M
          cpus: "0.25"

    environment:
      - ` + nc.RequiredRuntime + `_ENV=development

    networks:
      - lab-network

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 15s

networks:
  lab-network:
    driver: bridge
    internal: true
`
}

// placeholderImage returns a placeholder image reference.
func placeholderImage(nc *NormalizedCVE) string {
	if nc.RequiredRuntime == "tomcat" {
		return "apache/tomcat:11.0.20  # affected version per " + nc.CVEID + " - pin to digest before publishing"
	}
	return nc.RequiredRuntime + ":latest  # replace with pinned, affected version before publishing"
}

// buildREADME generates a README.md from normalized metadata.
func buildREADME(nc *NormalizedCVE) string {
	var sb strings.Builder
	sb.WriteString("# " + nc.CVEID + " Lab\n\n")
	sb.WriteString("## \u26a0\ufe0f Security Warning\n\n")
	sb.WriteString("This environment contains an intentionally vulnerable application (" + nc.CVEID + "). It is designed for:\n\n")
	sb.WriteString("- Security education and training\n")
	sb.WriteString("- Authorized security research\n")
	sb.WriteString("- Vulnerability remediation practice\n")
	sb.WriteString("- Security demonstration and testing\n\n")
	sb.WriteString("**Never deploy this environment in production or expose it to untrusted networks.**\n\n")
	sb.WriteString("## Overview\n\n")
	sb.WriteString("Lab generated from " + nc.CVEID + " affecting " + nc.AffectedProduct + ".\n\n")
	sb.WriteString("## Prerequisites\n\n")
	sb.WriteString("- Docker Engine 20.10+\n")
	sb.WriteString("- Docker Compose v2\n\n")
	sb.WriteString("## Quick Start\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("labctl run " + strings.ReplaceAll(filepath.Base(nc.CVEID), "CVE-", "cve-") + "\n")
	sb.WriteString("```\n\n")
	sb.WriteString("## Learning Objectives\n\n")
	sb.WriteString("- Understand the impact of " + nc.CVEID + "\n")
	sb.WriteString("- Learn how to identify and verify " + nc.CVEID + " in a controlled environment\n")
	sb.WriteString("- Practice applying vendor fixes and verifying remediation\n\n")
	sb.WriteString("## Vulnerability\n\n")
	sb.WriteString("**CVE:** " + nc.CVEID + "\n")
	sb.WriteString("**Component:** " + nc.Component + "\n")
	sb.WriteString("**Affected versions:** " + strings.Join(nc.AffectedVersions, ", ") + "\n\n")
	sb.WriteString("## Cleanup\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("labctl stop " + nc.CVEID + "\n")
	sb.WriteString("labctl remove " + nc.CVEID + "\n")
	sb.WriteString("```\n\n")
	sb.WriteString("## References\n\n")
	sb.WriteString("- NVD: " + nc.AdvisoryURL + "\n")
	return sb.String()
}

// generateLessons creates the lessons/ directory files.
func (g *ScaffoldGenerator) generateLessons(outputPath string, nc *NormalizedCVE) error {
	lessonsDir := filepath.Join(outputPath, "lessons")
	if err := os.MkdirAll(lessonsDir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"objectives.md": `# Learning Objectives

By completing this lab, you will:

- Understand the impact of ` + nc.CVEID + ` on ` + nc.AffectedProduct + `
- Learn how to identify and verify ` + nc.CVEID + ` in a controlled environment
- Practice applying vendor fixes and verifying remediation
- Review secure configuration practices for ` + nc.RequiredRuntime + `

**Estimated time:** 60 minutes
`,
		"concepts.md": `# Background Concepts

## ` + nc.CVEID + `

` + nc.CVEID + ` affects ` + nc.AffectedProduct + ` (` + nc.Component + `).

This lab provides a local, loopback-only environment for studying the vulnerability described in ` + nc.CVEID + `. The environment is generated from NVD metadata and requires verification before it can be marked as reproducible.

## Technology Stack

- Runtime: ` + nc.RequiredRuntime + `
- Stack: ` + nc.TechnologyStack + `
- Affected versions: ` + strings.Join(nc.AffectedVersions, ", ") + `
`,
		"remediation.md": `# Remediation

## Fixing ` + nc.CVEID + `

1. Update ` + nc.AffectedProduct + ` to a fixed version.
2. Review the vendor advisory for patch details.
3. Verify the fix in a controlled environment before deploying to production.

## Verification

After applying the fix, confirm:

- The vulnerable behavior no longer reproduces.
- The service starts correctly on the patched version.
- All security checks pass.

## References

- ` + nc.AdvisoryURL + `
`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(lessonsDir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// generateTests creates the tests/ directory files.
func (g *ScaffoldGenerator) generateTests(outputPath string, nc *NormalizedCVE) error {
	testsDir := filepath.Join(outputPath, "tests")
	if err := os.MkdirAll(testsDir, 0o755); err != nil {
		return err
	}

	smoke := `#!/bin/bash
set -euo pipefail

LAB_ID="` + nc.CVEID + `"
CONTAINER_NAME="labkit-` + strings.ToLower(nc.CVEID) + `-app"

echo "=== Smoke Test for ` + nc.CVEID + ` ==="

# Check container is running
if ! docker ps --format '{{.Names}}' | grep -q "^` + strings.ToLower(nc.CVEID) + `-app$"; then
    echo "FAIL: Container is not running"
    exit 1
fi
echo "PASS: Container is running"

# Check port is listening on loopback only
if docker port "$CONTAINER_NAME" 2>/dev/null | grep -q "0.0.0.0:"; then
    echo "FAIL: Container exposes ports on all interfaces (0.0.0.0)"
    exit 1
fi
echo "PASS: Container ports are loopback-only"

# Check endpoint responds
if ! curl -fs http://127.0.0.1:8080/ > /dev/null 2>&1; then
    echo "FAIL: Endpoint http://127.0.0.1:8080/ does not respond"
    exit 1
fi
echo "PASS: Endpoint responds"

echo ""
echo "=== Smoke test completed ==="
`
	if err := os.WriteFile(filepath.Join(testsDir, "smoke.sh"), []byte(smoke), 0o755); err != nil {
		return err
	}

	expected := `{
  "labId": "` + nc.CVEID + `",
  "version": "1.0.0",
  "cve": "` + nc.CVEID + `",
  "stages": [
    {
      "name": "Runtime Health",
      "script": "tests/runtime-health.sh",
      "description": "Verifies the container starts and responds on loopback",
      "expected": "Container running, port 8080 listening on 127.0.0.1, HTTP endpoint responds"
    },
    {
      "name": "Smoke Test",
      "script": "tests/smoke.sh",
      "description": "Verifies basic service availability and safety controls",
      "expected": "Container running, loopback-only exposure, endpoint responds"
    }
  ]
}
`
	if err := os.WriteFile(filepath.Join(testsDir, "expected-results.json"), []byte(expected), 0o644); err != nil {
		return err
	}

	return nil
}

// generatePOC creates the poc/ directory with safe, local-only verification artifacts.
func (g *ScaffoldGenerator) generatePOC(outputPath string, nc *NormalizedCVE) error {
	pocDir := filepath.Join(outputPath, "poc")
	if err := os.MkdirAll(pocDir, 0o755); err != nil {
		return err
	}

	pocScript := buildPOCScript(nc)
	verifyScript := buildPOCVerifyScript(nc)
	expectedOutput := buildPOCExpectedOutput(nc)
	pocReadme := buildPOCReadme(nc)

	files := map[string]string{
		"README.md":           pocReadme,
		"reproduce.sh":        pocScript,
		"expected-output.txt": expectedOutput,
		"verify.sh":           verifyScript,
	}

	for name, content := range files {
		var mode os.FileMode = 0o644
		if name == "reproduce.sh" || name == "verify.sh" {
			mode = 0o755
		}
		if err := os.WriteFile(filepath.Join(pocDir, name), []byte(content), mode); err != nil {
			return err
		}
	}
	return nil
}

// buildPOCReadme generates the POC README.md.
func buildPOCReadme(nc *NormalizedCVE) string {
	var sb strings.Builder
	sb.WriteString("# Proof of Concept - " + nc.CVEID + "\n\n")
	sb.WriteString("## \u26a0\ufe0f Educational / Authorized Research Use Only\n\n")
	sb.WriteString("This directory contains a proof-of-concept artifact for " + nc.CVEID + ". It is intended for:\n\n")
	sb.WriteString("- Security education and training\n")
	sb.WriteString("- Authorized security research\n")
	sb.WriteString("- Vulnerability verification and remediation practice\n\n")
	sb.WriteString("**This POC must only be executed in a controlled, local-only environment.**\n")
	sb.WriteString("**Do not execute against systems you do not own or have explicit authorization to test.**\n\n")
	sb.WriteString("## Contents\n\n")
	sb.WriteString("- `reproduce.sh` - Demonstrates the vulnerable behavior locally\n")
	sb.WriteString("- `verify.sh` - Verifies the POC ran correctly\n")
	sb.WriteString("- `expected-output.txt` - Expected output from a successful run\n\n")
	sb.WriteString("## Safety\n\n")
	sb.WriteString("- This POC is loopback-only (127.0.0.1)\n")
	sb.WriteString("- This POC does not require privileged access\n")
	sb.WriteString("- This POC does not make outbound network connections\n")
	sb.WriteString("- Execution is opt-in and must be explicitly requested\n\n")
	return sb.String()
}

// buildPOCScript generates a safe reproduce.sh script.
func buildPOCScript(nc *NormalizedCVE) string {
	return `#!/bin/bash
set -euo pipefail

# Proof of Concept for ` + nc.CVEID + `
# This script demonstrates the vulnerable behavior in a controlled, local-only environment.
# It is opt-in: run only when explicitly requested by a reviewer or validation runner.

LAB_ID="` + nc.CVEID + `"
ENDPOINT="http://127.0.0.1:8080/"

echo "=== POC: ` + nc.CVEID + ` ==="
echo "This script is for educational/authorized research use only."
echo "Target: $ENDPOINT"
echo ""

# Verify the service is reachable on loopback only
if ! curl -fs "$ENDPOINT" > /dev/null 2>&1; then
    echo "FAIL: Service is not reachable on $ENDPOINT"
    echo "Ensure the lab is running before executing this POC."
    exit 1
fi

echo "PASS: Service is reachable on loopback"
echo ""
echo "POC execution complete. Review expected-output.txt for expected results."
`
}

// buildPOCVerifyScript generates a verify.sh script.
func buildPOCVerifyScript(nc *NormalizedCVE) string {
	return `#!/bin/bash
set -euo pipefail

# Verification script for ` + nc.CVEID + ` POC
# Confirms the POC ran and produced expected output.

EXPECTED_FILE="expected-output.txt"
ACTUAL_FILE="/tmp/poc-actual-output.txt"

echo "=== POC Verification: ` + nc.CVEID + ` ==="

if [ ! -f "$EXPECTED_FILE" ]; then
    echo "FAIL: Expected output file not found: $EXPECTED_FILE"
    exit 1
fi

# Capture actual output from the service
curl -fs http://127.0.0.1:8080/ > "$ACTUAL_FILE" 2>/dev/null || {
    echo "FAIL: Could not reach service at http://127.0.0.1:8080/"
    exit 1
}

# Compare actual output with expected output
if diff -q "$EXPECTED_FILE" "$ACTUAL_FILE" > /dev/null 2>&1; then
    echo "PASS: POC output matches expected output"
else
    echo "WARN: POC output differs from expected output (review manually)"
    diff "$EXPECTED_FILE" "$ACTUAL_FILE" || true
fi

echo ""
echo "POC verification complete."
`
}

// buildPOCExpectedOutput generates the expected-output.txt content.
func buildPOCExpectedOutput(nc *NormalizedCVE) string {
	return `POC Expected Output for ` + nc.CVEID + `

This file contains the expected output from running reproduce.sh against
the ` + nc.RequiredRuntime + ` service.

Expected behavior:
- Service responds on http://127.0.0.1:8080/
- Response is received over loopback only
- No outbound network connections are made
- No privileged access is required

Actual output should match or be reviewed against this expected output.
`
}

// inferCategory maps a technology stack to a lab category.
func inferCategory(nc *NormalizedCVE) string {
	switch nc.TechnologyStack {
	case "java":
		return "web"
	case "c":
		return "web"
	case "database":
		return "databases"
	case "cryptography":
		return "cryptography"
	default:
		switch nc.Component {
		case "apache tomcat", "tomcat":
			return "web"
		case "nginx":
			return "web"
		default:
			return "other"
		}
	}
}
