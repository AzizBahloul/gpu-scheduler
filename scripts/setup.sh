#!/bin/bash

set -e

echo "======================================"
echo "GPU Scheduler - Quick Setup"
echo "======================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Go installation
echo -n "Checking Go installation... "
if ! command -v go &> /dev/null; then
    echo -e "${RED}FAILED${NC}"
    echo "Go is not installed. Please install Go 1.21+ from https://golang.org/"
    exit 1
fi
echo -e "${GREEN}OK${NC} ($(go version))"

# Check PostgreSQL
echo -n "Checking PostgreSQL... "
if command -v psql &> /dev/null; then
    echo -e "${GREEN}OK${NC}"
elif command -v docker &> /dev/null; then
    echo -e "${YELLOW}Not found, will use Docker${NC}"
    
    echo "Starting PostgreSQL in Docker..."
    docker run --name gpu-scheduler-db \
        -e POSTGRES_PASSWORD=postgres \
        -e POSTGRES_DB=gpu_scheduler \
        -p 5432:5432 \
        -d postgres:15 2>/dev/null || echo "Container already exists"
    
    echo "Waiting for PostgreSQL to be ready..."
    sleep 5
else
    echo -e "${RED}FAILED${NC}"
    echo "PostgreSQL not found and Docker not available"
    echo "Please install PostgreSQL or Docker"
    exit 1
fi

# Download dependencies
echo ""
echo "Downloading Go dependencies..."
go mod download
go mod tidy

# Build binaries
echo ""
echo "Building binaries..."
mkdir -p bin

echo "  - Building scheduler..."
go build -o bin/scheduler ./cmd/scheduler

echo "  - Building CLI..."
go build -o bin/gpu-cli ./cmd/cli

echo ""
echo -e "${GREEN}Build successful!${NC}"
echo ""
echo "======================================"
echo "Setup Complete!"
echo "======================================"
echo ""
echo "Next steps:"
echo ""
echo "1. Start the scheduler:"
echo "   ./bin/scheduler --config config/scheduler-config.yaml"
echo ""
echo "2. In another terminal, create a tenant:"
echo "   ./bin/gpu-cli create-tenant --name my-team --max-gpus 10"
echo ""
echo "3. Submit a test job:"
echo "   ./bin/gpu-cli submit --name test-job --gpus 1"
echo ""
echo "4. Check cluster status:"
echo "   ./bin/gpu-cli status"
echo ""
echo "API will be available at: http://localhost:8080"
echo "Health check: curl http://localhost:8080/health"
echo ""
