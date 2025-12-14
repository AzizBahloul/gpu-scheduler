#!/bin/bash
# Quick Start - GPU Scheduler with Docker PostgreSQL

set -e

echo "🚀 Starting GPU Scheduler"
echo "========================"
echo ""

# Stop any existing container
echo "🧹 Cleaning up old containers..."
docker rm -f gpu-scheduler-db 2>/dev/null || true

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
docker run -d \
  --name gpu-scheduler-db \
  -e POSTGRES_PASSWORD=gpu123 \
  -e POSTGRES_DB=gpu_scheduler \
  -p 5432:5432 \
  postgres:15

echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 5

# Test connection
until docker exec gpu-scheduler-db pg_isready -U postgres > /dev/null 2>&1; do
  echo "   Still waiting..."
  sleep 2
done

echo "✅ PostgreSQL is ready!"
echo ""

# Update config
echo "📝 Updating configuration..."
cat > config/scheduler-config.yaml <<EOF
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

scheduler:
  scheduling_interval_ms: 1000
  max_queue_size: 10000
  enable_preemption: true
  preemption_grace_period_s: 300
  aging_boost: 10
  aging_threshold_minutes: 5
  thermal_threshold: 75.0

database:
  host: localhost
  port: 5432
  user: postgres
  password: gpu123
  database: gpu_scheduler
  ssl_mode: disable
  max_connections: 25
  max_idle_connections: 5
  connection_lifetime_minutes: 30

redis:
  host: localhost
  port: 6379
  password: ""
  database: 0
  enabled: false

api:
  rate_limit: 1000
  enable_cors: true

logging:
  level: info
  format: json
  output: stdout
EOF

echo "✅ Configuration updated"
echo ""

# Start scheduler
echo "🎯 Starting GPU Scheduler..."
echo ""
./bin/scheduler

echo "To stop: docker stop gpu-scheduler-db"
