# Platform Compatibility - LabKit v0.1.0

## Supported Platforms

### Operating Systems

| Platform | Version | Status | Notes |
|----------|---------|--------|-------|
| macOS | 12+ (Monterey or later) | ✅ Tested | Docker Desktop required |
| Linux | Ubuntu 20.04+ | ✅ Tested | Docker Engine or Docker Desktop |
| Linux | Debian 11+ | ✅ Tested | Docker Engine or Docker Desktop |
| Windows | 10+ with WSL2 | ⚠️ Community tested | Docker Desktop in WSL2 required |

### Container Runtimes

| Runtime | Version | Status | Notes |
|---------|---------|--------|-------|
| Docker Engine | 20.10+ | ✅ Required | Docker Compose v2 required |
| Docker Compose | v2.0+ | ✅ Required | Included with Docker Desktop |
| Docker Desktop | 4.20+ | ✅ Tested | macOS and Windows |

### CPU Architectures

| Architecture | Status | Notes |
|--------------|--------|-------|
| amd64/x86_64 | ✅ Supported | Primary target |
| arm64 | ✅ Supported | Apple Silicon and ARM Linux |

## Lab-Specific Requirements

### web/intro-lab

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| Memory | 256 MB | 512 MB |
| CPU | 0.25 cores | 0.5 cores |
| Disk | 200 MB | 500 MB |
| Runtime | ~45s startup | N/A |

## Known Issues

### macOS
- Port 8080 commonly used by other services (use 8081 instead)
- Docker Desktop must be running before using labctl

### Linux
- User must be in docker group or use sudo
- Some distributions require rootless Docker configuration

### Windows
- Only tested with WSL2
- Docker Desktop must be running in WSL2

## Installation Methods

### Pre-built Binaries

| Platform | Binary | Architecture |
|----------|--------|--------------|
| Linux | labctl-linux-amd64 | amd64 |
| macOS | labctl-darwin-amd64 | Intel |
| macOS | labctl-darwin-arm64 | Apple Silicon |
| Windows | labctl-windows-amd64.exe | amd64 |

### Build from Source

Requires Go 1.21+:
```bash
go build -o labctl ./cmd/labctl
```

## Security Considerations

### Isolation Level
LabKit provides container-level isolation, which is appropriate for:
- Educational environments
- Authorized security research
- Development and testing

For high-risk labs, we recommend:
- Disposable VMs
- No sensitive host mounts
- Controlled network access

### Risk Assessment

| Risk Level | Lab | Mitigation |
|------------|-----|------------|
| Low | web/intro-lab | Loopback-only binding, no host access |
| Medium | API labs | Loopback-only, resource limits |
| High | System-level labs | VM isolation recommended |

## Verification

### Prerequisites Check
```bash
labctl doctor
```

### Template Validation
```bash
labctl validate ./templates/web/intro-lab
```

### Smoke Test
```bash
cd templates/web/intro-lab
./tests/smoke.sh
```

## Reporting Compatibility Issues

If you encounter compatibility issues:
1. Run `labctl doctor` and note the output
2. Run `docker --version` and `docker compose version`
3. Run `uname -a` (Linux/macOS) or `systeminfo` (Windows)
4. Report at: https://github.com/labkit-project/vulnerable-lab-platform/issues

## Future Platform Support

Planned for future releases:
- ✅ Static binary builds (no dependencies)
- ✅ Better Linux distribution support
- ✅ Native Windows support (without WSL2)
- ✅ ARM32 support for embedded devices
