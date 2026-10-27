# LabKit - Pre-Built Vulnerable Environments

A collection of pre-built, ready-to-use vulnerable environments that developers, students, security educators, and authorized security researchers can launch locally with a single command.

## Overview

LabKit packages intentionally vulnerable applications and services into repeatable Docker Compose projects with documentation, validation checks, and learning objectives. Each environment is designed for security education, authorized testing, and remediation practice.

## Quick Start

```bash
# Check prerequisites
labctl doctor

# Browse available labs
labctl search web

# Inspect a lab before launching
labctl info web/intro-lab

# Launch the environment
labctl run web/intro-lab

# View logs
labctl logs web/intro-lab

# Stop the environment
labctl stop web/intro-lab

# Remove the environment (with optional volume cleanup)
labctl remove web/intro-lab
```

## Core Features

- **One-command launch**: Start a lab without manually wiring containers together
- **Pluggable templates**: Add new environments through a documented, versioned contract
- **Guided learning**: Every lab includes objectives, background, expected results, and cleanup steps
- **Safe by default**: Isolated lab networking, local-only exposure, resource limits, and explicit warnings

## Security Model

⚠️ **Important**: These environments contain intentionally vulnerable applications. They are designed for:

- Security education and training
- Authorized security research
- Vulnerability remediation practice
- Security demonstration and testing

**Never deploy these environments in production or expose them to untrusted networks.**

### Safety Features

- **Local-only exposure**: All published ports bind to `127.0.0.1` by default
- **Per-lab isolation**: Each instance has its own network namespace and volumes
- **Restricted privileges**: Privileged containers and Docker socket mounts are rejected by default
- **Policy validation**: Automatic safety checks before launch
- **Explicit cleanup**: Clear commands for stopping and removing environments

## Project Structure

```
vulnerable-lab-platform/
├── cmd/labctl/          # CLI entry point
├── internal/            # Core runtime modules
│   ├── catalog/         # Lab discovery and resolution
│   ├── manifest/        # Template metadata parsing
│   ├── compose/         # Docker Compose integration
│   ├── lifecycle/       # Instance management
│   ├── policy/          # Safety validation
│   ├── state/           # Instance state tracking
│   └── packaging/       # Template packaging and distribution
├── schemas/             # JSON Schema for manifests
├── catalog/             # Lab catalog index
├── templates/           # Lab template packages
│   ├── _scaffold/       # Template scaffolding
│   ├── web/             # Web security labs
│   ├── api/             # API security labs
│   └── ...
├── docs/                # Architecture and authoring guides
└── scripts/             # Validation and build scripts
```

## Template Contract

Each lab follows a standardized package specification:

```
templates/web/intro-lab/
├── lab.yaml              # Lab manifest (metadata, objectives, safety policies)
├── compose.yaml          # Docker Compose configuration
├── README.md            # Lab documentation
├── lessons/              # Learning materials
│   ├── objectives.md
│   ├── concepts.md
│   └── remediation.md
├── tests/                # Validation and smoke tests
│   ├── smoke.sh
│   └── expected-results.json
└── LICENSE
```

See [docs/template-authoring.md](docs/template-authoring.md) for the complete template contract.

## Development

### Prerequisites

- Docker Engine 20.10+
- Docker Compose v2
- Go 1.21+ (for building the CLI)

### Building

```bash
# Build the CLI
go build -o labctl ./cmd/labctl

# Run validation on all templates
./scripts/validate-all.sh
```

### Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by [Vulhub](https://github.com/vulhub/vulhub) and the OWASP Juice Shop project.

## Disclaimer

This project is for educational and authorized security research purposes only. Users are responsible for ensuring compliance with applicable laws and regulations. The authors accept no liability for misuse of these environments.
