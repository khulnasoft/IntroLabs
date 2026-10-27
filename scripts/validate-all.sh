#!/bin/bash
# Validate all lab templates in the repository

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LABCTL="${PROJECT_ROOT}/labctl"

# Build labctl if it doesn't exist
if [ ! -f "$LABCTL" ]; then
    echo "Building labctl..."
    cd "$PROJECT_ROOT"
    go build -o labctl ./cmd/labctl
fi

echo "Validating all lab templates..."
echo ""

failed=0
total=0

# Find all template directories
for template_dir in "$PROJECT_ROOT"/templates/*/*/; do
    if [ -d "$template_dir" ]; then
        total=$((total + 1))
        template_name=$(basename "$template_dir")
        echo "Validating: $template_name"

        if "$LABCTL" validate "$template_dir"; then
            echo "✓ $template_name passed"
        else
            echo "✗ $template_name failed"
            failed=$((failed + 1))
        fi
        echo ""
    fi
done

echo "Validation complete"
echo "Total: $total, Passed: $((total - failed)), Failed: $failed"

if [ $failed -gt 0 ]; then
    exit 1
fi
