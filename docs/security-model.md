# Security Model

## Overview

LabKit is designed to safely distribute and run intentionally vulnerable applications for educational purposes. This document describes the security model, threat assumptions, and safety mechanisms.

## Threat Model

### What We Protect Against

LabKit is designed to protect against:

1. **Accidental exposure**: Users accidentally exposing vulnerable applications to public networks
2. **Unintended host access**: Vulnerable applications accessing the host system
3. **Resource exhaustion**: Labs consuming excessive host resources
4. **Privilege escalation**: Labs gaining elevated privileges on the host
5. **Data leakage**: Labs accessing sensitive host data

### What We Don't Protect Against

LabKit does NOT protect against:

1. **Deliberate misuse**: Users intentionally using labs for malicious purposes
2. **Zero-day vulnerabilities**: Unknown vulnerabilities in Docker or the host OS
5. **Physical access**: Attackers with physical access to the host
6. **Host compromise**: If the host is already compromised, labs provide no additional protection

### Trusted vs. Untrusted Components

**Trusted Components:**
- labctl CLI (runs on host)
- Docker Engine (runs on host)
- Host OS (provides isolation)

**Untrusted Components:**
- Lab containers (intentionally vulnerable)
- Lab application code
- Lab dependencies

## Defense in Depth

LabKit implements multiple layers of security:

### Layer 1: Template Contract

- Standardized manifest format
- Required metadata
- Declared endpoints
- Version tracking

### Layer 2: Policy Validation

- Pre-launch safety checks
- Configuration validation
- Exposure verification
- Resource limit enforcement

### Layer 3: Docker Isolation

- Container isolation
- Network namespaces
- Resource constraints
- Linux capabilities

### Layer 4: Network Restrictions

- Loopback-only bindings
- Per-lab networks
- Outbound restrictions
- No host networking

### Layer 5: Resource Limits

- Memory limits
- CPU limits
- Disk limits
- Cleanup guarantees

## Safety Policies

### Network Exposure Policy

**Default**: All published ports must bind to `127.0.0.1`

**Rationale**: Binding to all interfaces (`0.0.0.0`) exposes the application to the local network, which may include other devices or users.

**Implementation**:
```yaml
# Required
ports:
  - "127.0.0.1:8080:8080"

# Rejected
ports:
  - "8080:8080"
```

**Exceptions**: Requires explicit approval and documentation

### Privileged Container Policy

**Default**: Privileged containers are rejected

**Rationale**: Privileged containers have access to all host devices and can break out of container isolation.

**Implementation**:
```yaml
# Rejected
privileged: true
```

**Exceptions**: Requires security review and justification

### Docker Socket Mount Policy

**Default**: Docker socket mounts are rejected

**Rationale**: Mounting the Docker socket gives a container full control over the Docker daemon, including the ability to launch privileged containers on the host.

**Implementation**:
```yaml
# Rejected
volumes:
  - /var/run/docker.sock:/var/run/docker.sock
```

**Exceptions**: Never allowed

### Host Filesystem Mount Policy

**Default**: Unrestricted host filesystem mounts are rejected

**Rationale**: Host filesystem mounts can expose sensitive data or allow modification of host files.

**Implementation**:
```yaml
# Rejected
volumes:
  - /:/host
  - /home:/home

# Allowed (local project files only)
volumes:
  - ./config:/config
```

**Exceptions**: Only local, scoped mounts are allowed

### Host Networking Policy

**Default**: Host networking is rejected

**Rationale**: Host networking removes network isolation and can allow the container to access services bound to localhost on the host.

**Implementation**:
```yaml
# Rejected
network_mode: host
```

**Exceptions**: Requires explicit approval

### Resource Limit Policy

**Default**: Resource limits must be declared

**Rationale**: Without limits, a compromised container could consume all host resources.

**Implementation**:
```yaml
# Required
deploy:
  resources:
    limits:
      memory: 512M
      cpus: "0.5"
```

### Image Pinning Policy

**Default**: Images must be pinned to specific versions

**Rationale**: Unpinned images (e.g., `latest`) can change without warning, potentially introducing unexpected behavior or vulnerabilities.

**Implementation**:
```yaml
# Required
image: nginx:1.25
image: nginx:1.25@sha256:abc123...

# Discouraged
image: nginx:latest
```

## Container Isolation

### Linux Capabilities

LabKit containers run with minimal Linux capabilities:

- **Dropped by default**: All capabilities except those essential for the application
- **Audit**: Capabilities are logged for review

### Seccomp Profiles

Future enhancement: Apply seccomp profiles to restrict system calls.

### AppArmor/SELinux

Future enhancement: Integrate with AppArmor/SELinux for additional confinement.

