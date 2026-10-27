# Package Format

## Overview

LabKit packages are distributable archives containing complete lab templates with metadata and checksums for integrity verification.

## Package Structure

A LabKit package is a tar.gz archive with the following structure:

```
labkit-category-labname-20261027-120000.tar.gz
├── lab.yaml
├── compose.yaml
├── README.md
├── lessons/
│   ├── objectives.md
│   ├── concepts.md
│   └── remediation.md
├── tests/
│   ├── smoke.sh
│   └── expected-results.json
└── LICENSE
```

## Naming Convention

Package filenames follow this pattern:

```
labkit-<category>-<labname>-<timestamp>.tar.gz
```

Example: `labkit-web-intro-lab-20261027-120000.tar.gz`

The timestamp is in format: `YYYYMMDD-HHMMSS`

## Checksums

Each package includes a SHA-256 checksum file:

```
labkit-category-labname-20261027-120000.tar.gz.sha256
```

The checksum file contains a single line:

```
<sha256-hash>  labkit-category-labname-20261027-120000.tar.gz
```

## Creating a Package

Use the `labctl package` command:

```bash
labctl package ./templates/web/intro-lab
```

This will:

1. Create a tar.gz archive of the template
2. Calculate the SHA-256 checksum
3. Generate a checksum file
4. Display the checksum for verification

Example output:

```
Packaging template: ./templates/web/intro-lab

Creating archive...
Calculating checksum...

✓ Package created successfully
Archive: labkit-web-intro-lab-20261027-120000.tar.gz
Checksum: labkit-web-intro-lab-20261027-120000.tar.gz.sha256
SHA-256: a1b2c3d4e5f6...
```

## Verifying a Package

To verify a package's integrity:

```bash
# On Linux/macOS
sha256sum -c labkit-web-intro-lab-20261027-120000.tar.gz.sha256

# On macOS (alternative)
shasum -a 256 -c labkit-web-intro-lab-20261027-120000.tar.gz.sha256
```

Expected output:

```
labkit-web-intro-lab-20261027-120000.tar.gz: OK
```

## Installing a Package

### From Local Archive

```bash
# Extract the archive
tar -xzf labkit-web-intro-lab-20261027-120000.tar.gz

# Move to templates directory
mv web-intro-lab templates/web/

# Validate
labctl validate templates/web/intro-lab
```

### From Remote Archive (Future)

```bash
# Download
curl -O https://example.com/packages/labkit-web-intro-lab.tar.gz

# Verify
sha256sum -c labkit-web-intro-lab.tar.gz.sha256

# Install
labctl install labkit-web-intro-lab.tar.gz
```

## Package Metadata

The package metadata is stored in the `lab.yaml` manifest:

```yaml
metadata:
  id: web/intro-lab
  name: "Introduction to Web Security"
  version: "1.0.0"
  description: "A local learning environment for web security."
  category: web
  difficulty: beginner
  license: Apache-2.0
  author: "LabKit Project"
```

## Catalog Integration

When you create a package, update the catalog index:

```yaml
# catalog/index.yaml
version: catalog/v1
lastUpdated: "2026-10-27T12:00:00Z"
labs:
  - id: web/intro-lab
    name: "Introduction to Web Security"
    version: "1.0.0"
    category: web
    difficulty: beginner
    path: templates/web/intro-lab
    description: "A local learning environment for web security"
    tags:
      - web
      - security
    architectures:
      - amd64
      - arm64
    license: Apache-2.0
    author: "LabKit Project"
```

## Versioning

Package versions follow semantic versioning:

- **MAJOR**: Breaking changes to the lab structure or learning objectives
- **MINOR**: New features, additional vulnerabilities, or expanded content
- **PATCH**: Bug fixes, documentation updates, or minor improvements

Example:

```
1.0.0 → 1.0.1 (Documentation update)
1.0.1 → 1.1.0 (Added new vulnerability)
1.1.0 → 2.0.0 (Restructured lab completely)
```

## Distribution

### GitHub Releases

For official LabKit releases:

1. Create a GitHub release
2. Upload the package archive
3. Upload the checksum file
4. Update the catalog index
5. Announce in release notes

### Package Registry (Future)

Future enhancement: Integrate with a package registry:

- Public registry for community templates
- Private registry for internal labs
- Automatic metadata indexing
- Search and discovery

### CDN Distribution (Future)

