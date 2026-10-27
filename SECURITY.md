# Security Policy

## Supported Versions

Only the latest version of LabKit is actively maintained with security updates.

## Reporting a Vulnerability

If you discover a security vulnerability in LabKit itself (not in the intentionally vulnerable lab environments), please report it responsibly.

### How to Report

Send an email to: security@labkit-project.org

Please include:

- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes (if available)

### What to Expect

- We will acknowledge receipt within 48 hours
- We will provide a detailed response within 7 days
- We will coordinate a fix and disclosure timeline with you

### Security Policy for Lab Environments

**Important**: The lab environments distributed by LabKit are intentionally vulnerable applications designed for education and authorized security research. Vulnerabilities in these environments are expected and documented as learning objectives.

**Do not report**:
- Vulnerabilities that are documented as part of the lab's learning objectives
- Security weaknesses that the lab is designed to demonstrate
- Expected insecure configurations described in the lab documentation

**Do report**:
- Security vulnerabilities in the `labctl` CLI tool itself
- Issues with the platform's safety policies or validation
- Unexpected exposure risks not documented in the lab
- Problems with the template validation or packaging system

## Security Best Practices for Users

### Isolation

- Run labs in an isolated environment (VM or dedicated test machine)
- Do not run labs on production systems or systems with sensitive data
- Use rootless Docker where possible
- Consider network-level isolation for high-risk labs

### Access Control

- Never expose lab ports to public networks
- Keep all published ports bound to `127.0.0.1`
- Review the manifest's `network.exposure` field before running
- Be cautious with labs that require elevated privileges

### Data Safety

- Assume lab containers can access mounted volumes
- Do not mount sensitive host directories
- Use `labctl remove` with caution when deleting volumes
- Back up important data before running experimental labs

### Updates

- Keep Docker and Docker Compose updated
- Use the latest version of `labctl`
- Review lab manifest updates before upgrading
- Check for security advisories in the repository

## Platform Security Model

LabKit enforces several safety policies:

1. **Local-only exposure**: Ports bind to loopback by default
2. **Privilege restrictions**: Privileged containers rejected by default
3. **No host mounts**: Host filesystem mounts rejected by default
4. **No Docker socket**: Docker socket mounts rejected by default
5. **Resource limits**: Memory and CPU limits enforced where supported
6. **Network isolation**: Per-lab networks with restricted outbound access

These policies are validated before launch and cannot be bypassed without explicit configuration.

## Disclosure Policy

For vulnerabilities in the LabKit platform itself:

- We will follow responsible disclosure
- We will coordinate public disclosure with the reporter
- We will provide credit in security advisories (with permission)
- We will aim to release fixes within 30 days of confirmation

## Contact

For security-related questions not involving vulnerability reports, please open an issue with the `security` label.
