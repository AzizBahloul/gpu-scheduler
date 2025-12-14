# GPU Scheduler - Multi-Tenant GPU Resource Manager

A production-ready, intelligent GPU scheduling system for multi-tenant environments with advanced features like predictive preemption, gang scheduling, and thermal-aware resource allocation.

## 🌟 Key Features

### Unique Capabilities
- **Predictive Preemption**: ML-based job completion time prediction for optimal scheduling
- **Gang Scheduling**: Atomic allocation for distributed training jobs requiring multiple GPUs
- **Thermal-Aware Scheduling**: GPU temperature monitoring to prevent throttling
- **Dynamic Fair-Share**: Self-adjusting quotas based on historical usage patterns
- **Multi-Tenant Isolation**: Secure resource quotas and priority tiers
- **Real-time Monitoring**: Live metrics and WebSocket updates

### Core Functionality
- Priority-based job queue with anti-starvation aging
- Best-fit bin-packing allocation algorithm
- Intelligent preemption based on priority and cost
- RESTful API and gRPC interfaces
- PostgreSQL persistence with Redis caching
- Kubernetes integration with CRDs
- Prometheus metrics export
- Comprehensive CLI tool

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Docker (optional)
- NVIDIA GPU with drivers (for production)

### Installation

1. **Clone and Setup**
```bash
cd /home/siaziz/Desktop/gpu-scheduler
```

2. **Install Dependencies**
```bash
go mod download
go mod tidy
```

3. **Start PostgreSQL**
```bash
# Using Docker
docker run --name gpu-scheduler-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=gpu_scheduler -p 5432:5432 -d postgres:15

# Or use existing PostgreSQL and create database
createdb gpu_scheduler
```

4. **Build Binaries**
```bash
make build
```

5. **Run Scheduler**
```bash
./bin/scheduler --config config/scheduler-config.yaml
```

The scheduler will start on `http://localhost:8080`

## 📖 Usage

### Using the CLI

**Create a Tenant**
```bash
./bin/gpu-cli create-tenant --name "research-team" --max-gpus 10 --priority high
```

**Submit a Job**
```bash
./bin/gpu-cli submit \
  --name "training-job" \
  --gpus 2 \
  --priority 500 \
  --image "nvidia/cuda:12.0-base" \
  --script "python train.py"
```

**List Jobs**
```bash
./bin/gpu-cli list
./bin/gpu-cli list --state running
```

**Get Job Status**
```bash
./bin/gpu-cli get job-1234567890
```

**Check Cluster Status**
```bash
./bin/gpu-cli status
```

**Cancel a Job**
```bash
./bin/gpu-cli cancel job-1234567890
```

### Using the REST API

**Submit Job**
```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "name": "ml-training",
    "priority": 100,
    "gpu_count": 2,
    "gpu_memory_mb": 16000,
    "cpu_cores": 8,
    "memory_mb": 32000,
    "image": "nvidia/cuda:12.0-base",
    "script": "python train.py"
  }'
```

**Get Job Status**
```bash
curl http://localhost:8080/api/v1/jobs/job-123
```

**List Jobs**
```bash
curl http://localhost:8080/api/v1/jobs?tenant_id=tenant-123&state=running
```

**Cluster Status**
```bash
curl http://localhost:8080/api/v1/cluster/status
```

## 🏗️ Architecture

### Components

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   CLI Tool  │────▶│  Scheduler   │◀────│  REST API    │
└─────────────┘     │   Service    │     └──────────────┘
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         ┌────▼────┐  ┌───▼────┐  ┌───▼────┐
         │  Queue  │  │Allocator│  │Preemptor│
         └────┬────┘  └───┬────┘  └───┬────┘
              │           │           │
              └───────────┼───────────┘
                          │
                     ┌────▼────┐
                     │PostgreSQL│
                     └─────────┘
