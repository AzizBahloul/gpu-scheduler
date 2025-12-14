# GPU Scheduler - Getting Started Guide

## Prerequisites

### Required
- **Go 1.21+**: [Download here](https://golang.org/dl/)
- **PostgreSQL 15+**: Either installed locally or via Docker

### Optional
- Docker (for running PostgreSQL easily)
- jq (for JSON formatting in tests)

## Step 1: Clone and Build

```bash
cd /home/siaziz/Desktop/gpu-scheduler

# Download dependencies
go mod download
go mod tidy

# Build binaries
make build
```

This creates three binaries:
- `bin/scheduler` - Main scheduling service
- `bin/gpu-cli` - Command-line interface
- `bin/agent` - GPU agent (placeholder)

## Step 2: Start PostgreSQL

### Option A: Using Docker (Recommended)
```bash
docker run --name gpu-scheduler-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gpu_scheduler \
  -p 5432:5432 \
  -d postgres:15
```

### Option B: Using Local PostgreSQL
```bash
createdb gpu_scheduler
```

Update `config/scheduler-config.yaml` with your PostgreSQL credentials if different.

## Step 3: Start the Scheduler

```bash
./bin/scheduler --config config/scheduler-config.yaml
```

You should see:
```
INFO    Starting GPU Scheduler
INFO    Connected to database
INFO    Scheduler started
INFO    Starting HTTP server    {"port": 8080}
```

The scheduler is now running and ready to accept jobs!

## Step 4: Verify Installation

In a new terminal:

```bash
# Health check
curl http://localhost:8080/health

# Should return: {"status":"healthy"}
```

## Step 5: Create Your First Tenant

```bash
./bin/gpu-cli create-tenant \
  --name "my-team" \
  --max-gpus 10 \
  --priority high
```

## Step 6: Submit a Job

```bash
./bin/gpu-cli submit \
  --name "test-job" \
  --gpus 1 \
  --priority 100 \
  --image "nvidia/cuda:12.0-base" \
  --script "echo 'Hello from GPU Scheduler'"
```

## Step 7: Monitor Jobs

```bash
# List all jobs
./bin/gpu-cli list

# Get cluster status
./bin/gpu-cli status

# Check specific job
./bin/gpu-cli get <job-id>
```

## Quick Demo

Run the demo script to see the scheduler in action:

```bash
./scripts/demo.sh
```

This will:
1. Create a sample tenant
2. Submit 5 jobs with different priorities
3. Show cluster status
4. List the job queue

## Common Commands

### Job Management
```bash
# Submit job
./bin/gpu-cli submit --name my-job --gpus 2 --priority 500

# List jobs by state
./bin/gpu-cli list --state running
./bin/gpu-cli list --state pending

# Cancel job
./bin/gpu-cli cancel job-123

# Get job details
./bin/gpu-cli get job-123
```

### Cluster Monitoring
```bash
# Cluster status
./bin/gpu-cli status

# Watch jobs in real-time
watch -n 1 ./bin/gpu-cli list
```

### Using REST API
```bash
# Submit job via API
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "name": "api-job",
    "priority": 200,
    "gpu_count": 1,
    "gpu_memory_mb": 16000,
    "cpu_cores": 4,
    "memory_mb": 32000,
    "image": "nvidia/cuda:12.0-base",
    "script": "python train.py"
  }'

# List jobs
curl http://localhost:8080/api/v1/jobs

# Cluster status
curl http://localhost:8080/api/v1/cluster/status
```

## Configuration

Edit `config/scheduler-config.yaml` to customize:

```yaml
scheduler:
  scheduling_interval_ms: 1000    # How often to schedule (ms)
  max_queue_size: 10000            # Max pending jobs
  enable_preemption: true          # Allow preemption
  thermal_threshold: 75.0          # GPU temp limit (°C)

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: gpu_scheduler

api:
  http_port: 8080                  # REST API port
  grpc_port: 9090                  # gRPC port
```

## Troubleshooting

### "Failed to connect to database"
- Ensure PostgreSQL is running: `docker ps` or `pg_isready`
- Check credentials in config file
- Verify port 5432 is accessible

### "Scheduler is not running"
- Check if another process is using port 8080
- View logs for errors
- Verify config file path is correct

### Jobs stuck in pending
- Check cluster status: `./bin/gpu-cli status`
- Verify GPUs are available
- Check tenant quota limits

## Next Steps

1. **Add GPU Nodes**: Register nodes with GPU resources
2. **Create Tenants**: Set up teams with quotas
3. **Submit Jobs**: Run your ML training workloads
4. **Monitor**: Use Prometheus metrics at `:9091/metrics`
5. **Scale**: Deploy with Kubernetes

## Architecture Overview

```
┌──────────────┐
│   CLI Tool   │
└──────┬───────┘
       │
       ▼
┌──────────────┐     ┌──────────────┐
│  REST API    │────▶│  Scheduler   │
│  :8080       │     │   Service    │
└──────────────┘     └──────┬───────┘
                            │
                     ┌──────┴───────┐
                     │  PostgreSQL  │
                     │  :5432       │
                     └──────────────┘
```

## Resources

- **README.md**: Full project documentation
- **Makefile**: Build commands
- **config/**: Configuration files
- **scripts/**: Automation scripts

## Support

For issues or questions:
1. Check troubleshooting section
2. Review logs in scheduler output
3. Open an issue on GitHub

Happy Scheduling! 🚀