Future enhancement: Use CDN for package distribution:

- Faster downloads
- Geographic distribution
- Bandwidth optimization
- Caching

## Package Signing (Future)

Future enhancement: Sign packages with GPG:

```bash
# Sign package
gpg --detach-sign --armor labkit-web-intro-lab.tar.gz

# Verify signature
gpg --verify labkit-web-intro-lab.tar.gz.asc labkit-web-intro-lab.tar.gz
```

## Package Dependencies

### Runtime Dependencies

Packages declare runtime dependencies in `lab.yaml`:

```yaml
runtime:
  dockerVersion: "20.10+"
  composeVersion: "2.0+"
  architectures:
    - amd64
    - arm64
```

### Template Dependencies

Future enhancement: Declare template dependencies:

```yaml
dependencies:
  - id: web/base-image
    version: "1.0.0"
```

## Package Size Guidelines

To ensure fast downloads and efficient storage:

- **Small labs**: < 50 MB (text-based apps, simple services)
- **Medium labs**: 50-200 MB (web apps with databases)
- **Large labs**: 200-500 MB (complex multi-service labs)
- **Maximum**: 1 GB (hard limit)

Optimization tips:

- Use multi-stage builds in Dockerfiles
- Minimize included assets
- Compress static files
- Use efficient base images

## Package Validation

Before distribution, validate packages:

```bash
# Validate template
labctl validate ./templates/web/intro-lab

# Run smoke test
cd templates/web/intro-lab
./tests/smoke.sh

# Create package
labctl package ./templates/web/intro-lab

# Verify checksum
sha256sum -c labkit-web-intro-lab.tar.gz.sha256

# Test installation
tar -xzf labkit-web-intro-lab.tar.gz
labctl validate web-intro-lab
```

## Package Lifecycle

### Creation

1. Develop lab template
2. Validate with `labctl validate`
3. Test with smoke tests
4. Create package with `labctl package`
5. Verify checksum
6. Update catalog

### Distribution

1. Upload to distribution channel
2. Update catalog index
3. Publish release notes
4. Announce to users

### Installation

1. Download package
2. Verify checksum
3. Extract to templates directory
4. Validate with `labctl validate`
5. Run with `labctl run`

### Updates

1. Create new version
2. Update version in lab.yaml
3. Create new package
4. Update catalog
5. Notify users

### Deprecation

1. Mark as deprecated in catalog
2. Provide migration guide
3. Remove from search results
4. Archive old packages

## Security Considerations

### Checksum Verification

Always verify checksums before installing:

```bash
sha256sum -c package.tar.gz.sha256
```

### Source Verification

Only install packages from trusted sources:

- Official LabKit repository
- Trusted community members
- Verified GitHub releases

### Content Scanning

Scan packages for malicious content:

```bash
# Scan with virus scanner
clamscan package.tar.gz

# Scan for secrets
trufflehog package.tar.gz
```

### Sandbox Testing

Test packages in isolated environment before production use:

1. Extract to test directory
2. Run in disposable VM
3. Monitor for suspicious behavior
4. Verify cleanup

## Troubleshooting

### Checksum Mismatch

If checksum verification fails:

1. Re-download the package
2. Verify the download URL
3. Check for corruption
4. Report to package maintainer

### Extraction Errors

If extraction fails:

1. Verify the file is a valid tar.gz
2. Check disk space
3. Try with verbose output: `tar -xvzf package.tar.gz`

### Validation Errors

If validation fails after installation:

1. Check all required files are present
2. Verify file permissions
3. Validate lab.yaml against schema
4. Review error messages

## Best Practices

### For Package Authors

1. **Version consistently**: Follow semantic versioning
2. **Document changes**: Maintain CHANGELOG
3. **Test thoroughly**: Run smoke tests before packaging
4. **Verify checksums**: Always verify after creation
5. **Update catalog**: Keep catalog in sync

### For Package Users

1. **Verify checksums**: Always verify before installation
2. **Trust sources**: Only install from trusted sources
3. **Validate after install**: Run `labctl validate`
4. **Test in isolation**: Try in test environment first
5. **Report issues**: Report bugs to maintainers

## References

- [Semantic Versioning](https://semver.org/)
- [SHA-256 Checksums](https://en.wikipedia.org/wiki/SHA-2)
- [Docker Compose](https://docs.docker.com/compose/)
