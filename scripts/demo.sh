#!/bin/bash

# Demo script showing GPU Scheduler in action

set -e

echo "======================================"
echo "GPU Scheduler - Demo"
echo "======================================"
echo ""

# Check if scheduler is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "⚠️  Scheduler is not running!"
    echo ""
    echo "Please start the scheduler first:"
    echo "  ./bin/scheduler --config config/scheduler-config.yaml"
    echo ""
    echo "Or run in Docker:"
    echo "  docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=gpu_scheduler --name gpu-db postgres:15"
    echo "  ./bin/scheduler --config config/scheduler-config.yaml"
    exit 1
fi

echo "✓ Scheduler is running"
echo ""

CLI="./bin/gpu-cli"

# Create a tenant
echo "1. Creating tenant 'research-team'..."
TENANT_OUTPUT=$($CLI create-tenant --name "research-team" --max-gpus 8 --priority high 2>&1 || true)
echo "$TENANT_OUTPUT"
TENANT_ID=$(echo "$TENANT_OUTPUT" | grep "Tenant ID:" | awk '{print $3}')

if [ -z "$TENANT_ID" ]; then
    echo "Using default tenant ID"
    TENANT_ID="default"
fi

echo ""

# Submit multiple jobs
echo "2. Submitting 5 test jobs with different priorities..."
for i in {1..5}; do
    PRIORITY=$((i * 100))
    echo "  Submitting job-$i (priority: $PRIORITY, GPUs: 1)..."
    $CLI submit \
        --tenant-id "$TENANT_ID" \
        --name "training-job-$i" \
        --gpus 1 \
        --priority $PRIORITY \
        --image "nvidia/cuda:12.0-base" \
        --script "sleep $((i * 10))" > /dev/null
done

echo ""
echo "✓ Jobs submitted"
echo ""

# Show cluster status
echo "3. Cluster Status:"
$CLI status
echo ""

# List all jobs
echo "4. Job Queue:"
$CLI list
echo ""

# Check specific job
echo "5. Checking job details..."
FIRST_JOB=$($CLI list | tail -n 1 | awk '{print $1}')
if [ ! -z "$FIRST_JOB" ]; then
    echo "Details for job: $FIRST_JOB"
    $CLI get "$FIRST_JOB"
fi

echo ""
echo "======================================"
echo "Demo Complete!"
echo "======================================"
echo ""
echo "The scheduler is processing jobs in priority order."
echo "Higher priority jobs (job-5) will be scheduled first."
echo ""
echo "To see real-time updates:"
echo "  watch -n 1 ./bin/gpu-cli list"
echo ""
echo "To submit more jobs:"
echo "  ./bin/gpu-cli submit --name my-job --gpus 2 --priority 1000"
echo ""
echo "To cancel a job:"
echo "  ./bin/gpu-cli cancel <job-id>"
echo ""
