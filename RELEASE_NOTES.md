# LabKit v0.1.0 Release Notes

## Release Date: 2026-10-27

## Overview

LabKit v0.1.0 is the initial stable release of the vulnerable environment platform. This release provides a secure, reproducible foundation for launching pre-built security education labs with a single command.

## What's New

### Core Platform
- **labctl CLI**: Complete command-line interface for managing lab environments
- **Template System**: Standardized, validated lab template contract
- **Policy Validation**: Enforced safety policies for loopback-only access, no privileged containers, no Docker socket mounts
- **Docker Compose Integration**: Uses Docker Compose v2 for container orchestration
- **Packaging**: Template packaging with SHA-256 checksums for distribution

### Available Commands
- `labctl doctor` - Check system prerequisites
- `labctl list` - List available labs
- `labctl info <id>` - Show lab details
- `labctl search <query>` - Search labs
- `labctl run <id>` - Launch a lab
- `labctl status` - Show running instances
- `labctl logs <id>` - View service logs
- `labctl stop <id>` - Stop a lab
- `labctl reset <id>` - Reset a lab
- `labctl remove <id>` - Remove a lab
- `labctl validate <path>` - Validate a template
- `labctl template create <name>` - Create a new template
- `labctl package <path>` - Package a template for distribution

### Included Labs
- **web/intro-lab**: Introduction to Web Security
  - Demonstrates reflected XSS
  - Demonstrates IDOR (Insecure Direct Object Reference)
  - Demonstrates stored XSS
  - Complete learning materials with objectives, concepts, and remediation guides

## Security Features

### Enforced Policies
- ✅ No privileged containers
- ✅ No Docker socket mounts
- ✅ No host networking by default
- ✅ Compose configuration validation
- ✅ Manifest schema validation

### Safety by Default
- All ports bind to loopback (127.0.0.1) by default
- Per-lab network isolation
- Resource limits enforced where supported
- Explicit cleanup with confirmation for destructive operations

## Platform Requirements

### Supported Operating Systems
- macOS 12+ (Monterey or later)
- Linux (tested on Ubuntu 20.04+, Debian 11+)
- Windows 10+ with WSL2

### Required Software
- Docker Engine 20.10+
- Docker Compose v2
- (For development) Go 1.21+

### Hardware Requirements
- Minimum: 2 CPU cores, 4GB RAM
- Recommended: 4 CPU cores, 8GB RAM
- Per lab: 256MB-1GB RAM, 0.25-1 CPU core (varies by lab)

## Installation

### Option 1: Build from Source
```bash
git clone https://github.com/labkit-project/vulnerable-lab-platform.git
cd vulnerable-lab-platform
go build -o labctl ./cmd/labctl
sudo mv labctl /usr/local/bin/
```

### Option 2: Download Binary
Download the appropriate binary for your platform from the [GitHub Releases](https://github.com/labkit-project/vulnerable-lab-platform/releases/tag/v0.1.0) page.

## Quick Start

```bash
# Check prerequisites
labctl doctor

# List available labs
labctl list

# Launch a lab
labctl run web/intro-lab

# View logs
labctl logs web/intro-lab

# Stop the lab
labctl stop web/intro-lab
```

## Known Limitations

### Policy Validation
- Port binding validation is currently skipped (relies on manual review)
- Host mount validation is simplified for MVP
- Future versions will include comprehensive Compose validation

### CLI Features
- No automatic port allocation (ports must be specified in compose.yaml)
- No persistent state tracking across restarts
- Limited error recovery in some failure scenarios

### Template Features
- No external template registry support
- No template version management
- No dependency resolution between templates

### Documentation
- No video tutorials
- Limited integration examples
- No interactive learning guides

## Breaking Changes from Pre-Releases

None - this is the first stable release.

## Migration Guide

No migration needed for first-time users.

## Testing

### Automated Tests
- Template validation: All templates pass schema validation
- Policy validation: Enforced policies are validated before launch
- Package round-trip: Labs can be packaged, extracted, and launched independently
- Smoke tests: web/intro-lab passes smoke tests on disposable infrastructure

### Manual Testing
- Tested on macOS 13 (Ventura)
- Tested on Ubuntu 22.04
- Tested with Docker Desktop 4.20+
- Tested with Docker Compose v2.20+

## Security Considerations

### Intentional Vulnerabilities
The included labs contain intentionally vulnerable applications designed for:
- Security education and training
- Authorized security research
- Vulnerability remediation practice

**Never deploy these environments in production or expose them to untrusted networks.**

### Isolation
LabKit provides container-level isolation but is not equivalent to a dedicated VM. For high-risk labs, we recommend running in a disposable VM with no sensitive host mounts.

### Trusted Distribution
- All templates in this release are reviewed and approved
- Images are pinned to specific versions
- SHA-256 checksums provided for all packages

## Documentation

- [Architecture](docs/architecture.md) - Platform architecture and component responsibilities
- [Template Authoring](docs/template-authoring.md) - Guide for creating new templates
- [Security Model](docs/security-model.md) - Security model and threat analysis
- [Package Format](docs/package-format.md) - Package format and distribution
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Areas for Contribution
- Additional lab templates
- Policy validation improvements
- Documentation improvements
- Bug fixes and enhancements

## Acknowledgments

- Inspired by [Vulhub](https://github.com/vulhub/vulhub)
- Built with [Docker Compose](https://docs.docker.com/compose/)
- Uses [compose-go](https://github.com/compose-spec/compose-go) for Compose parsing

## Support

- Documentation: https://github.com/labkit-project/vulnerable-lab-platform/docs
- Issues: https://github.com/labkit-project/vulnerable-lab-platform/issues
- Discussions: https://github.com/labkit-project/vulnerable-lab-platform/discussions

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Disclaimer

This project is for educational and authorized security research purposes only. Users are responsible for ensuring compliance with applicable laws and regulations. The authors accept no liability for misuse of these environments.

---

**Thank you for using LabKit!**
