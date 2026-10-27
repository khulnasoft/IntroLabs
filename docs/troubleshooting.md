# Troubleshooting

This guide helps you diagnose and resolve common issues with LabKit.

## Installation Issues

### Go Build Fails

**Problem**: `go build` command fails with errors

**Solution**:
```bash
# Ensure Go 1.21+ is installed
go version

# Update Go modules
go mod tidy

# Build with verbose output
go build -v -o labctl ./cmd/labctl
```

### Missing Dependencies

**Problem**: Import errors or missing packages

**Solution**:
```bash
# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Tidy modules
go mod tidy
```

## Docker Issues

### Docker Not Found

**Problem**: `labctl doctor` reports Docker not found

**Solution**:
```bash
# Install Docker
# macOS: https://docs.docker.com/desktop/install/mac-install/
# Linux: https://docs.docker.com/engine/install/
# Windows: https://docs.docker.com/desktop/install/windows-install/

# Verify installation
docker --version
```

### Docker Daemon Not Running

**Problem**: `labctl doctor` reports Docker daemon not running

**Solution**:
```bash
# macOS/Windows: Start Docker Desktop

# Linux: Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Verify
docker info
```

### Docker Compose Not Available

**Problem**: `labctl doctor` reports Docker Compose not found

**Solution**:
```bash
# Docker Compose v2 is included with Docker Desktop
# Verify version
docker compose version

# If using standalone, install:
# https://docs.docker.com/compose/install/
```

### Permission Denied

**Problem**: `permission denied while trying to connect to the Docker daemon`

**Solution**:
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER

# Log out and log back in

# Or use sudo (not recommended)
sudo docker ps
```

### Rootless Docker Issues

**Problem**: Issues with rootless Docker

**Solution**:
```bash
# Check rootless Docker status
docker context use rootless

# Verify
docker info
```

## Lab Launch Issues

### Template Not Found

**Problem**: `labctl run web/intro-lab` reports template not found

**Solution**:
```bash
# Verify template exists
ls templates/web/intro-lab

# Check catalog
labctl list

# Use full path
labctl run ./templates/web/intro-lab
```

### Port Already in Use

**Problem**: `Error: port 8080 is already allocated`

**Solution**:
```bash
# Find what's using the port
lsof -i :8080

# Stop the conflicting service
# Or modify the port in compose.yaml:
ports:
  - "127.0.0.1:8081:8080"
```

### Container Won't Start

**Problem**: Container exits immediately or fails to start

**Solution**:
```bash
# Check logs
labctl logs web/intro-lab

# Or with Docker Compose
cd templates/web/intro-lab
docker compose logs app

# Check container status
docker ps -a

# Inspect container
docker inspect labkit-web-intro-lab-app
```

### Image Pull Failed

**Problem**: Failed to pull Docker image

**Solution**:
```bash
# Check Docker Hub status
curl https://hub.docker.com/

# Check network connectivity
ping registry-1.docker.io

# Pull manually
docker pull <image-name>

# Check disk space
df -h
```

### Health Check Timeout

**Problem**: Lab never becomes ready, health check times out

**Solution**:
```bash
# Check if service is actually running
curl http://127.0.0.1:8080/health

# Increase timeout in lab.yaml
checks:
  healthCheck:
    timeout: 60

# Remove health check temporarily
# In compose.yaml, comment out healthcheck section
```

## Validation Issues

### Schema Validation Failed

**Problem**: `labctl validate` reports schema validation errors

**Solution**:
```bash
# Check lab.yaml syntax
cat templates/web/intro-lab/lab.yaml

# Validate against schema manually
# (Use a YAML validator tool)

# Compare with scaffold
diff templates/_scaffold/lab.yaml templates/web/intro-lab/lab.yaml
```

### Required Files Missing

**Problem**: Validation reports missing required files

**Solution**:
```bash
# Check template structure
ls -la templates/web/intro-lab/

# Copy missing files from scaffold
cp templates/_scaffold/lessons/objectives.md templates/web/intro-lab/lessons/

# Ensure directories exist
mkdir -p templates/web/intro-lab/lessons
mkdir -p templates/web/intro-lab/tests
```

### Smoke Test Failed

**Problem**: Smoke test reports failures

**Solution**:
```bash
# Run smoke test manually
cd templates/web/intro-lab
./tests/smoke.sh

# Check each test step
# 1. Verify container is running
docker ps | grep labkit

# 2. Verify port is listening
nc -z 127.0.0.1 8080

# 3. Verify endpoint responds
curl http://127.0.0.1:8080/health
```

## Runtime Issues

### Lab Not Accessible

**Problem**: Browser can't connect to lab URL

**Solution**:
```bash
# Verify lab is running
labctl status

# Check port binding
docker port labkit-web-intro-lab-app

# Verify with curl
curl -v http://127.0.0.1:8080

# Check firewall
# macOS: System Preferences > Security > Firewall
# Linux: sudo ufw status
```

### Slow Performance

**Problem**: Lab is slow or unresponsive

**Solution**:
```bash
# Check resource usage
docker stats

# Increase resource limits in compose.yaml
deploy:
  resources:
    limits:
      memory: 1G
      cpus: "1.0"

# Check host resources
# macOS: Activity Monitor
# Linux: top, htop
```

### Unexpected Behavior

**Problem**: Lab behaves differently than expected

**Solution**:
```bash
# Check application logs
labctl logs web/intro-lab

# Review lab documentation
cat templates/web/intro-lab/README.md

# Check learning objectives
cat templates/web/intro-lab/lessons/objectives.md

