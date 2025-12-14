#!/bin/bash

# GPU Scheduler Quick Start Demo
# This script demonstrates the GPU scheduler without requiring PostgreSQL

echo "🚀 GPU Scheduler Demo"
echo "===================="
echo ""

# Check if binaries exist
if [ ! -f "./bin/scheduler" ]; then
    echo "❌ Scheduler binary not found. Building..."
    make build || exit 1
fi

echo "✅ Binaries ready:"
ls -lh bin/

echo ""
echo "📋 Testing without database (dry-run mode)..."
echo ""

# Test 1: Show help
echo "1️⃣  Scheduler Help:"
echo "-------------------"
./bin/scheduler --help 2>&1 | head -10 || true

echo ""
echo "2️⃣  CLI Tool Help:"
echo "------------------"
./bin/gpu-cli --help 2>&1 | head -15 || true

echo ""
echo "3️⃣  Running Tests:"
echo "------------------"
go test -v ./pkg/models/... -run TestJobIsTerminal 2>&1 | grep -E "(RUN|PASS|FAIL)"

echo ""
echo "4️⃣  Benchmark Performance:"
echo "-------------------------"
go test -bench=BenchmarkTenantQuotaCheck -benchmem ./tests/benchmarks/... 2>&1 | grep -E "Benchmark|ns/op"

echo ""
echo "📊 Test Coverage:"
echo "-----------------"
go test -cover ./pkg/models/... 2>&1 | grep coverage

echo ""
echo "✨ Demo Complete!"
echo ""
echo "To run with PostgreSQL:"
echo "  1. Start PostgreSQL: docker run -d -p 5433:5432 -e POSTGRES_PASSWORD=postgres postgres:15"
echo "  2. Update config/scheduler-config.yaml to use port 5433"
echo "  3. Run: ./bin/scheduler"
echo ""
echo "Or run the full demo:"
echo "  ./scripts/demo.sh"
