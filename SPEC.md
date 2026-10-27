# LabKit Specification

## 1. Product Vision

LabKit is a platform for distributing pre-built, ready-to-use vulnerable environments that developers, students, security educators, and authorized security researchers can launch locally with a single command.

Each environment packages an intentionally vulnerable application or service, its dependencies, configuration, documentation, and validation checks into a repeatable Docker Compose project.

### Core Product Promise

- **One-command launch**: Start a lab without manually wiring containers together
- **Pluggable templates**: Add new environments through a documented, versioned contract
- **Guided learning**: Every lab includes objectives, background, expected results, and cleanup steps
- **Safe by default**: Isolated lab networking, local-only exposure, resource limits, and explicit warnings

## 2. Product Boundaries

The platform contains three distinct parts:

1. **The Runtime**: A small CLI (`labctl`) that validates, launches, inspects, resets, and removes environments
2. **The Template Ecosystem**: Independent, versioned lab packages that describe the application, services, ports, dependencies, resources, and learning material
3. **The Catalog and Delivery Pipeline**: A searchable registry, reusable templates, contributor validation, versioned releases, and documentation

### Architectural Decision

The runtime is kept separate from the environments. Adding a new lab should normally require a template directory and metadata—not a code change to the core CLI.

## 3. Core Product Capabilities

### A. One-Command Environment Launcher

Users select a lab ID and start it. The launcher checks prerequisites, retrieves the correct template version, validates configuration, starts services, waits for readiness, and prints the local URL.

```bash
labctl run web/intro-lab
```

**Expected output:**
```
Environment: web/intro-lab
Version:     1.0.0
Runtime:     Docker Compose

[check] Docker is available
[check] Compose configuration is valid
[check] Environment safety policies passed
[info]  Starting services...
[check]  App is ready

Lab URL: http://127.0.0.1:<allocated-port>

Use 'labctl logs web/intro-lab' to inspect logs.
Use 'labctl stop web/intro-lab' to stop the environment.
```

### B. Pluggable Template System

Each lab is an independent, versioned package. Contributors can add new applications, services, configurations, documentation, and validation tests without changing the launcher.

```bash
labctl template create my-lab
```

### C. Searchable Environment Catalog

Discover labs by technology, learning objective, difficulty, vulnerability category, architecture, and required resources. Display tested versions and compatibility before installation.

```bash
labctl search api
```

### D. Shippable Template Packages

Validate a template, package its files, generate checksums, publish a versioned release, and allow users to install it from a registry or local archive.

```bash
labctl package ./templates/my-lab
```

## 4. Recommended Architecture

Keep the first version modular, but do not build a large plugin framework before there are enough environments to justify one.

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        labctl CLI                           │
│                    (User Interface)                         │
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
│              (Metadata, Packages, Versions)                 │
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

### Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| `labctl` CLI | User-facing commands and useful errors |
| Manifest engine | Validate template metadata and schema versions |
| Catalog resolver | Discover templates and retrieve pinned releases |
| Compose adapter | Use Docker Compose v2 for configuration and lifecycle |
| Policy engine | Reject unsafe host access and unexpected network exposure |
| Lifecycle manager | Track instances, allocated ports, status, and cleanup |
| Template validator | Check manifests, Compose configuration, docs, and smoke tests |
| Package publisher | Build archives, checksums, and release metadata |

**Important**: Use Docker Compose as the container orchestration engine rather than implementing a competing one. It already provides service lifecycle management, inter-service DNS, networks, and configuration semantics.

## 5. Pluggable Template Contract

Every environment should follow the same package specification.

### Template Directory Structure

```
templates/
├── _scaffold/
│   ├── lab.yaml
│   ├── compose.yaml
│   ├── README.md
│   ├── lessons/
│   │   └── introduction.md
│   ├── tests/
│   │   └── smoke.sh
│   └── LICENSE
│
└── web/
    └── intro-lab/
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
        ├── .env.example
        └── LICENSE
```

### A. `lab.yaml`: Standardized Metadata

A template's manifest describes what the lab is and how the platform should manage it.

**Schema Version:** `labkit/v1`

