// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/scheduler/core"
	"github.com/azizbahloul/gpu-scheduler/pkg/storage/postgres"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedulerIntegration(t *testing.T) {
	// Skip if no database available
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup database
	config := &utils.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Database: "gpu_scheduler_test",
		SSLMode:  "disable",
	}

	storage, err := postgres.NewPostgresRepository(config)
	require.NoError(t, err)
	defer storage.Close()

	// Create scheduler
	schedulerConfig := &utils.SchedulerConfig{
		SchedulingInterval: 100, // 100ms for faster testing
		MaxQueueSize:      100,
		EnablePreemption:  true,
	}

	scheduler := core.NewScheduler(schedulerConfig, storage)

	// Create test tenant
	ctx := context.Background()
	tenant := &models.Tenant{
		ID:                "test-tenant",
		Name:              "Test Tenant",
		MaxGPUs:           10,
		MaxCPUCores:       64,
		MaxMemoryMB:       256000,
		MaxConcurrentJobs: 20,
		Active:            true,
		PriorityTier:      models.PriorityMedium,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = storage.CreateTenant(ctx, tenant)
	require.NoError(t, err)

	// Submit multiple jobs
	jobs := []*models.Job{
		{
			ID:          "job-low",
			TenantID:    "test-tenant",
			Name:        "Low Priority Job",
			Priority:    100,
			GPUCount:    1,
			GPUMemoryMB: 16000,
			CPUCores:    4,
			MemoryMB:    32000,
		},
		{
			ID:          "job-high",
			TenantID:    "test-tenant",
			Name:        "High Priority Job",
			Priority:    1000,
			GPUCount:    1,
			GPUMemoryMB: 16000,
			CPUCores:    4,
			MemoryMB:    32000,
		},
	}

	for _, job := range jobs {
		err = scheduler.SubmitJob(ctx, job)
		require.NoError(t, err)
	}

	// Verify jobs are in queue
	allJobs, err := storage.ListJobsByState(ctx, models.JobStatePending)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allJobs), 2)

	// Cleanup
	for _, job := range jobs {
		storage.DeleteJob(ctx, job.ID)
	}
	storage.DeleteTenant(ctx, tenant.ID)
}

func TestEndToEndJobLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	config := &utils.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Database: "gpu_scheduler_test",
		SSLMode:  "disable",
	}

	storage, err := postgres.NewPostgresRepository(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Create tenant
	tenant := &models.Tenant{
		ID:          "lifecycle-tenant",
		Name:        "Lifecycle Test",
		MaxGPUs:     5,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = storage.CreateTenant(ctx, tenant)
	require.NoError(t, err)

	// Create job
	job := &models.Job{
		ID:          "lifecycle-job",
		TenantID:    "lifecycle-tenant",
		Name:        "Lifecycle Test Job",
		State:       models.JobStatePending,
		Priority:    500,
		GPUCount:    1,
		SubmittedAt: time.Now(),
	}

	err = storage.CreateJob(ctx, job)
	require.NoError(t, err)

	// Verify job exists
	fetchedJob, err := storage.GetJob(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, fetchedJob.ID)
	assert.Equal(t, models.JobStatePending, fetchedJob.State)

	// Update job to running
	job.State = models.JobStateRunning
	now := time.Now()
	job.StartedAt = &now
	err = storage.UpdateJob(ctx, job)
	require.NoError(t, err)

	// Verify update
	fetchedJob, err = storage.GetJob(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, models.JobStateRunning, fetchedJob.State)

	// Complete job
	job.State = models.JobStateCompleted
	completed := time.Now()
	job.CompletedAt = &completed
	err = storage.UpdateJob(ctx, job)
	require.NoError(t, err)

	// Cleanup
	storage.DeleteJob(ctx, job.ID)
	storage.DeleteTenant(ctx, tenant.ID)
}
