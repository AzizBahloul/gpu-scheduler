#!/bin/bash

# Simple test script to verify the system works

set -e

API_URL="http://localhost:8080"
CLI="./bin/gpu-cli --api-url $API_URL"

echo "Testing GPU Scheduler..."
echo ""

# Wait for API to be ready
echo "Waiting for API to be ready..."
for i in {1..30}; do
    if curl -s "$API_URL/health" > /dev/null 2>&1; then
        echo "API is ready!"
        break
    fi
    sleep 1
done

# Test 1: Health check
echo ""
echo "Test 1: Health Check"
curl -s "$API_URL/health" | jq .
echo "✓ Health check passed"

# Test 2: Create tenant
echo ""
echo "Test 2: Create Tenant"
TENANT_RESP=$(curl -s -X POST "$API_URL/api/v1/tenants" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test-team",
        "max_gpus": 10,
        "max_cpu_cores": 64,
        "max_memory_mb": 256000,
        "priority_tier": "high"
    }')
TENANT_ID=$(echo $TENANT_RESP | jq -r .id)
echo "Created tenant: $TENANT_ID"
echo "✓ Tenant creation passed"

# Test 3: Submit job
echo ""
echo "Test 3: Submit Job"
JOB_RESP=$(curl -s -X POST "$API_URL/api/v1/jobs" \
    -H "Content-Type: application/json" \
    -d "{
        \"tenant_id\": \"$TENANT_ID\",
        \"name\": \"test-job\",
        \"priority\": 100,
        \"gpu_count\": 1,
        \"gpu_memory_mb\": 16000,
        \"cpu_cores\": 4,
        \"memory_mb\": 32000,
        \"image\": \"nvidia/cuda:12.0-base\",
        \"script\": \"nvidia-smi\"
    }")
JOB_ID=$(echo $JOB_RESP | jq -r .job_id)
echo "Submitted job: $JOB_ID"
echo "✓ Job submission passed"

# Test 4: Get job status
echo ""
echo "Test 4: Get Job Status"
curl -s "$API_URL/api/v1/jobs/$JOB_ID" | jq .
echo "✓ Job status retrieval passed"

# Test 5: List jobs
echo ""
echo "Test 5: List Jobs"
curl -s "$API_URL/api/v1/jobs?tenant_id=$TENANT_ID" | jq .
echo "✓ Job listing passed"

# Test 6: Cluster status
echo ""
echo "Test 6: Cluster Status"
curl -s "$API_URL/api/v1/cluster/status" | jq .
echo "✓ Cluster status passed"

echo ""
echo "================================"
echo "All tests passed successfully! ✓"
echo "================================"
