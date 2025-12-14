package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/scheduler/core"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
)

type MockRepository struct{}

func (m *MockRepository) CreateJob(ctx context.Context, job *models.Job) error { return nil }
func (m *MockRepository) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	return &models.Job{ID: jobID}, nil
}
func (m *MockRepository) UpdateJob(ctx context.Context, job *models.Job) error { return nil }
func (m *MockRepository) DeleteJob(ctx context.Context, jobID string) error    { return nil }
func (m *MockRepository) ListJobs(ctx context.Context, limit, offset int) ([]*models.Job, error) {
	return []*models.Job{}, nil
}
func (m *MockRepository) ListJobsByTenant(ctx context.Context, tenantID string) ([]*models.Job, error) {
	return []*models.Job{}, nil
}
func (m *MockRepository) ListJobsByState(ctx context.Context, state models.JobState) ([]*models.Job, error) {
	return []*models.Job{}, nil
}
func (m *MockRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) error { return nil }
func (m *MockRepository) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	return &models.Tenant{ID: tenantID, MaxGPUs: 10, MaxConcurrentJobs: 100}, nil
}
func (m *MockRepository) UpdateTenant(ctx context.Context, tenant *models.Tenant) error { return nil }
func (m *MockRepository) DeleteTenant(ctx context.Context, tenantID string) error       { return nil }
func (m *MockRepository) ListTenants(ctx context.Context) ([]*models.Tenant, error) {
	return []*models.Tenant{}, nil
}
func (m *MockRepository) CreateGPU(ctx context.Context, gpu *models.GPU) error                 { return nil }
func (m *MockRepository) GetGPU(ctx context.Context, gpuID string) (*models.GPU, error)        { return nil, nil }
func (m *MockRepository) UpdateGPU(ctx context.Context, gpu *models.GPU) error                 { return nil }
func (m *MockRepository) DeleteGPU(ctx context.Context, gpuID string) error                    { return nil }
func (m *MockRepository) ListGPUs(ctx context.Context) ([]*models.GPU, error)                  { return []*models.GPU{}, nil }
func (m *MockRepository) ListGPUsByNode(ctx context.Context, nodeID string) ([]*models.GPU, error) {
	return []*models.GPU{}, nil
}
func (m *MockRepository) ListAvailableGPUs(ctx context.Context) ([]*models.GPU, error) {
	return []*models.GPU{}, nil
}
func (m *MockRepository) CreateNode(ctx context.Context, node *models.Node) error { return nil }
func (m *MockRepository) GetNode(ctx context.Context, nodeID string) (*models.Node, error) {
	return nil, nil
}
func (m *MockRepository) UpdateNode(ctx context.Context, node *models.Node) error     { return nil }
func (m *MockRepository) DeleteNode(ctx context.Context, nodeID string) error         { return nil }
func (m *MockRepository) ListNodes(ctx context.Context) ([]*models.Node, error)       { return []*models.Node{}, nil }
func (m *MockRepository) CreateAllocation(ctx context.Context, allocation *models.Allocation) error {
	return nil
}
func (m *MockRepository) GetAllocation(ctx context.Context, allocationID string) (*models.Allocation, error) {
	return nil, nil
}
func (m *MockRepository) UpdateAllocation(ctx context.Context, allocation *models.Allocation) error {
	return nil
}
func (m *MockRepository) DeleteAllocation(ctx context.Context, allocationID string) error { return nil }
func (m *MockRepository) GetJobAllocations(ctx context.Context, jobID string) ([]*models.Allocation, error) {
	return []*models.Allocation{}, nil
}
func (m *MockRepository) ListActiveAllocations(ctx context.Context) ([]*models.Allocation, error) {
	return []*models.Allocation{}, nil
}
func (m *MockRepository) Ping(ctx context.Context) error { return nil }
func (m *MockRepository) Close() error                   { return nil }

func BenchmarkJobSubmission(b *testing.B) {
	storage := &MockRepository{}
	config := &utils.SchedulerConfig{
		SchedulingInterval: 1000,
		MaxQueueSize:       100000,
	}
	scheduler := core.NewScheduler(config, storage)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := &models.Job{
			ID:          "bench-job-" + string(rune(i)),
			TenantID:    "bench-tenant",
			Name:        "Benchmark Job",
			Priority:    100,
			GPUCount:    1,
			GPUMemoryMB: 16000,
			CPUCores:    4,
			MemoryMB:    32000,
		}
		scheduler.SubmitJob(ctx, job)
	}
}

func BenchmarkQueueOperations(b *testing.B) {
	q := core.NewQueue(100000)

	b.Run("Enqueue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			job := &models.Job{
				ID:       "job-" + string(rune(i)),
				Priority: i % 1000,
				GPUCount: 1,
			}
			q.Enqueue(job)
		}
	})

	b.Run("Peek", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			q.Peek()
		}
	})

	b.Run("GetPosition", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			q.GetPosition("job-100")
		}
	})
}

func BenchmarkPriorityQueueWithAging(b *testing.B) {
	q := core.NewQueue(10000)

	// Populate queue
	for i := 0; i < 1000; i++ {
		q.Enqueue(&models.Job{
			ID:       "job-" + string(rune(i)),
			Priority: i % 100,
			GPUCount: 1,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.ApplyAging(10, 5*time.Minute)
	}
}

func BenchmarkTenantQuotaCheck(b *testing.B) {
	tenant := &models.Tenant{
		MaxGPUs:           10,
		MaxGPUMemoryMB:    160000,
		MaxCPUCores:       64,
		MaxMemoryMB:       256000,
		MaxConcurrentJobs: 20,
		CurrentGPUs:       5,
		CurrentGPUMemory:  80000,
		CurrentCPUCores:   32,
		CurrentMemory:     128000,
		CurrentJobs:       10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tenant.HasAvailableQuota(2, 32000, 16, 64000)
	}
}

func BenchmarkGPUHealthUpdate(b *testing.B) {
	gpu := &models.GPU{
		Temperature: 75.0,
		ErrorCount:  0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gpu.UpdateMetrics(80.0, 77.0, 250.0)
	}
}