# Rebuild containers
cd templates/web/intro-lab
docker compose down
docker compose up -d --build
```

## Cleanup Issues

### Lab Won't Stop

**Problem**: `labctl stop` fails to stop the lab

**Solution**:
```bash
# Force stop with Docker
docker stop labkit-web-intro-lab-app

# Or use Docker Compose
cd templates/web/intro-lab
docker compose down

# Kill if necessary
docker kill labkit-web-intro-lab-app
```

### Lab Won't Remove

**Problem**: `labctl remove` fails to remove the lab

**Solution**:
```bash
# Force remove with Docker
docker rm -f labkit-web-intro-lab-app

# Remove volumes
docker volume rm $(docker volume ls -q | grep labkit)

# Remove networks
docker network rm $(docker network ls -q | grep labkit)

# Clean everything
docker system prune -a
```

### Orphaned Resources

**Problem**: Resources left behind after lab removal

**Solution**:
```bash
# List all labkit resources
docker ps -a | grep labkit
docker volume ls | grep labkit
docker network ls | grep labkit

# Remove manually
docker rm -f $(docker ps -a -q --filter name=labkit)
docker volume rm $(docker volume ls -q --filter name=labkit)
docker network rm $(docker network ls -q --filter name=labkit)
```

## State File Issues

### State File Corrupted

**Problem**: labctl can't read state file

**Solution**:
```bash
# Check state file
cat ~/.labkit/state.json

# Backup state file
cp ~/.labkit/state.json ~/.labkit/state.json.backup

# Delete state file (will be recreated)
rm ~/.labkit/state.json

# Run labctl again
labctl status
```

### Incorrect State Information

**Problem**: `labctl status` shows incorrect information

**Solution**:
```bash
# Reset state file
rm ~/.labkit/state.json

# Rebuild state from running containers
labctl status  # Will rebuild state
```

## Network Issues

### DNS Resolution Fails

**Problem**: Containers can't resolve hostnames

**Solution**:
```bash
# Check Docker DNS
docker run --rm alpine nslookup google.com

# Use Docker's embedded DNS
# Add to compose.yaml:
dns:
  - 8.8.8.8
  - 8.8.4.4

# Check /etc/resolv.conf on host
cat /etc/resolv.conf
```

### Can't Access External Services

**Problem**: Lab can't access internet/outbound services

**Solution**:
```bash
# Check if outbound is restricted
# In lab.yaml:
runtime:
  network:
    outbound: restricted  # Change to unrestricted if needed

# Check firewall rules
# macOS: System Preferences > Security > Firewall
# Linux: sudo iptables -L
```

## CI/CD Issues

### Validation Workflow Fails

**Problem**: GitHub Actions validation fails

**Solution**:
```bash
# Run validation locally
./scripts/validate-all.sh

# Check workflow logs in GitHub Actions
# Look for specific error messages

# Fix validation errors
# See "Validation Issues" section above
```

### Smoke Test Workflow Fails

**Problem**: GitHub Actions smoke test fails

**Solution**:
```bash
# Run smoke test locally
cd templates/web/intro-lab
./tests/smoke.sh

# Check Docker is running locally
docker ps

# Check test script permissions
chmod +x tests/smoke.sh

# Review workflow logs for specific failures
```

## Getting Help

### Debug Mode

Enable debug output:

```bash
# Set environment variable
export LABKIT_DEBUG=1

# Run labctl
labctl run web/intro-lab
```

### Verbose Output

Get more detailed output:

```bash
# Use verbose flag
labctl --verbose run web/intro-lab

# Or with Docker Compose
docker compose --verbose up
```

### Log Collection

Collect logs for bug reports:

```bash
# Collect labctl logs
labctl logs web/intro-lab > labkit-debug.log

# Collect Docker logs
docker logs labkit-web-intro-lab-app > docker-debug.log

# Collect system info
docker info > system-info.log
docker version > docker-version.log
```

### Report Issues

When reporting issues, include:

1. LabKit version: `labctl --version`
2. Docker version: `docker --version`
3. Docker Compose version: `docker compose version`
4. OS: `uname -a`
5. Error message
6. Steps to reproduce
7. Relevant logs

Report issues at: https://github.com/labkit-project/vulnerable-lab-platform/issues

## Common Error Messages

### "permission denied"

**Cause**: Permission issue with Docker or files

**Solution**: Check user permissions, add to docker group, or use sudo

### "no such file or directory"

**Cause**: File or directory doesn't exist

**Solution**: Verify path, check current directory, use absolute paths

### "connection refused"

**Cause**: Service not running or not accessible

**Solution**: Check if service is running, verify port binding, check firewall

### "timeout"

**Cause**: Operation took too long

**Solution**: Increase timeout, check resource usage, verify network connectivity

### "invalid configuration"

**Cause**: Compose or manifest file has errors

**Solution**: Validate YAML, check syntax, compare with examples

## Prevention

### Regular Maintenance

```bash
# Clean up unused resources
docker system prune -a

# Update Docker images
docker pull <image-name>

# Update labctl
go build -o labctl ./cmd/labctl
```

### Backup State

```bash
# Backup state file
cp ~/.labkit/state.json ~/.labkit/state.json.backup

# Backup templates
tar -czf labkit-templates-backup.tar.gz templates/
```

### Monitoring

```bash
# Monitor resource usage
docker stats

# Monitor disk usage
df -h

# Monitor Docker logs
docker events
```

## Additional Resources

- [Docker Troubleshooting](https://docs.docker.com/config/troubleshooting/)
- [Docker Compose Troubleshooting](https://docs.docker.com/compose/troubleshooting/)
- [LabKit Documentation](https://github.com/labkit-project/vulnerable-lab-platform/docs)
- [LabKit Issues](https://github.com/labkit-project/vulnerable-lab-platform/issues)
