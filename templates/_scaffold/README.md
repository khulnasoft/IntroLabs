# Lab Name

A brief description of what this lab teaches and its learning objectives.

## ⚠️ Security Warning

This environment contains intentionally vulnerable applications designed for:
- Security education and training
- Authorized security research
- Vulnerability remediation practice

**Never deploy this environment in production or expose it to untrusted networks.**

## Prerequisites

- Docker Engine 20.10+
- Docker Compose v2
- At least 512 MiB RAM available
- Basic understanding of [topic]

## Quick Start

```bash
# From the project root
labctl run category/lab-name

# Or manually with Docker Compose
cd templates/category/lab-name
docker compose up -d
```

The application will be available at: http://127.0.0.1:8080

## Learning Objectives

By completing this lab, you will learn:

1. First learning objective
2. Second learning objective
3. Third learning objective

## Lab Structure

```
templates/category/lab-name/
├── lab.yaml              # Lab manifest
├── compose.yaml          # Docker Compose configuration
├── README.md            # This file
├── lessons/              # Learning materials
│   ├── objectives.md    # Detailed objectives
│   ├── concepts.md      # Background concepts
│   └── remediation.md   # How to fix the vulnerabilities
├── tests/                # Validation tests
│   ├── smoke.sh         # Smoke test script
│   └── expected-results.json
└── LICENSE
```

## Usage

### Starting the Lab

```bash
labctl run category/lab-name
```

### Viewing Logs

```bash
labctl logs category/lab-name
```

### Stopping the Lab

```bash
labctl stop category/lab-name
```

### Resetting the Lab

```bash
labctl reset category/lab-name
```

### Removing the Lab

```bash
labctl remove category/lab-name
```

## Manual Docker Compose Usage

If you prefer to use Docker Compose directly:

```bash
# Start services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down

# Stop and remove volumes
docker compose down -v
```

## Verification

Run the smoke test to verify the lab is working correctly:

```bash
./tests/smoke.sh
```

## Expected Results

After completing this lab, you should be able to:

- [ ] Identify the vulnerability
- [ ] Exploit the vulnerability (if appropriate for the exercise)
- [ ] Understand the root cause
- [ ] Implement the remediation

## Cleanup

To completely remove the lab and all its data:

```bash
labctl remove category/lab-name --volumes
```

Or with Docker Compose:

```bash
docker compose down -v
```

## Troubleshooting

### Port Already in Use

If port 8080 is already in use, you can modify the port mapping in `compose.yaml`:

```yaml
ports:
  - "127.0.0.1:8081:8080"  # Use 8081 instead
```

### Container Won't Start

Check the logs:

```bash
docker compose logs app
```

### Health Check Failing

Ensure the application is responding correctly:

```bash
curl http://127.0.0.1:8080/health
```

## References

- [Reference Documentation](https://example.com/docs)
- [Additional Reading](https://example.com)

## License

This lab template is licensed under Apache-2.0.

## Contributing

To improve this lab or report issues, please refer to the main project's CONTRIBUTING.md.
