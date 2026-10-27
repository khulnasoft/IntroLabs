# Template Authoring Guide

This guide explains how to create new lab templates for LabKit.

## Quick Start

The easiest way to create a new lab is to use the scaffold:

```bash
labctl template create my-lab
```

This copies the scaffold directory to `templates/my-lab` with all the required files.

## Template Structure

Every lab template must follow this structure:

```
templates/category/lab-name/
├── lab.yaml              # Lab manifest (required)
├── compose.yaml          # Docker Compose configuration (required)
├── README.md            # Lab documentation (required)
├── lessons/              # Learning materials (required)
│   ├── objectives.md    # Learning objectives (required)
│   ├── concepts.md      # Background concepts (required)
│   └── remediation.md   # Remediation guide (required)
├── tests/                # Validation tests (required)
│   ├── smoke.sh         # Smoke test script (required)
│   └── expected-results.json (optional)
├── .env.example          # Example environment variables (optional)
└── LICENSE               # License (required)
```

## Writing lab.yaml

The `lab.yaml` file is the manifest that describes your lab. It must conform to the JSON Schema in `schemas/lab.schema.json`.

### Required Fields

```yaml
schemaVersion: "labkit/v1"
kind: Environment

metadata:
  id: category/lab-name          # Unique identifier
  name: "Human-readable name"    # Display name
  version: "1.0.0"               # Semantic version
  description: "Brief description"
  category: web                  # Category: web, api, authentication, etc.
  difficulty: beginner           # beginner, intermediate, advanced, expert
  license: Apache-2.0            # SPDX identifier

learning:
  objectives:
    - First learning objective
    - Second learning objective

runtime:
  engine: docker-compose
  composeFile: compose.yaml
  architectures:
    - amd64
    - arm64
  resources:
    memoryMiB: 512
    cpu: 0.5
  network:
    exposure: loopback          # loopback, local, or public
    outbound: restricted       # restricted or unrestricted

endpoints:
  - id: web-ui
    service: app
    protocol: http
    containerPort: 8080
    hostBinding: 127.0.0.1

checks:
  smokeTest: tests/smoke.sh
```

### Best Practices

1. **ID Format**: Use `category/lab-name` format with lowercase, hyphen-separated names
2. **Versioning**: Follow semantic versioning (MAJOR.MINOR.PATCH)
3. **Categories**: Choose from: web, api, authentication, databases, misconfiguration, network, cryptography, injection, xss, csrf, other
4. **Resources**: Be conservative with resource limits to ensure the lab runs on modest hardware
5. **Network Exposure**: Always use `loopback` unless there's a specific educational reason for broader exposure

## Writing compose.yaml

The `compose.yaml` file defines the actual Docker services for your lab.

### Security Requirements

1. **Port Bindings**: Always bind to `127.0.0.1` unless explicitly documented otherwise
   ```yaml
   ports:
     - "127.0.0.1:8080:8080"  # ✓ Correct
     - "8080:8080"            # ✗ Wrong (binds to all interfaces)
   ```

2. **Image Pinning**: Pin images to specific versions or digests
   ```yaml
   image: nginx:1.25@sha256:abc123...  # ✓ Best
   image: nginx:1.25                    # ✓ Good
   image: nginx:latest                  # ✗ Avoid
   ```

3. **Resource Limits**: Define resource limits
   ```yaml
   deploy:
     resources:
       limits:
         memory: 512M
         cpus: "0.5"
   ```

4. **Health Checks**: Add health checks where possible
   ```yaml
   healthcheck:
     test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
     interval: 30s
     timeout: 10s
     retries: 3
   ```

5. **Avoid Privileges**: Don't use privileged mode unless absolutely necessary
   ```yaml
   privileged: true  # ✗ Avoid
   ```

6. **Avoid Host Mounts**: Don't mount host filesystem unless necessary
   ```yaml
   volumes:
     - /var/data:/data  # ✗ Avoid
     - ./config:/config  # ✓ OK for local files
   ```

### Example compose.yaml

```yaml
services:
  app:
    image: python:3.11-slim@sha256:digest
    container_name: labkit-category-labname-app

    ports:
      - "127.0.0.1:8080:8080"

    deploy:
      resources:
        limits:
          memory: 256M
          cpus: "0.25"

    environment:
      - APP_ENV=development

    networks:
      - lab-network

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3

networks:
  lab-network:
    driver: bridge
    internal: false
```

## Writing Documentation

### README.md

The README should include:

1. **Security Warning**: Always start with a warning that this is for educational purposes only
2. **Prerequisites**: What users need before running the lab
3. **Quick Start**: How to launch the lab
4. **Learning Objectives**: What users will learn
5. **Vulnerabilities**: What vulnerabilities are demonstrated
6. **Exercise Steps**: Step-by-step instructions
7. **Troubleshooting**: Common issues and solutions

