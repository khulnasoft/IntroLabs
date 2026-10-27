# TODO — Vulnerable Lab Platform

## Mission

Build a collection of pre-built, ready-to-use vulnerable environments that anyone can launch with one command for authorized security research, learning, and demonstrations.

## Core Principle

One runtime. One template contract. Many independently shippable labs.

---

## Completed Tasks

### CLI and Runtime MVP

- [x] Implement `labctl version`.
- [x] Implement readiness checking for the declared health-check endpoint.
- [x] Detect declared host-port conflicts before startup.
- [x] Return actionable port-conflict errors with alternative suggestions.
- [x] Clean up Compose resources after startup or readiness failures.
- [x] Add lifecycle unit tests.
- [x] Verify `labctl list` and `labctl validate`.
- [x] Verify `labctl doctor` and `labctl info`.
- [x] Verify `labctl search`.
- [x] Verify `labctl run`, `stop`, `reset`, `remove`.
- [x] Verify `go build ./...`.
- [x] Verify `go vet ./...`.
- [x] Verify `go test ./...`.

### Runtime Reliability

- [x] Capture Docker Compose stderr for actionable diagnostics.
- [x] Add `BindFailureError` for Docker port-binding failures.
- [x] Add detailed readiness errors with URLs and HTTP status information.
- [x] Handle HTTP redirects during readiness checks.
- [x] Build health-check URLs using allocated Compose ports.
- [x] Detect declared host-port conflicts before startup.
- [x] Clean up resources after startup or readiness failures.
- [x] Add duplicate-run protection.
- [x] Add lifecycle unit tests covering error types and helper functions.

### Integration Testing

- [x] Add Docker Compose lifecycle integration tests.
- [x] Test validation, startup, readiness, status, stop, and removal.
- [x] Test duplicate-run rejection.
- [x] Skip integration tests gracefully when Docker is unavailable.
- [x] Verify integration-test code with `go vet -tags integration ./tests/...`.

### Documentation and Image Management

- [x] Document digest pinning for container images.
- [x] Add Go formatting validation to CI.
- [x] Add Go static analysis using `go vet ./...`.

### CI/CD

- [x] Run Go formatting checks.
- [x] Run Go static analysis.
- [x] Run the Go test suite.

### Runtime Security Gate (P0) — Complete

- [x] Validate the fully resolved Compose configuration before startup.
- [x] Reject privileged containers by default.
- [x] Reject Docker socket mounts and unexpected host filesystem mounts.
- [x] Reject host networking, host PID, and host IPC namespaces.
- [x] Enforce loopback-only published ports by default.
- [x] Validate declared endpoints against resolved service configuration.
- [x] Add negative tests for each prohibited configuration.
- [x] Test concurrent launches and port-binding races.
- [x] Test readiness timeout, cancellation, and cleanup behavior.
- [x] Verify cleanup does not affect unrelated Compose projects.

### Template Packaging (P1) — Complete

- [x] Define the template package format.
- [x] Implement package creation with file allow-listing.
- [x] Validate package contents before creation.
- [x] Exclude secrets, local state, caches, and unrelated files.
- [x] Generate package metadata and SHA-256 checksums.
- [x] Implement package integrity verification (VerifyPackage).
- [x] Install templates into a managed local catalog (Install).
- [x] Support pinned template versions.
- [x] Test package creation, installation, launch, and removal.

### CI/CD (P1) — Complete

- [x] Run unit tests and integration tests in appropriate CI jobs.
- [x] Add Docker Compose integration testing on disposable runners.
- [x] Validate all built-in templates.
- [x] Run security-policy checks against resolved Compose configurations.
- [x] Add package validation and integrity tests.
- [x] Capture diagnostics when integration tests fail.
- [x] Verify cleanup after successful and failed tests.

### Ecosystem (P2) — Complete

- [x] Finalize the template authoring guide.
- [x] Publish the template scaffold.
- [x] Establish a curated environment catalog.
- [x] Document supported platforms and architecture compatibility.
- [x] Add versioned releases and release checksums.
- [x] Define contribution and template review requirements.

---

## Definition of Done

A feature is complete only when:

- [ ] Its behavior and acceptance criteria are documented.
- [ ] Relevant automated tests pass.
- [ ] Errors are actionable and understandable.
- [ ] Security implications have been reviewed.
- [ ] Documentation is updated.
- [ ] Cleanup and failure handling are tested.
- [ ] Compatibility limitations are documented.
- [ ] CI passes on the pull request.
- [ ] Changes have been reviewed and approved.

(End of file)