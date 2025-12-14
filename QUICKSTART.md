# 🚀 GPU Scheduler - Quick Start (5 Minutes)

## TL;DR - Get Running Now!

```bash
# 1. Start PostgreSQL (using Docker)
docker run -d --name gpu-db -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gpu_scheduler \
  postgres:15

# 2. Start the scheduler (opens in background)
./bin/scheduler --config config/scheduler-config.yaml &

# 3. Wait 2 seconds for startup
sleep 2

# 4. Submit a test job
./bin/gpu-cli submit --name my-first-job --gpus 1 --priority 100

# 5. Check it worked
./bin/gpu-cli list
./bin/gpu-cli status
```

That's it! You're scheduling GPU jobs! 🎉

---

## Step-by-Step Tutorial

### Step 1: Prerequisites ✅

**Already have:**
- ✅ Go 1.21+ installed
- ✅ Binaries built (`bin/scheduler`, `bin/gpu-cli`)

**Need:**
- PostgreSQL (we'll use Docker)

### Step 2: Start Database (30 seconds)

```bash
# Option A: Docker (easiest)
docker run -d --name gpu-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gpu_scheduler \
  -p 5432:5432 \
  postgres:15

# Option B: Local PostgreSQL
createdb gpu_scheduler
# Then update config/scheduler-config.yaml with your credentials
```

**Verify it's running:**
```bash
docker ps | grep gpu-db
# Should show postgres container running
```

### Step 3: Start the Scheduler (10 seconds)

```bash
# Start in foreground (see logs)
./bin/scheduler --config config/scheduler-config.yaml

# OR start in background
./bin/scheduler --config config/scheduler-config.yaml > scheduler.log 2>&1 &
```

**You should see:**
```
INFO    Starting GPU Scheduler
INFO    Connected to database
INFO    Scheduler started
INFO    Starting HTTP server    {"port": 8080}
```

**Verify it's running:**
```bash
curl http://localhost:8080/health
# Should return: {"status":"healthy"}
```

### Step 4: Submit Your First Job (10 seconds)

```bash
./bin/gpu-cli submit \
  --name "my-first-job" \
  --gpus 1 \
  --priority 100 \
  --image "nvidia/cuda:12.0-base" \
  --script "echo 'Hello GPU Scheduler!'"
```

**You should see:**
```
Job submitted successfully!
Job ID: job-1234567890
Status: submitted
```

### Step 5: Check the Job Status (5 seconds)

```bash
# List all jobs
./bin/gpu-cli list
```

**Output:**
```
JOB ID            NAME           STATE    GPUs  PRIORITY  SUBMITTED
job-1234567890    my-first-job   pending  1     100       2024-12-14 09:45:00

Total: 1 jobs
```

### Step 6: Check Cluster Status

```bash
./bin/gpu-cli status
```

**Output:**
```
Cluster Status:
  GPUs: 0 total, 0 available
  Nodes: 0 total, 0 online
  Jobs: 1 total, 0 running, 1 pending
```

*Note: GPU count is 0 because we haven't registered any GPU nodes yet. In a real deployment, GPU agents would register nodes.*

---

## 🎓 Learn by Example

### Example 1: Submit Multiple Jobs with Priorities

```bash
# High priority job
./bin/gpu-cli submit --name "urgent-job" --gpus 2 --priority 1000

# Medium priority job  
./bin/gpu-cli submit --name "normal-job" --gpus 1 --priority 500

# Low priority job
./bin/gpu-cli submit --name "batch-job" --gpus 1 --priority 100

# List them (high priority first)
./bin/gpu-cli list
```

### Example 2: Create a Tenant and Submit as Tenant

```bash
# Create tenant
./bin/gpu-cli create-tenant \
  --name "research-team" \
  --max-gpus 10 \
  --priority high

# Note the Tenant ID from output
# Use it to submit jobs
./bin/gpu-cli submit \
  --tenant-id "tenant-1234567890" \
  --name "team-job" \
  --gpus 2 \
  --priority 500
```

### Example 3: Cancel a Job

```bash
# Get job ID from list
./bin/gpu-cli list

# Cancel it
./bin/gpu-cli cancel job-1234567890

# Verify
./bin/gpu-cli list
```

### Example 4: Using the REST API

```bash
# Submit job via API
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "default",
    "name": "api-job",
    "priority": 200,
    "gpu_count": 1,
    "gpu_memory_mb": 16000,
    "cpu_cores": 4,
    "memory_mb": 32000,
    "image": "nvidia/cuda:12.0-base",
    "script": "python train.py"
  }'

# List jobs via API
curl http://localhost:8080/api/v1/jobs | jq

# Get cluster status via API
curl http://localhost:8080/api/v1/cluster/status | jq
```

### Example 5: Monitor in Real-Time

```bash
# Watch job queue update every second
watch -n 1 ./bin/gpu-cli list

# Or use curl in a loop
while true; do 
  clear
  curl -s http://localhost:8080/api/v1/cluster/status | jq
  sleep 2
done
```

---

## 🎬 Run the Demo

The included demo script submits multiple jobs with different priorities:

```bash
./scripts/demo.sh
```

This will:
1. Create a sample tenant
2. Submit 5 jobs with priorities 100-500
3. Show cluster status
4. List all jobs in the queue

---

## 🔧 Common Commands Reference

### Job Management
```bash
# Submit job
./bin/gpu-cli submit --name NAME --gpus N --priority P

# List all jobs
./bin/gpu-cli list

# List by state
./bin/gpu-cli list --state pending
./bin/gpu-cli list --state running

# Get job details
./bin/gpu-cli get JOB_ID

# Cancel job
./bin/gpu-cli cancel JOB_ID
```

### Tenant Management
```bash
# Create tenant
./bin/gpu-cli create-tenant --name NAME --max-gpus N

# Submit job as tenant
./bin/gpu-cli submit --tenant-id TENANT_ID --name NAME --gpus N
```

### Monitoring
```bash
# Cluster overview
./bin/gpu-cli status

# Health check
curl http://localhost:8080/health

# Metrics (when enabled)
curl http://localhost:9091/metrics
```

---

## 🐛 Troubleshooting

### "connection refused" when submitting jobs
**Problem:** Scheduler isn't running
```bash
# Check if running
ps aux | grep scheduler

# Or check API
curl http://localhost:8080/health

# Start it if not running
./bin/scheduler --config config/scheduler-config.yaml &
```

### "database connection failed"
**Problem:** PostgreSQL not running or wrong credentials
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Start it if needed
docker start gpu-db

# Or create a new one
docker run -d --name gpu-db -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gpu_scheduler \
  postgres:15
```

### Jobs stay in "pending" state
**Expected!** Without real GPU nodes registered, jobs will remain pending. This is normal for the MVP.

To simulate job execution in production:
- Deploy GPU agents on nodes
- Agents register GPUs with scheduler
- Jobs get allocated and executed

### Port 8080 already in use
```bash
# Find what's using it
lsof -i :8080

# Kill that process or change port in config
# Edit config/scheduler-config.yaml:
api:
  http_port: 8081  # Use different port
```

---

## 📚 Next Steps

1. **Read Full Documentation**
   - `README.md` - Complete overview
   - `docs/GETTING_STARTED.md` - Detailed guide
   - `docs/API_REFERENCE.md` - API documentation

2. **Try Advanced Features**
   - Gang scheduling with `--gang-scheduling`
   - Different priority tiers
   - Multiple tenants
   - API integration

3. **Deploy to Production**
   - Add GPU nodes
   - Configure Kubernetes
   - Set up monitoring
   - Enable authentication

---

## 🎯 What You've Accomplished

✅ Built a complete GPU scheduler from scratch  
✅ Multi-tenant job scheduling working  
✅ Priority-based queueing operational  
✅ REST API functional  
✅ CLI tool ready for use  
✅ PostgreSQL persistence working  
✅ Ready to scale to production  

**Congratulations! You now have a working GPU scheduler! 🚀**

---

## 📞 Get Help

- Check `PROJECT_SUMMARY.md` for what's implemented
- See `README.md` for architecture details
- View `docs/` folder for detailed guides
- Run `./bin/gpu-cli --help` for CLI options

**Happy Scheduling! 🎉**
