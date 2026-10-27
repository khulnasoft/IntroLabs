# LabKit Architecture

## Overview

LabKit is designed as a modular platform with clear separation between the runtime CLI, template ecosystem, and delivery pipeline.

## Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        labctl CLI                           │
│                    (User Interface)                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Commands: doctor, list, info, search, run, status,   │  │
│  │           logs, stop, reset, remove, validate,       │  │
│  │           template, package                          │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Core Runtime                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Manifest     │  │ Catalog      │  │ Policy       │      │
│  │ Parser       │  │ Resolver     │  │ Validator    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐                          │
│  │ Lifecycle    │  │ Compose      │                          │
│  │ Manager      │  │ Adapter      │                          │
│  └──────────────┘  └──────────────┘                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Template Registry                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Built-in     │  │ Local        │  │ External     │      │
│  │ Catalog      │  │ Templates    │  │ Packages     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Safety and Validation                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Compose      │  │ Policy       │  │ Readiness     │      │
│  │ Checks       │  │ Engine       │  │ Checks        │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Docker Compose v2                          │
│  (Application containers, Supporting services, Networks)    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Local Lab Instance                         │
│         (Loopback-only access, Status, Logs, Reset)          │
└─────────────────────────────────────────────────────────────┘
```

## Module Responsibilities

### CLI Module (`cmd/labctl/`, `internal/cli/`)

**Responsibilities:**
- Parse command-line arguments
- Display user-friendly error messages
- Coordinate between runtime modules
- Provide interactive confirmations for destructive actions

**Key Commands:**
- `doctor`: Prerequisite checking
- `list/info/search`: Catalog browsing
- `run`: Lab instantiation
- `status/logs/stop/reset/remove`: Lifecycle management
- `validate`: Template validation
- `template`: Template scaffolding
- `package`: Distribution packaging

### Manifest Module (`internal/manifest/`)

**Responsibilities:**
- Parse `lab.yaml` manifests
- Validate against JSON Schema
- Extract metadata (version, dependencies, endpoints)
- Compare manifest with Compose configuration

**Key Functions:**
- Schema validation
- Metadata extraction
- Version comparison
- Compatibility checking

### Catalog Module (`internal/catalog/`)

**Responsibilities:**
- Resolve lab IDs to template paths
- Search and filter labs by metadata
- Handle multiple template sources (built-in, local, external)
- Manage catalog index

**Key Functions:**
- Lab discovery
- Metadata search
- Path resolution
- Catalog updates

### Compose Module (`internal/compose/`)

**Responsibilities:**
- Parse Docker Compose files
- Validate Compose syntax
- Extract service configuration
- Interact with Docker Compose CLI

**Key Functions:**
- Compose file parsing
- Service configuration extraction
- Port mapping analysis
- Network configuration review

### Policy Module (`internal/policy/`)

**Responsibilities:**
- Enforce safety policies before launch
- Reject unsafe configurations
- Validate declared vs. actual exposure
- Check for privileged containers, host mounts, etc.

**Key Policies:**
- Loopback-only port bindings
- No privileged containers
- No Docker socket mounts
- No unrestricted host filesystem mounts
- No host networking by default
- Resource limit enforcement

### Lifecycle Module (`internal/lifecycle/`)

**Responsibilities:**
- Track running lab instances
- Manage instance state
- Allocate and record ports
- Handle start, stop, reset, remove operations

**Key Functions:**
- Instance tracking
- State persistence
- Port allocation
- Cleanup orchestration

### State Module (`internal/state/`)

**Responsibilities:**
- Persist instance state to disk
- Manage state file format
- Handle state migrations
- Provide state queries

**State Information:**
- Running instances
- Allocated ports
- Start timestamps
- Template versions

### Packaging Module (`internal/packaging/`)

**Responsibilities:**
- Create template archives
- Generate checksums
- Build release metadata
- Validate package integrity

**Key Functions:**
- Archive creation (tar.gz)
- SHA-256 checksums
- Version metadata
- Package validation

## Data Flow

### Lab Launch Flow

```
User: labctl run web/intro-lab
    ↓