### lessons/objectives.md

Detail the learning objectives:

- What concepts will be learned
- What skills will be developed
- Success criteria for each objective
- Estimated time to complete

### lessons/concepts.md

Provide background knowledge:

- Explain the vulnerability types
- Describe how they work
- Discuss real-world impact
- Provide context and examples

### lessons/remediation.md

Explain how to fix the vulnerabilities:

- Identify the vulnerable code
- Show remediation options
- Recommend the best approach
- Provide code examples
- Discuss defense in depth

## Writing Tests

### smoke.sh

The smoke test should verify:

1. Docker is available
2. Containers are running
3. Ports are listening
4. Endpoints are responding
5. Health checks pass

Example:

```bash
#!/bin/bash
set -e

LAB_ID="${LAB_ID:-my-lab}"
ENDPOINT="${ENDPOINT:-http://127.0.0.1:8080}"

# Check container is running
docker ps --format '{{.Names}}' | grep -q "^labkit-${LAB_ID}-app$"

# Check port is listening
nc -z 127.0.0.1 8080

# Check endpoint responds
curl -f "${ENDPOINT}/health"

echo "All checks passed"
```

Make it executable:

```bash
chmod +x tests/smoke.sh
```

### expected-results.json

Document expected test results:

```json
{
  "labId": "category/lab-name",
  "version": "1.0.0",
  "tests": [
    {
      "name": "Container Running",
      "description": "Verifies that the main application container is running",
      "expected": "Container should be in running state"
    }
  ]
}
```

## Validation

Validate your template before submitting:

```bash
labctl validate ./templates/my-lab
```

This checks:

- Required files exist
- lab.yaml is valid YAML
- compose.yaml exists
- smoke.sh is executable

## Packaging

Package your template for distribution:

```bash
labctl package ./templates/my-lab
```

This creates:

- A tar.gz archive
- A SHA-256 checksum file

## Common Mistakes

### 1. Port Binding to All Interfaces

**Wrong:**
```yaml
ports:
  - "8080:8080"
```

**Right:**
```yaml
ports:
  - "127.0.0.1:8080:8080"
```

### 2. Using Latest Tags

**Wrong:**
```yaml
image: nginx:latest
```

**Right:**
```yaml
image: nginx:1.25@sha256:digest
```

### 3. Missing Health Checks

**Wrong:**
```yaml
services:
  app:
    image: myapp:1.0
    # No health check
```

**Right:**
```yaml
services:
  app:
    image: myapp:1.0
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
```

### 4. Exposing Internal Services

**Wrong:**
```yaml
services:
  database:
    ports:
      - "127.0.0.1:5432:5432"
```

**Right:**
```yaml
services:
  database:
    # No ports - internal only
    networks:
      - lab-network
```

### 5. Inconsistent Metadata

**Wrong:**
```yaml
# lab.yaml says loopback
network:
  exposure: loopback

# compose.yaml binds to all interfaces
ports:
  - "8080:8080"
```

**Right:**
```yaml
# lab.yaml says loopback
network:
  exposure: loopback

# compose.yaml respects that
ports:
  - "127.0.0.1:8080:8080"
```

## Review Checklist

Before submitting your template, verify:

- [ ] All required files are present
- [ ] lab.yaml validates against the schema
- [ ] compose.yaml is valid Docker Compose syntax
- [ ] All ports bind to 127.0.0.1 (unless documented otherwise)
- [ ] Images are pinned to specific versions
- [ ] Resource limits are defined
- [ ] Health checks are implemented where possible
- [ ] No privileged containers (unless documented)
- [ ] No Docker socket mounts
- [ ] No unrestricted host filesystem mounts
- [ ] README.md is comprehensive
- [ ] Learning objectives are clear
- [ ] Concepts are well explained
- [ ] Remediation guide is detailed
- [ ] Smoke test passes
- [ ] smoke.sh is executable
- [ ] License is included

## Submission

To submit your template:

1. Fork the repository
2. Create a feature branch
3. Add your template to `templates/`
4. Update `catalog/index.yaml` with your lab
5. Run `labctl validate ./templates/your-lab`
6. Commit and push
7. Open a pull request

The CI system will automatically:
- Validate your template
- Run smoke tests
- Check for policy violations
- Verify cleanup

## Getting Help

- Check existing templates in `templates/` for examples
- Review the scaffold in `templates/_scaffold/`
- Consult the JSON Schema in `schemas/lab.schema.json`
- Open an issue for questions