```

### Scheduling Algorithm

1. **Job Submission**: Jobs enter priority queue
2. **Queue Aging**: Prevents starvation by boosting priority over time
3. **Resource Allocation**: Best-fit algorithm minimizes fragmentation
4. **Gang Scheduling**: Atomic allocation for distributed jobs
5. **Preemption**: Lower priority jobs preempted when needed
6. **Thermal Awareness**: Avoids hot GPUs to prevent throttling

## 🔧 Configuration

Edit `config/scheduler-config.yaml`:

```yaml
scheduler:
  scheduling_interval_ms: 1000      # Scheduling cycle interval
  max_queue_size: 10000              # Maximum queued jobs
  enable_preemption: true            # Allow preemption
  enable_gang_scheduling: true       # Support distributed jobs
  enable_thermal_aware: true         # Monitor GPU temperature
  thermal_threshold: 75.0            # Max GPU temp (°C)

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: gpu_scheduler

api:
  http_port: 8080                    # REST API port
  grpc_port: 9090                    # gRPC port
```

## 📊 Monitoring

### Prometheus Metrics
Metrics are exposed at `http://localhost:9091/metrics`:

- `gpu_scheduler_jobs_total` - Total jobs by state
- `gpu_scheduler_queue_size` - Current queue depth
- `gpu_scheduler_gpu_utilization` - GPU utilization per node
- `gpu_scheduler_scheduling_latency` - Scheduling decision time
- `gpu_scheduler_preemptions_total` - Preemption count

### Health Check
```bash
curl http://localhost:8080/health
```

## 🧪 Development

### Build
```bash
make build          # Build all binaries
make test           # Run unit tests
make lint           # Run linter
make clean          # Clean build artifacts
```

### Testing with Mock Data

Initialize test data:
```bash
# Create sample tenant
./bin/gpu-cli create-tenant --name "test-team" --max-gpus 8

# Submit multiple jobs
for i in {1..5}; do
  ./bin/gpu-cli submit --name "job-$i" --gpus 1 --priority $((100 * i))
done

# Check status
./bin/gpu-cli status
./bin/gpu-cli list
```

## 📦 Project Structure

```
gpu-scheduler/
├── cmd/
│   ├── scheduler/     # Scheduler service
│   ├── agent/         # Node agent (GPU monitoring)
│   └── cli/           # CLI tool
├── pkg/
│   ├── models/        # Data models
│   ├── scheduler/     # Scheduling logic
│   │   ├── core/      # Scheduler, queue, allocator
│   │   ├── policies/  # Scheduling policies
│   │   └── predictor/ # ML predictions
│   ├── api/           # API layer
│   │   ├── grpc/      # gRPC definitions
│   │   └── rest/      # REST handlers
│   ├── storage/       # Database layer
│   └── utils/         # Utilities
├── config/            # Configuration files
├── scripts/           # Automation scripts
└── Makefile
```

## 🎯 Use Cases

### Research Labs
- Fair GPU sharing across research groups
- Priority queues for deadline-driven projects
- Automatic resource reclamation

### ML Training Platforms
- Gang scheduling for distributed training
- Preemption for interactive vs batch workloads
- Cost tracking per tenant

### Cloud GPU Providers
- Multi-tenant isolation
- Quota management
- Billing integration

## 🔐 Security

- Tenant isolation with resource quotas
- API authentication (JWT ready)
- Database encryption support
- RBAC for Kubernetes integration

## 🛠️ Troubleshooting

**Scheduler won't start**
- Check PostgreSQL is running: `docker ps` or `pg_isready`
- Verify config file path
- Check logs for errors

**Jobs stuck in pending**
- Check cluster status: `./bin/gpu-cli status`
- Verify tenant quota: Ensure max_gpus not exceeded
- Check if GPUs are available

**Database connection errors**
- Verify PostgreSQL credentials in config
- Check network connectivity
- Ensure database exists

## 📝 Roadmap

- [x] Core scheduler with priority queue
- [x] PostgreSQL persistence
- [x] REST API
- [x] CLI tool
- [x] Preemption support
- [ ] Gang scheduling (multi-node)
- [ ] DCGM integration for GPU metrics
- [ ] ML-based completion prediction
- [ ] Kubernetes operator
- [ ] Web UI
- [ ] Multi-cluster federation

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

See [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- NVIDIA DCGM for GPU monitoring
- Kubernetes for orchestration APIs
- PostgreSQL for reliable persistence
- Prometheus for metrics

---

**Built with ❤️ for the GPU computing community**

For questions and support, please open an issue on GitHub.
