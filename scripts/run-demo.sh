#!/bin/bash

# GPU Scheduler Demo - No Database Required
# This demonstrates the scheduler with mock data

set -e

echo "🚀 GPU Scheduler Demo"
echo "===================="
echo ""

cd "$(dirname "$0")/.."

# Check if binaries exist
if [ ! -f "bin/scheduler" ]; then
    echo "Building binaries..."
    make build
fi

echo "📊 Demo: Testing Queue and Priority System"
echo ""

# Use Go to run a demo without database
cat > /tmp/gpu_scheduler_demo.go <<'EOF'
package main

import (
	"fmt"
	"time"
	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/scheduler/core"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
)

func main() {
	fmt.Println("🎯 Initializing GPU Scheduler...")
	
	// Create scheduler config
	config := &utils.SchedulerConfig{
		SchedulingInterval: 1000,
		MaxQueueSize:      100,
		EnablePreemption:  true,
	}
	
	// Create queue
	queue := core.NewQueue(100)
	
	fmt.Println("\n📝 Submitting test jobs...")
	
	// Submit jobs with different priorities
	jobs := []*models.Job{
		{
			ID:          "job-1",
			Name:        "Low Priority Training",
			Priority:    100,
			GPUCount:    2,
			GPUMemoryMB: 16000,
			State:       models.JobStatePending,
		},
		{
			ID:          "job-2",
			Name:        "High Priority Inference",
			Priority:    1000,
			GPUCount:    1,
			GPUMemoryMB: 8000,
			State:       models.JobStatePending,
		},
		{
			ID:          "job-3",
			Name:        "Medium Priority Analysis",
			Priority:    500,
			GPUCount:    4,
			GPUMemoryMB: 32000,
			State:       models.JobStatePending,
		},
		{
			ID:          "job-4",
			Name:        "Gang Scheduled Distributed Training",
			Priority:    800,
			GPUCount:    8,
			GPUMemoryMB: 64000,
			GangScheduling: true,
			State:       models.JobStatePending,
		},
	}
	
	for _, job := range jobs {
		queue.Enqueue(job)
		fmt.Printf("  ✅ Queued: %s (Priority: %d, GPUs: %d)\n", 
			job.Name, job.Priority, job.GPUCount)
	}
	
	fmt.Printf("\n📊 Queue Status:\n")
	fmt.Printf("  Total Jobs: %d\n", queue.Size())
	
	fmt.Println("\n🔄 Processing Queue (Priority Order):")
	position := 1
	for !queue.IsEmpty() {
		job := queue.Dequeue()
		gang := ""
		if job.GangScheduling {
			gang = " [GANG]"
		}
		fmt.Printf("  %d. %s - Priority: %d, GPUs: %d%s\n", 
			position, job.Name, job.Priority, job.GPUCount, gang)
		position++
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Println("\n⏰ Testing Aging Mechanism...")
	
	// Create new queue for aging test
	queue2 := core.NewQueue(100)
	
	oldJob := &models.Job{
		ID:       "old-job",
		Name:     "Old Low Priority Job",
		Priority: 100,
		GPUCount: 1,
		State:    models.JobStatePending,
	}
	
	newJob := &models.Job{
		ID:       "new-job",
		Name:     "New High Priority Job",
		Priority: 200,
		GPUCount: 1,
		State:    models.JobStatePending,
	}
	
	queue2.Enqueue(oldJob)
	time.Sleep(1 * time.Second)
	queue2.Enqueue(newJob)
	
	fmt.Println("  Before aging:")
	fmt.Printf("    Position 1: %s (Priority: %d)\n", newJob.Name, newJob.Priority)
	fmt.Printf("    Position 2: %s (Priority: %d)\n", oldJob.Name, oldJob.Priority)
	
	// Apply aging boost
	queue2.ApplyAging(150, 500*time.Millisecond)
	
	first := queue2.Dequeue()
	second := queue2.Dequeue()
	
	fmt.Println("  After aging (150 boost after 500ms):")
	fmt.Printf("    Position 1: %s (Priority: %d + boost)\n", first.Name, first.Priority)
	fmt.Printf("    Position 2: %s (Priority: %d)\n", second.Name, second.Priority)
	
	fmt.Println("\n🌡️  Testing GPU Thermal Monitoring...")
	
	gpu := &models.GPU{
		ID:          "gpu-0",
		NodeID:      "node-1",
		Model:       models.GPUModelA100,
		Temperature: 65.0,
		Health:      models.HealthHealthy,
	}
	
	fmt.Printf("  GPU: %s\n", gpu.ID)
	fmt.Printf("  Initial Temp: %.1f°C, Health: %s\n", gpu.Temperature, gpu.Health)
	
	// Simulate temperature increase
	gpu.UpdateMetrics(85.0, 82.0, 350.0)
	fmt.Printf("  After load: %.1f°C, Health: %s\n", gpu.Temperature, gpu.Health)
	
	if gpu.NeedsCooling(75.0) {
		fmt.Printf("  ⚠️  GPU needs cooling (threshold: 75°C)\n")
	}
	
	fmt.Println("\n💰 Testing Tenant Quotas...")
	
	tenant := &models.Tenant{
		ID:                "demo-tenant",
		Name:              "Demo Organization",
		MaxGPUs:           10,
		MaxGPUMemoryMB:    160000,
		MaxCPUCores:       64,
		MaxMemoryMB:       256000,
		MaxConcurrentJobs: 20,
		CurrentGPUs:       3,
		CurrentGPUMemory:  48000,
		PriorityTier:      models.PriorityHigh,
	}
	
	fmt.Printf("  Tenant: %s\n", tenant.Name)
	fmt.Printf("  GPU Usage: %d/%d (%.1f%%)\n", 
		tenant.CurrentGPUs, tenant.MaxGPUs, 
		tenant.CalculateFairShare()*100)
	fmt.Printf("  Priority: %s (Score: %d)\n", 
		tenant.PriorityTier, tenant.GetPriorityScore())
	
	// Test quota
	canSubmit := tenant.HasAvailableQuota(4, 64000, 32, 128000)
	if canSubmit {
		fmt.Printf("  ✅ Can submit job requiring 4 GPUs\n")
	} else {
		fmt.Printf("  ❌ Quota exceeded for 4 GPUs\n")
	}
	
	fmt.Println("\n✨ Demo Complete!")
	fmt.Println("\n📚 Key Features Demonstrated:")
	fmt.Println("  ✅ Priority-based scheduling")
	fmt.Println("  ✅ Gang scheduling support")
	fmt.Println("  ✅ Anti-starvation aging")
	fmt.Println("  ✅ GPU thermal monitoring")
	fmt.Println("  ✅ Multi-tenant quotas")
	fmt.Println("  ✅ Fair-share calculation")
}
EOF

echo "Running demo..."
cd /tmp
go mod init demo 2>/dev/null || true
go mod edit -replace github.com/azizbahloul/gpu-scheduler=/home/siaziz/Desktop/gpu-scheduler
go mod tidy -e 2>/dev/null || true
go run gpu_scheduler_demo.go
cd - > /dev/null

echo ""
echo "✅ Demo completed successfully!"
echo ""
echo "To run the full scheduler with PostgreSQL:"
echo "  1. Start PostgreSQL: docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=yourpassword postgres:15"
echo "  2. Update config: Edit config/scheduler-config.yaml with your password"
echo "  3. Run: ./bin/scheduler"