**Example:**
```yaml
schemaVersion: "labkit/v1"
kind: Environment

metadata:
  id: web/intro-lab
  name: "Introduction to Web Security"
  version: "1.0.0"
  description: "A local learning environment for web security."
  category: web
  difficulty: beginner
  license: Apache-2.0

learning:
  objectives:
    - Understand common web security concepts
    - Identify insecure application behavior
    - Review defensive remediation techniques

runtime:
  engine: docker-compose
  composeFile: compose.yaml
  architectures:
    - amd64
    - arm64

  resources:
    memoryMiB: 1024
    cpu: 1

  network:
    exposure: loopback
    outbound: restricted

endpoints:
  - id: web-ui
    service: app
    protocol: http
    containerPort: 3000
    hostBinding: 127.0.0.1

checks:
  smokeTest: tests/smoke.sh
```

**Key Principles:**

- The actual Compose file remains the source of truth for services and runtime configuration
- The manifest provides catalog metadata, learning objectives, declared access points, compatibility requirements, and the location of checks
- The validator must compare the two so that metadata cannot claim an environment is isolated while the Compose configuration exposes it differently

### B. `compose.yaml`: The Executable Environment

The template defines the actual containers, images, volumes, dependencies, health checks, and networks.

**Standards for the first release:**

- Pin images to immutable digests where feasible, or use explicitly versioned images and a documented update process
- Bind web interfaces to `127.0.0.1` by default, not all host interfaces
- Keep databases and internal services unpublished unless the lab explicitly requires access to them
- Avoid host filesystem mounts and privileged containers unless a reviewed lab requirement genuinely depends on them
- Define resource limits where supported and configure health checks for services that need startup time

**Security Note:** Docker's documentation explicitly warns that publishing a port without a host IP binds it to all interfaces by default. That is a significant exposure risk for intentionally vulnerable applications.

### C. Extending the Platform

Support three template sources in a deliberate order:

| Source | Use case | Priority |
|--------|----------|----------|
| Built-in catalog | Trusted, maintained official labs | MVP |
| Local template path | Development and authoring | MVP |
| External versioned repository or package | Community extensions | After contract stabilizes |

**Initially**, use declarative templates rather than arbitrary executable plugin code. A template should not need permission to execute arbitrary host commands just to describe a lab.

**If extension hooks are added later**, require explicit declarations, limit supported hooks, and treat external hook code as untrusted. The loader should never run a template's setup script merely because it was downloaded.

## 6. CLI Experience and User Journey

### First-Time Experience

```bash
# Check Docker, Compose, architecture and basic prerequisites
labctl doctor

# Browse the built-in catalog
labctl search web

# Inspect a lab before launching it
labctl info web/intro-lab

# Launch the environment
labctl run web/intro-lab
```

### Lifecycle Commands

| Command | Behavior |
|---------|----------|
| `labctl list` | List available labs |
| `labctl info <id>` | Show requirements and safety details |
| `labctl run <id>` | Validate and start a new instance |
| `labctl status` | Show running instances |
| `labctl logs <id>` | Display service logs |
| `labctl stop <id>` | Stop a lab while retaining its state |
| `labctl reset <id>` | Recreate lab state after confirmation |
| `labctl remove <id>` | Remove a lab and optionally its volumes |
| `labctl template create <name>` | Generate a template from the scaffold |
| `labctl validate <path>` | Run local template validation |
| `labctl package <path>` | Build a distributable template archive |

**Important:** Destructive actions should have explicit confirmations. Deleting volumes should be a clearly documented operation because it removes persistent lab data.

## 7. Repository Structure

```
vulnerable-lab-platform/
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md
├── CODE_OF_CONDUCT.md
├── LICENSE
├── CHANGELOG.md
│
├── cmd/
│   └── labctl/
│
├── internal/
│   ├── catalog/
│   ├── manifest/
│   ├── compose/
│   ├── lifecycle/
│   ├── policy/
│   ├── state/
│   └── packaging/
│
├── schemas/
│   ├── lab.schema.json
│   └── catalog.schema.json
│
├── catalog/
│   └── index.yaml
│
├── templates/
│   ├── _scaffold/
│   ├── web/
│   ├── api/
│   ├── authentication/
│   ├── databases/
│   └── misconfiguration/
│
├── docs/
│   ├── architecture.md
│   ├── template-authoring.md
│   ├── security-model.md
│   ├── package-format.md
│   └── troubleshooting.md
│
├── scripts/
│   ├── validate-all.sh
│   └── build-release.sh
│
└── .github/
    ├── workflows/
    │   ├── validate.yml
    │   ├── template-smoke-test.yml
    │   └── release.yml
    ├── ISSUE_TEMPLATE/
    └── pull_request_template.md
```

