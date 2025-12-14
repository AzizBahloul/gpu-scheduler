package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/storage"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
	"go.uber.org/zap"
)

// Scheduler is the main scheduling orchestrator
type Scheduler struct {
	queue       *Queue
	allocator   *Allocator
	preemptor   *Preemptor
	storage     storage.Repository
	config      *utils.SchedulerConfig
	
	mu          sync.RWMutex
	running     bool
	stopChan    chan struct{}
	
	// Metrics
	scheduledJobs   int64
	failedJobs      int64
	preemptedJobs   int64
}

// NewScheduler creates a new scheduler instance
func NewScheduler(config *utils.SchedulerConfig, storage storage.Repository) *Scheduler {
	queue := NewQueue(config.MaxQueueSize)
	allocator := NewAllocator(storage)
	preemptor := NewPreemptor(storage)

	return &Scheduler{
		queue:     queue,
		allocator: allocator,
		preemptor: preemptor,
		storage:   storage,
		config:    config,
		stopChan:  make(chan struct{}),
	}
}

// Start begins the scheduling loop
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	utils.Info("Starting scheduler", zap.Int("interval_ms", s.config.SchedulingInterval))

	ticker := time.NewTicker(time.Duration(s.config.SchedulingInterval) * time.Millisecond)
	defer ticker.Stop()

	// Load pending jobs from storage
	if err := s.loadPendingJobs(ctx); err != nil {
		utils.Error("Failed to load pending jobs", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			utils.Info("Scheduler stopping due to context cancellation")
			return ctx.Err()
		case <-s.stopChan:
			utils.Info("Scheduler stopping")
			return nil
		case <-ticker.C:
			if err := s.schedulingCycle(ctx); err != nil {
				utils.Error("Scheduling cycle error", zap.Error(err))
			}
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false
	utils.Info("Scheduler stopped")
}

// SubmitJob submits a new job to the scheduler
func (s *Scheduler) SubmitJob(ctx context.Context, job *models.Job) error {
	utils.Info("Submitting job", zap.String("job_id", job.ID), zap.String("tenant_id", job.TenantID))

	// Validate job
	if err := s.validateJob(ctx, job); err != nil {
		return fmt.Errorf("job validation failed: %w", err)
	}

	// Check tenant quota
	tenant, err := s.storage.GetTenant(ctx, job.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if !tenant.HasAvailableQuota(job.GPUCount, job.GPUMemoryMB, job.CPUCores, job.MemoryMB) {
		return &utils.QuotaExceededError{
			TenantID: tenant.ID,
			Resource: "GPUs",
			Requested: job.GPUCount,
			Quota: tenant.MaxGPUs,
			Current: tenant.CurrentGPUs,
		}
	}

	// Set job state
	job.State = models.JobStatePending
	job.SubmittedAt = time.Now()

	// Save to storage
	if err := s.storage.CreateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Add to queue
	if err := s.queue.Enqueue(job); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	utils.Info("Job submitted successfully", 
		zap.String("job_id", job.ID),
		zap.Int("queue_size", s.queue.Size()))

	return nil
}

// CancelJob cancels a running or pending job
func (s *Scheduler) CancelJob(ctx context.Context, jobID string) error {
	utils.Info("Cancelling job", zap.String("job_id", jobID))

	job, err := s.storage.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	switch job.State {
	case models.JobStatePending:
		// Remove from queue
		s.queue.Remove(jobID)
		job.State = models.JobStateCancelled
		job.CompletedAt = timePtr(time.Now())
		
	case models.JobStateRunning:
		// Cancel running job
		job.State = models.JobStateCancelled
		job.CompletedAt = timePtr(time.Now())
		
		// Free resources
		if err := s.freeJobResources(ctx, job); err != nil {
			utils.Error("Failed to free job resources", zap.Error(err))
		}
		
	default:
		return fmt.Errorf("cannot cancel job in state: %s", job.State)
	}

	if err := s.storage.UpdateJob(ctx, job); err != nil {
		return err
	}

	utils.Info("Job cancelled", zap.String("job_id", jobID))
	return nil
}

// GetJobStatus returns the current status of a job
func (s *Scheduler) GetJobStatus(ctx context.Context, jobID string) (*models.JobStatus, error) {
	job, err := s.storage.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	status := &models.JobStatus{
		JobID:   job.ID,
		State:   job.State,
		Message: "",
	}

	if job.State == models.JobStatePending {
		status.QueuePosition = s.queue.GetPosition(jobID)
		status.EstimatedWait = s.estimateWaitTime(job)
	}

	// Get allocation info if running
	if job.State == models.JobStateRunning {
		allocations, err := s.storage.GetJobAllocations(ctx, jobID)
		if err == nil && len(allocations) > 0 {
			status.AllocatedGPUs = allocations[0].GPUIDs
			status.NodeName = allocations[0].NodeID
		}
	}

	return status, nil
}

// schedulingCycle performs one scheduling cycle
func (s *Scheduler) schedulingCycle(ctx context.Context) error {
	// Apply aging to prevent starvation
	s.queue.ApplyAging(10, 5*time.Minute)

	// Process pending jobs
	for !s.queue.IsEmpty() {
		job := s.queue.Peek()
		if job == nil {
			break
		}

		// Try to allocate resources
		allocated, err := s.tryAllocateJob(ctx, job)
		if err != nil {
			utils.Error("Allocation error", 
				zap.String("job_id", job.ID), 
				zap.Error(err))
			
			// If resource error, try preemption if enabled
			if s.config.EnablePreemption && utils.IsResourceError(err) {
				if s.tryPreemption(ctx, job) {
					continue
				}
			}
			
			// Can't schedule this job now, try next one
			break
		}

		if allocated {
			// Remove from queue and start job
			s.queue.Dequeue()
			if err := s.startJob(ctx, job); err != nil {
				utils.Error("Failed to start job", 
					zap.String("job_id", job.ID), 
					zap.Error(err))
				s.failedJobs++
			} else {
				s.scheduledJobs++
			}
		} else {
			// No resources available, stop trying
			break
		}
	}

	return nil
}

// tryAllocateJob attempts to allocate resources for a job
func (s *Scheduler) tryAllocateJob(ctx context.Context, job *models.Job) (bool, error) {
	request := &models.AllocationRequest{
		JobID:          job.ID,
		TenantID:       job.TenantID,
		GPUCount:       job.GPUCount,
		GPUMemoryMB:    job.GPUMemoryMB,
		CPUCores:       job.CPUCores,
		MemoryMB:       job.MemoryMB,
		GangScheduling: job.GangScheduling,
	}

	result, err := s.allocator.Allocate(ctx, request)
	if err != nil {
		return false, err
	}

	return result.Success, nil
}

// startJob transitions a job to running state
func (s *Scheduler) startJob(ctx context.Context, job *models.Job) error {
	now := time.Now()
	job.State = models.JobStateRunning
	job.ScheduledAt = &now
	job.StartedAt = &now

	if err := s.storage.UpdateJob(ctx, job); err != nil {
		return err
	}

	// Update tenant usage
	tenant, err := s.storage.GetTenant(ctx, job.TenantID)
	if err != nil {
		return err
	}

	tenant.UpdateUsage(job.GPUCount, job.GPUMemoryMB, job.CPUCores, job.MemoryMB, 1)
	if err := s.storage.UpdateTenant(ctx, tenant); err != nil {
		return err
	}

	utils.Info("Job started", 
		zap.String("job_id", job.ID),
		zap.String("tenant_id", job.TenantID))

	return nil
}

// tryPreemption attempts to preempt lower priority jobs
func (s *Scheduler) tryPreemption(ctx context.Context, job *models.Job) bool {
	if !s.config.EnablePreemption {
		return false
	}

	victims, err := s.preemptor.SelectVictims(ctx, job)
	if err != nil || len(victims) == 0 {
		return false
	}

	utils.Info("Attempting preemption", 
		zap.String("job_id", job.ID),
		zap.Int("victims", len(victims)))

	for _, victim := range victims {
		if err := s.preemptor.Preempt(ctx, victim, job.ID); err != nil {
			utils.Error("Preemption failed", 
				zap.String("victim_id", victim.ID),
				zap.Error(err))
			return false
		}
		s.preemptedJobs++
	}

	return true
}

// freeJobResources releases resources allocated to a job
func (s *Scheduler) freeJobResources(ctx context.Context, job *models.Job) error {
	allocations, err := s.storage.GetJobAllocations(ctx, job.ID)
	if err != nil {
		return err
	}

	for _, alloc := range allocations {
		if err := s.allocator.Free(ctx, alloc.ID); err != nil {
			utils.Error("Failed to free allocation", 
				zap.String("allocation_id", alloc.ID),
				zap.Error(err))
		}
	}

	// Update tenant usage
	tenant, err := s.storage.GetTenant(ctx, job.TenantID)
	if err != nil {
		return err
	}

	tenant.UpdateUsage(-job.GPUCount, -job.GPUMemoryMB, -job.CPUCores, -job.MemoryMB, -1)
	return s.storage.UpdateTenant(ctx, tenant)
}

// validateJob validates job parameters
func (s *Scheduler) validateJob(ctx context.Context, job *models.Job) error {
	if job.GPUCount <= 0 {
		return fmt.Errorf("GPU count must be positive")
	}
	if job.GPUCount > 8 {
		return fmt.Errorf("GPU count cannot exceed 8")
	}
	if job.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	return nil
}

// loadPendingJobs loads pending jobs from storage into queue
func (s *Scheduler) loadPendingJobs(ctx context.Context) error {
	jobs, err := s.storage.ListJobsByState(ctx, models.JobStatePending)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := s.queue.Enqueue(job); err != nil {
			utils.Error("Failed to enqueue pending job", 
				zap.String("job_id", job.ID),
				zap.Error(err))
		}
	}

	utils.Info("Loaded pending jobs", zap.Int("count", len(jobs)))
	return nil
}

// estimateWaitTime estimates wait time for a job
func (s *Scheduler) estimateWaitTime(job *models.Job) time.Duration {
	position := s.queue.GetPosition(job.ID)
	if position <= 0 {
		return 0
	}

	// Simple estimation: 5 minutes per job ahead
	return time.Duration(position-1) * 5 * time.Minute
}

func timePtr(t time.Time) *time.Time {
	return &t
}
