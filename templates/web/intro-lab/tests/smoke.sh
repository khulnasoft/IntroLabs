#!/bin/bash
# Smoke test for the Introduction to Web Security lab
# This script verifies that the lab environment is running correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Lab configuration
LAB_ID="${LAB_ID:-web-intro-lab}"
ENDPOINT="${ENDPOINT:-http://127.0.0.1:8081}"
TIMEOUT=60

echo "Running smoke test for lab: ${LAB_ID}"
echo "Endpoint: ${ENDPOINT}"
echo ""

# Function to check if a service is responding
check_endpoint() {
    local endpoint=$1
    local expected_status=${2:-200}
    local timeout=$3

    echo -n "Checking ${endpoint}... "

    local start_time=$(date +%s)
    local end_time=$((start_time + timeout))

    while [ $(date +%s) -lt $end_time ]; do
        local status=$(curl -s -o /dev/null -w "%{http_code}" "${endpoint}" 2>/dev/null || echo "000")

        if [ "$status" = "$expected_status" ]; then
            echo -e "${GREEN}OK${NC} (HTTP ${status})"
            return 0
        fi

        sleep 2
    done

    echo -e "${RED}FAILED${NC} (Expected ${expected_status}, got ${status})"
    return 1
}

# Function to check if a container is running
check_container() {
    local container_name=$1

    echo -n "Checking container ${container_name}... "

    if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAILED${NC} (Container not running)"
        return 1
    fi
}

# Function to check if a port is listening
check_port() {
    local port=$1
    local host=${2:-127.0.0.1}

    echo -n "Checking port ${port} on ${host}... "

    if lsof -i :${port} >/dev/null 2>&1 || nc -z ${host} ${port} 2>/dev/null; then
        echo -e "${GREEN}OK${NC}"
        return 0
    else
        echo -e "${RED}FAILED${NC} (Port not listening)"
        return 1
    fi
}

# Run checks
failed=0

# Check Docker is available
echo -n "Checking Docker... "
if command -v docker &> /dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC} (Docker not found)"
    exit 1
fi

# Check Docker Compose is available
echo -n "Checking Docker Compose... "
if docker compose version &> /dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC} (Docker Compose not found)"
    exit 1
fi

echo ""

# Check containers are running
check_container "labkit-${LAB_ID}-app" || failed=$((failed + 1))

# Check port is listening
check_port 8080 127.0.0.1 || failed=$((failed + 1))

# Check endpoint is responding
check_endpoint "${ENDPOINT}" 200 ${TIMEOUT} || failed=$((failed + 1))

# Check health endpoint
check_endpoint "${ENDPOINT}/health" 200 10 || failed=$((failed + 1))

# Check profile endpoint exists
check_endpoint "${ENDPOINT}/profile?id=1" 200 10 || failed=$((failed + 1))

echo ""

# Summary
if [ $failed -eq 0 ]; then
    echo -e "${GREEN}All smoke tests passed!${NC}"
    echo ""
    echo "The lab is ready for use."
    echo "Access it at: ${ENDPOINT}"
    exit 0
else
    echo -e "${RED}${failed} smoke test(s) failed${NC}"
    exit 1
fi
