# API Reference - GPU Scheduler

Base URL: `http://localhost:8080/api/v1`

## Authentication
Currently no authentication required. JWT authentication ready for production.

## Common Headers
```
Content-Type: application/json
```

---

## Jobs

### Submit Job
Create and submit a new job to the scheduler.

**Endpoint:** `POST /jobs`

**Request Body:**
```json
{
  "tenant_id": "string",
  "name": "string",
  "priority": 100,
  "gpu_count": 2,
  "gpu_memory_mb": 16000,
  "cpu_cores": 8,
  "memory_mb": 32000,
  "script": "string",
  "environment": {
    "KEY": "value"
  },
  "image": "nvidia/cuda:12.0-base",
  "command": ["python"],
  "args": ["train.py"],
  "gang_scheduling": false,
  "max_runtime_minutes": 120
}
```

**Response:** `201 Created`
```json
{
  "job_id": "job-1234567890",
  "status": "submitted",
  "message": "Job submitted successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-123",
    "name": "training-job",
    "priority": 500,
    "gpu_count": 2,
    "gpu_memory_mb": 16000,
    "cpu_cores": 8,
    "memory_mb": 32000,
    "image": "nvidia/cuda:12.0-base",
    "script": "python train.py"
  }'
```

---

### Get Job Status
Retrieve detailed status of a specific job.

**Endpoint:** `GET /jobs/{jobID}`

**Response:** `200 OK`
```json
{
  "job_id": "job-1234567890",
  "state": "running",
  "message": "",
  "allocated_gpus": ["gpu-1", "gpu-2"],
  "node_name": "node-1",
  "queue_position": 0,
  "estimated_wait": "0s"
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/jobs/job-1234567890
```

---

### List Jobs
List all jobs with optional filtering.

**Endpoint:** `GET /jobs`

**Query Parameters:**
- `tenant_id` (optional): Filter by tenant
- `state` (optional): Filter by state (pending, running, completed, failed)
- `limit` (optional): Max results (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response:** `200 OK`
```json
{
  "jobs": [
    {
      "id": "job-1234567890",
      "tenant_id": "tenant-123",
      "name": "training-job",
      "state": "running",
      "priority": 500,
      "gpu_count": 2,
      "submitted_at": "2024-01-15T10:30:00Z",
      "started_at": "2024-01-15T10:31:00Z"
    }
  ],
  "total": 1
}
```

**Examples:**
```bash
# All jobs
curl http://localhost:8080/api/v1/jobs

# Jobs by tenant
curl http://localhost:8080/api/v1/jobs?tenant_id=tenant-123

# Running jobs only
curl http://localhost:8080/api/v1/jobs?state=running

# With pagination
curl http://localhost:8080/api/v1/jobs?limit=10&offset=20
```

---

### Cancel Job
Cancel a pending or running job.

**Endpoint:** `DELETE /jobs/{jobID}`

**Response:** `200 OK`
```json
{
  "message": "Job cancelled successfully"
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/jobs/job-1234567890
```

---

## Tenants

### Create Tenant
Create a new tenant with resource quotas.

**Endpoint:** `POST /tenants`

**Request Body:**
```json
{
  "name": "string",
  "email": "string",
  "organization": "string",
  "max_gpus": 10,
  "max_gpu_memory_mb": 160000,
  "max_cpu_cores": 64,
  "max_memory_mb": 256000,
  "max_concurrent_jobs": 20,
  "priority_tier": "high",
  "allow_preemption": true,
  "can_preempt_others": false
}
```

**Response:** `201 Created`
```json
{
  "id": "tenant-1234567890",
  "name": "research-team",
  "max_gpus": 10,
  "current_gpus": 0,
  "priority_tier": "high",
  "active": true,
  "created_at": "2024-01-15T10:00:00Z"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "research-team",
    "max_gpus": 10,
    "max_cpu_cores": 64,
    "max_memory_mb": 256000,
    "priority_tier": "high"
  }'
```

---

## Cluster

### Get Cluster Status
Get overall cluster resource status.

**Endpoint:** `GET /cluster/status`

**Response:** `200 OK`
```json
{
  "total_gpus": 32,
  "available_gpus": 20,
  "total_nodes": 4,
  "online_nodes": 4,
  "total_jobs": 15,
  "pending_jobs": 5,
  "running_jobs": 10
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/cluster/status
```

---

## Health

### Health Check
Check if the API is healthy.

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "healthy"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

---

## Job States

Jobs progress through the following states:

- `pending`: Job is queued, waiting for resources
- `running`: Job is currently executing
- `completed`: Job finished successfully
- `failed`: Job encountered an error
- `preempted`: Job was preempted by higher priority job
- `cancelled`: Job was manually cancelled

## Priority Tiers

Tenants can have the following priority tiers:

- `low`: Priority 100
- `medium`: Priority 500
- `high`: Priority 1000
- `critical`: Priority 5000

## Error Responses

All endpoints may return error responses:

**400 Bad Request**
```json
{
  "error": "Invalid request body"
}
```

**404 Not Found**
```json
{
  "error": "Job not found"
}
```

**403 Forbidden**
```json
{
  "error": "Tenant quota exceeded"
}
```

**500 Internal Server Error**
```json
{
  "error": "Failed to process request"
}
```

## Rate Limiting

Default rate limit: 100 requests/second per client

## WebSocket Support

Coming soon for real-time job status updates.

## Examples with Different Languages

### Python
```python
import requests

# Submit job
response = requests.post('http://localhost:8080/api/v1/jobs', json={
    'tenant_id': 'tenant-123',
    'name': 'python-job',
    'priority': 100,
    'gpu_count': 1,
    'gpu_memory_mb': 16000,
    'cpu_cores': 4,
    'memory_mb': 32000,
    'image': 'nvidia/cuda:12.0-base',
    'script': 'python train.py'
})

job_id = response.json()['job_id']
print(f'Job submitted: {job_id}')

# Get status
status = requests.get(f'http://localhost:8080/api/v1/jobs/{job_id}')
print(status.json())
```

### Node.js
```javascript
const axios = require('axios');

// Submit job
const submitJob = async () => {
  const response = await axios.post('http://localhost:8080/api/v1/jobs', {
    tenant_id: 'tenant-123',
    name: 'nodejs-job',
    priority: 100,
    gpu_count: 1,
    gpu_memory_mb: 16000,
    cpu_cores: 4,
    memory_mb: 32000,
    image: 'nvidia/cuda:12.0-base',
    script: 'python train.py'
  });
  
  console.log('Job submitted:', response.data.job_id);
};
```

### Go
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func submitJob() {
    job := map[string]interface{}{
        "tenant_id":    "tenant-123",
        "name":         "go-job",
        "priority":     100,
        "gpu_count":    1,
        "gpu_memory_mb": 16000,
        "cpu_cores":    4,
        "memory_mb":    32000,
        "image":        "nvidia/cuda:12.0-base",
        "script":       "python train.py",
    }
    
    body, _ := json.Marshal(job)
    resp, _ := http.Post("http://localhost:8080/api/v1/jobs",
        "application/json", bytes.NewBuffer(body))
    defer resp.Body.Close()
}
```

---

For more details, see the [Getting Started Guide](GETTING_STARTED.md).
