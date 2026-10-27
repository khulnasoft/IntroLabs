#!/bin/bash
# Build release binaries for labctl

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="${1:-dev}"

cd "$PROJECT_ROOT"

echo "Building labctl release binaries..."
echo "Version: $VERSION"
echo ""

# Build for multiple platforms
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o labctl-linux-amd64 ./cmd/labctl

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o labctl-darwin-amd64 ./cmd/labctl

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$VERSION" -o labctl-darwin-arm64 ./cmd/labctl

echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o labctl-windows-amd64.exe ./cmd/labctl

echo ""
echo "Creating release archives..."

# Create archives
tar -czf labctl-linux-amd64.tar.gz labctl-linux-amd64
tar -czf labctl-darwin-amd64.tar.gz labctl-darwin-amd64
tar -czf labctl-darwin-arm64.tar.gz labctl-darwin-arm64
zip labctl-windows-amd64.zip labctl-windows-amd64.exe

# Calculate checksums
echo "Calculating checksums..."
sha256sum labctl-*.tar.gz labctl-*.zip > checksums.txt

echo ""
echo "Build complete!"
echo "Artifacts:"
ls -lh labctl-*.tar.gz labctl-*.zip checksums.txt

# Cleanup binaries
rm labctl-linux-amd64 labctl-darwin-amd64 labctl-darwin-arm64 labctl-windows-amd64.exe