CLI: Parse command, validate arguments
    ↓
Catalog: Resolve 'web/intro-lab' to template path
    ↓
Manifest: Parse lab.yaml, validate schema
    ↓
Compose: Parse compose.yaml, extract services
    ↓
Policy: Compare manifest with Compose, enforce safety rules
    ↓
Lifecycle: Allocate port, record instance state
    ↓
Compose Adapter: docker compose up -d
    ↓
Readiness Check: Wait for health check / endpoint
    ↓
CLI: Display URL and usage instructions
```

### Validation Flow

```
User: labctl validate ./templates/web/intro-lab
    ↓
CLI: Parse command
    ↓
Manifest: Validate lab.yaml against schema
    ↓
File System: Check required files exist
    ↓
Compose: Validate compose.yaml syntax
    ↓
Policy: Check safety policies
    ↓
Tests: Execute smoke test (optional)
    ↓
CLI: Report validation results
```

## Security Architecture

### Defense in Depth

LabKit implements multiple layers of security:

1. **Template Contract**: Standardized, validated metadata
2. **Policy Engine**: Enforced safety rules before launch
3. **Docker Isolation**: Container-level isolation
4. **Network Restrictions**: Loopback-only bindings
5. **Resource Limits**: CPU and memory constraints

### Trust Boundaries

```
┌─────────────────────────────────────────┐
│           User's Host System           │
│  ┌───────────────────────────────────┐ │
│  │      labctl CLI (Trusted)         │ │
│  └───────────────────────────────────┘ │
│              │                         │
│  ┌───────────▼───────────────────────┐ │
│  │   Docker Engine (Trusted)         │ │
│  └───────────────────────────────────┘ │
│              │                         │
│  ┌───────────▼───────────────────────┐ │
│  │   Lab Containers (Untrusted)      │ │
│  │   - Isolated networks             │ │
│  │   - Loopback-only ports           │ │
│  │   - Resource limits               │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### Policy Enforcement Points

1. **Manifest Validation**: Schema-level checks
2. **Compose Analysis**: Configuration-level checks
3. **Runtime Verification**: Runtime-level checks
4. **Container Launch**: Docker-level checks

## Extension Points

### Future Template Sources

The catalog module is designed to support:

- Local file system (current)
- Built-in catalog (current)
- Git repositories (future)
- Package registries (future)
- HTTP endpoints (future)

### Future Policy Extensions

The policy module can be extended with:

- Custom policy rules
- Per-lab policy overrides
- User-configurable policies
- Policy templates

### Future Validation Extensions

The validation system can support:

- Custom validators
- Third-party security scanners
- Compliance checks
- Automated remediation suggestions

## Performance Considerations

### Startup Time

- Manifest parsing: < 100ms
- Compose validation: < 200ms
- Policy checks: < 100ms
- Container startup: Variable (depends on images)

### Memory Usage

- CLI memory: ~50MB
- Lab containers: Configurable (default 256MB-1GB)
- State file: < 1MB

### Disk Usage

- CLI binary: ~15MB
- Template packages: Variable (10MB-500MB)
- Container images: Variable (100MB-2GB)

## Scalability

### Concurrent Labs

- Limited by host resources (CPU, memory, disk)
- Each lab gets isolated network namespace
- Port allocation managed by lifecycle module

### Catalog Size

- Built-in catalog: No practical limit
- Search performance: O(n) for linear search (can be optimized with indexing)
- Template validation: Parallelizable

## Monitoring and Observability

### Logging

- CLI logs to stdout/stderr
- Container logs accessible via `labctl logs`
- State file tracks instance metadata

### Metrics (Future)

- Lab launch success rate
- Average startup time
- Resource utilization
- Policy violation counts

## Failure Modes

### Docker Unavailable

- Detected by `doctor` command
- Clear error message
- Suggest installation steps

### Template Invalid

- Detected by validation
- Specific error messages
- Suggestions for fixes

### Port Conflict

- Detected by lifecycle module
- Automatic port allocation (future)
- Manual port override option

### Container Startup Failure

- Logs displayed via `labctl logs`
- Health check timeout
- Automatic cleanup on failure
