package core

import (
	"context"
	"fmt"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/storage"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
	"go.uber.org/zap"
)

// Preemptor handles job preemption
type Preemptor struct {
	storage storage.Repository
}

// NewPreemptor creates a new preemptor
func NewPreemptor(storage storage.Repository) *Preemptor {
	return &Preemptor{
		storage: storage,
	}
}

// SelectVictims selects jobs to preempt
func (p *Preemptor) SelectVictims(ctx context.Context, requestingJob *models.Job) ([]*models.Job, error) {
	// Get all running jobs
	runningJobs, err := p.storage.ListJobsByState(ctx, models.JobStateRunning)
	if err != nil {
		return nil, err
	}

	var candidates []*models.Job

	// Find lower priority jobs
	for _, job := range runningJobs {
		if job.Priority < requestingJob.Priority {
			// Check if tenant allows preemption
			tenant, err := p.storage.GetTenant(ctx, job.TenantID)
			if err != nil {
				continue
			}

			if tenant.AllowPreemption {
				candidates = append(candidates, job)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Select victims based on cost
	// For now, select the lowest priority job
	var victim *models.Job
	lowestPriority := 999999

	for _, candidate := range candidates {
		if candidate.Priority < lowestPriority {
			lowestPriority = candidate.Priority
			victim = candidate
		}
	}

	if victim != nil {
		return []*models.Job{victim}, nil
	}

	return nil, nil
}

// Preempt preempts a running job
func (p *Preemptor) Preempt(ctx context.Context, victim *models.Job, preemptorID string) error {
	utils.Info("Preempting job", 
		zap.String("victim_id", victim.ID),
		zap.String("preemptor_id", preemptorID))

	// Update job state
	victim.State = models.JobStatePreempted
	victim.PreemptedCount++
	now := time.Now()

	if err := p.storage.UpdateJob(ctx, victim); err != nil {
		return fmt.Errorf("failed to update victim job: %w", err)
	}

	// Update allocations
	allocations, err := p.storage.GetJobAllocations(ctx, victim.ID)
	if err != nil {
		return err
	}

	for _, alloc := range allocations {
		alloc.State = models.AllocationPreempted
		alloc.PreemptedAt = &now
		alloc.PreemptedBy = preemptorID
		alloc.PreemptionReason = "higher priority job"

		if err := p.storage.UpdateAllocation(ctx, alloc); err != nil {
			utils.Error("Failed to update allocation", 
				zap.String("allocation_id", alloc.ID),
				zap.Error(err))
		}

		// Free the GPUs
		for _, gpuID := range alloc.GPUIDs {
			gpu, err := p.storage.GetGPU(ctx, gpuID)
			if err != nil {
				continue
			}

			gpu.Allocated = false
			gpu.AllocationID = ""
			gpu.JobID = ""
			gpu.TenantID = ""

			if err := p.storage.UpdateGPU(ctx, gpu); err != nil {
				utils.Error("Failed to free GPU", 
					zap.String("gpu_id", gpuID),
					zap.Error(err))
			}
		}

		// Update node capacity
		node, err := p.storage.GetNode(ctx, alloc.NodeID)
		if err != nil {
			continue
		}

		node.AvailableGPUs += len(alloc.GPUIDs)
		node.AvailableCPUCores += alloc.CPUCores
		node.AvailableMemoryMB += alloc.MemoryMB

		if err := p.storage.UpdateNode(ctx, node); err != nil {
			utils.Error("Failed to update node", 
				zap.String("node_id", node.ID),
				zap.Error(err))
		}
	}

	utils.Info("Job preempted successfully", zap.String("victim_id", victim.ID))
	return nil
}