**Implementation Language:** Go is recommended for a portable CLI distributed as a single executable. Python is also viable if rapid iteration and a simpler initial implementation matter more than a self-contained binary.

**Do not create separate services** for the catalog, package registry, and lifecycle manager in the first release. They can remain modules in the CLI until usage or maintenance needs justify independent components.

## 8. Security Model

The project intentionally distributes vulnerable applications, so safe defaults must be part of the platform design and enforced automatically.

### Local-Only Exposure

All published application ports bind to loopback by default. Block `0.0.0.0`, public host bindings, and unsupported host-network configurations unless a separately reviewed feature explicitly enables them.

### Per-Lab Isolation

Give each running instance its own project namespace, network, and volumes. Keep internal data services off host-published ports. Use internal networks where outbound connectivity is unnecessary and compatible with the exercise.

### Restricted Container Privileges

Reject privileged mode, Docker socket mounts, and broad host filesystem mounts by default. Drop unnecessary Linux capabilities, prefer read-only filesystems where supported, and document every exception.

### Trusted Distribution

Pin package versions and image references, verify package checksums, record upstream sources and license notices, and warn before running community templates that have not passed the platform's checks.

**Important Note:** Docker containers share the host kernel, and Docker daemon access can provide powerful host-level capabilities. Therefore, Compose isolation is useful but should not be presented as equivalent to a dedicated virtual machine. Rootless Docker can reduce some risks associated with daemon and container operation, where the host setup supports it.

For high-risk labs, recommend a disposable VM with no sensitive host mounts and controlled network access. The project should focus on lab-only research, education, demonstrations, and remediation—not production deployment of intentionally vulnerable services.

### Policy Validation Before Launch

The launcher should inspect the resolved Compose configuration and reject unsafe settings before creating containers.

**Initial Policy Rules:**

1. No privileged containers
2. No Docker socket mounts
3. No host networking or host PID namespace by default
4. No unrestricted host filesystem mounts
5. No published ports that violate loopback-only defaults
6. No undeclared external networks
7. No committed secrets in manifests or environment files
8. No unpinned production-style image references unless an explicit, reviewed exception exists
9. No template hooks executing by default
10. Enforce declared resource limits where technically supported

A blocked environment should produce a useful error explaining which policy failed and what configuration needs review.

## 9. CI/CD and Quality Gates

The repository should validate templates independently from the core CLI.

### Validation Pipeline

1. **Pull request opened**: Manifest, Compose file, documentation and tests change
2. **Static validation**: Schema · Compose syntax · Safety policies · Secret scan · License checks
3. **Disposable smoke test**: Build or pull images · Start services · Verify health · Check declared ports
4. **Cleanup and results**: Collect relevant logs · Remove test instances · Verify cleanup · Report compatibility
5. **Release**: Versioned package · SHA-256 checksum · Catalog update · Release notes

### Automated Checks

| Check | Required result |
|-------|-----------------|
| Manifest schema | Every required field is valid |
| Compose validation | Configuration resolves successfully |
| Policy validation | No unapproved risky settings |
| Documentation | Setup, learning objectives, and cleanup instructions exist |
| Smoke test | Services reach expected readiness |
| Port check | Only declared, allowed interfaces are exposed |
| Package integrity | Archive, checksum, and version metadata agree |
| Dependency checks | Launcher dependencies and packaging are reviewed |
| Cleanup check | Test containers, networks, and volumes are removed |
| Compatibility | Tested architectures and supported runtime versions are documented |

**Important:** Do not reject a deliberately vulnerable target merely because a vulnerability scanner identifies the weakness the lab is intended to teach. Separate expected target weaknesses from unexpected packaging risks, exposed host resources, secrets, and vulnerabilities in the launcher itself.

Run smoke tests on disposable CI workers with resource limits and controlled network permissions. Avoid running untrusted templates with secrets or access to persistent release credentials.

## 10. Development Roadmap

Use milestone-based development so every phase produces something testable.