### User Namespaces

Future enhancement: Use user namespaces for rootless containers.

## Network Isolation

### Per-Lab Networks

Each lab gets its own Docker bridge network:

- **Network name**: `labkit-<lab-id>-network`
- **Driver**: bridge
- **Internal**: Configurable (default: false for outbound access)

### Loopback-Only Binding

All application ports bind to `127.0.0.1`:

```yaml
ports:
  - "127.0.0.1:8080:8080"
```

### Outbound Access

Configurable outbound access:

- **Restricted**: Limited or no outbound access (default for most labs)
- **Unrestricted**: Full outbound access (when required)

### DNS

Containers use Docker's embedded DNS server. Future: customizable DNS settings.

## Resource Management

### Memory Limits

Each lab declares memory requirements:

```yaml
resources:
  memoryMiB: 512
```

Docker enforces these limits:

```yaml
deploy:
  resources:
    limits:
      memory: 512M
```

### CPU Limits

Each lab declares CPU requirements:

```yaml
resources:
  cpu: 0.5
```

Docker enforces these limits:

```yaml
deploy:
  resources:
    limits:
      cpus: "0.5"
```

### Disk Limits

Future enhancement: Implement disk usage limits using Docker volume drivers.

## State and Cleanup

### Instance Tracking

LabKit tracks running instances in a state file:

- **Location**: `~/.labkit/state.json`
- **Content**: Instance IDs, allocated ports, start times
- **Format**: JSON

### Cleanup Guarantees

When a lab is removed:

1. Containers are stopped
2. Networks are removed
3. Volumes are removed (if `--volumes` flag is used)
4. State is updated

### Orphaned Resources

LabKit provides commands to clean up orphaned resources:

```bash
labctl cleanup  # Future command
```

## Audit Logging

### Current Logging

- CLI commands are logged to stdout/stderr
- Container logs are accessible via `labctl logs`

### Future Logging

- Audit log of all lab launches
- Policy violation logging
- Resource usage tracking
- Security event logging

## Rootless Docker

LabKit supports rootless Docker (recommended):

- **Benefit**: Containers run without root privileges
- **Benefit**: Reduced attack surface
- **Benefit**: Better isolation
- **Requirement**: User must configure rootless Docker

## VM Isolation

For high-risk labs, LabKit recommends running in a disposable VM:

- **Benefit**: Complete isolation from host
- **Benefit**: Easy rollback
- **Benefit**: No persistent data on host
- **Tools**: VirtualBox, VMware, Vagrant, cloud VMs

## Incident Response

If a Lab is Compromised

1. **Stop the lab immediately**:
   ```bash
   labctl stop <lab-id>
   ```

2. **Remove the lab**:
   ```bash
   labctl remove <lab-id> --volumes
   ```

3. **Review logs**:
   ```bash
   labctl logs <lab-id>
   ```

4. **Check host**:
   - Review Docker daemon logs
   - Check for unusual processes
   - Review network connections

5. **Report**:
   - Document the incident
   - Report to LabKit maintainers if policy failed
   - Share findings with community

## Compliance Considerations

LabKit is designed for:

- **Educational use**: Security training and awareness
- **Authorized testing**: Penetration testing with authorization
- **Research**: Security research in controlled environments

LabKit is NOT designed for:

- **Production use**: Never deploy labs in production
- **Unauthorized testing**: Never use labs against targets without authorization
- **Malicious purposes**: Never use labs for harm

## Limitations

### Docker Limitations

LabKit inherits Docker's security limitations:

- Containers share the host kernel
- Root containers can escape isolation in some configurations
- Resource isolation is not perfect
- Network isolation has known limitations

### LabKit Limitations

- Policy enforcement is only as good as the policies
- Templates must be trusted (or reviewed)
- No runtime monitoring of container behavior
- No automated intrusion detection

## Future Enhancements

### Planned Security Features

1. **Seccomp Profiles**: Restrict system calls
2. **AppArmor/SELinux**: Additional confinement
3. **User Namespaces**: Rootless containers by default
4. **Runtime Monitoring**: Detect suspicious behavior
5. **Network Policies**: Fine-grained network control
6. **Secrets Management**: Secure secret handling
7. **Audit Logging**: Comprehensive audit trail

### Community Contributions

We welcome security-focused contributions:

- Additional policy rules
- Enhanced validation
- Better isolation mechanisms
- Security testing
- Documentation improvements

## References

- [Docker Security](https://docs.docker.com/engine/security/)
- [Container Security Best Practices](https://snyk.io/blog/10-docker-image-security-best-practices/)
- [OWASP Docker Security](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [Rootless Docker](https://docs.docker.com/engine/security/rootless/)