### M0: Specification
Finalize the product and contract. Define supported host platforms, CLI commands, the manifest schema, template lifecycle, safety policies, versioning, and packaging format.

**Exit criteria:** An approved specification and a minimal valid template fixture.

### M1: Core MVP
Build the first working launcher. Implement prerequisite detection, manifest validation, Compose configuration checks, run, status, logs, stop, reset, and cleanup. Start with one beginner-friendly, intentionally insecure web application and its supporting services.

**Exit criteria:** A new user can launch, access, stop, and reset the sample lab using documented commands.

### M2: Template SDK
Make templates first-class packages. Generate new templates from the scaffold, add contract validation, support local template paths, and implement a catalog index.

**Exit criteria:** An independent contributor can build a new lab without modifying core runtime code.

### M3: Shipping
Implement release and distribution. Publish versioned template archives, checksums, catalog metadata, installation instructions, and compatibility information. Add automated release validation.

**Exit criteria:** A packaged lab can be installed and launched independently of the development checkout.

### M4: Expansion
Grow the lab collection. Add a small set of well-documented labs across web security, API security, authentication, databases, and misconfiguration. Add architecture compatibility reports and regression tests.

**Exit criteria:** Consistent quality and lifecycle management across multiple independent templates.

### M5: Ecosystem
Support community distribution. Add external repository imports, template conformance tooling, author documentation, contribution review, package provenance, and optional web-based catalog browsing.

**Exit criteria:** Independently maintained templates can be discovered, reviewed, validated, and released consistently.

## 11. First-Release Definition of Done

Use these as acceptance criteria before calling the MVP ready:

- A new user can check prerequisites with `labctl doctor`
- One command launches a documented lab and prints its actual local URL
- The template contract is machine-validated
- The runtime uses Docker Compose rather than implementing its own container engine
- The default network exposure is loopback-only, with policy checks enforced before launch
- Every running instance can be inspected, stopped, reset, and removed
- Destructive cleanup is explicit and documented
- A third party can create and validate a lab without editing the CLI
- A template can be packaged and distributed with its manifest, documentation, tests, version, and checksum
- CI verifies template validity, readiness, policy compliance, and cleanup
- All examples clearly distinguish training environments from production services
- Known limitations, image sources, licenses, architecture support, and expected resource requirements are documented

## 12. What Not to Build Initially

Keep the first release focused. Avoid:

- A web control panel
- Kubernetes support
- A cloud-hosted deployment service
- Arbitrary runtime plugins
- Multi-user orchestration
- A custom container scheduler

Those are potential future extensions, not prerequisites for a useful local lab platform. The most valuable initial proof is that a person can discover a lab, start it, learn from it, and clean it up without hand-editing Compose files.

## 13. Recommended Next Implementation Step

Create the repository around four initial deliverables:

1. **SPEC.md** — Product requirements, template contract, package format, CLI behavior, and compatibility policy
2. **schemas/lab.schema.json** — Versioned manifest validation
3. **templates/_scaffold/** — A complete, copyable lab template with Compose configuration, learning material, smoke tests, and cleanup guidance
4. **A minimal `labctl` prototype** — Prerequisite detection, `validate`, `run`, `status`, `logs`, `stop`, and `remove`

Prove the full workflow using one deliberately vulnerable, local-only training application before expanding the catalog. Once the template format, lifecycle, and safety checks work end-to-end, adding more environments becomes a content and package-maintenance task rather than repeated infrastructure development.

## 14. Technical References

- [Vulhub — pre-built vulnerable Docker Compose environments](https://github.com/vulhub/vulhub)
- [Docker Compose networking](https://docs.docker.com/compose/how-tos/networking/)
- [Docker Compose service configuration](https://docs.docker.com/reference/compose-file/services/)
- [Docker Engine security](https://docs.docker.com/engine/security/)
- [Docker rootless mode](https://docs.docker.com/engine/security/rootless/)
- [OWASP Juice Shop guidance](https://devguide.owasp.org/en/07-training-education/01-vulnerable-apps/01-juice-shop/)

## 15. Key Design Principle

**One runtime, one stable template contract, many independently shippable labs.**

This gives you a manageable MVP today and a credible path to a community-maintained vulnerable-environment ecosystem later.
