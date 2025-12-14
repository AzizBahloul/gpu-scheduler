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

// Allocator handles resource allocation
type Allocator struct {
	storage storage.Repository
}

// NewAllocator creates a new allocator
func NewAllocator(storage storage.Repository) *Allocator {
	return &Allocator{
		storage: storage,
	}
}

// Allocate attempts to allocate resources for a job
func (a *Allocator) Allocate(ctx context.Context, request *models.AllocationRequest) (*models.AllocationResult, error) {
	utils.Debug("Attempting allocation", 
		zap.String("job_id", request.JobID),
		zap.Int("gpu_count", request.GPUCount))

	// Get available nodes
	nodes, err := a.storage.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Filter schedulable nodes
	var availableNodes []*models.Node
	for _, node := range nodes {
		if node.HasCapacity(request.GPUCount, request.CPUCores, request.MemoryMB) {
			availableNodes = append(availableNodes, node)
		}
	}

	if len(availableNodes) == 0 {
		return &models.AllocationResult{
			Success: false,
			Message: "no nodes with sufficient capacity",
		}, utils.ErrInsufficientResources
	}

	// Try gang scheduling if requested
	if request.GangScheduling {
		return a.gangSchedule(ctx, request, availableNodes)
	}

	// Try to allocate on best-fit node
	return a.bestFitSchedule(ctx, request, availableNodes)
}

// bestFitSchedule uses best-fit algorithm
func (a *Allocator) bestFitSchedule(ctx context.Context, request *models.AllocationRequest, nodes []*models.Node) (*models.AllocationResult, error) {
	var bestNode *models.Node
	var bestGPUs []*models.GPU

	// Find node with least fragmentation
	minWaste := int64(999999)

	for _, node := range nodes {
		gpus, err := a.storage.ListGPUsByNode(ctx, node.ID)
		if err != nil {
			continue
		}

		// Find available GPUs
		var availGPUs []*models.GPU
		for _, gpu := range gpus {
			if gpu.IsAvailable() {
				availGPUs = append(availGPUs, gpu)
			}
		}

		if len(availGPUs) >= request.GPUCount {
			waste := int64(len(availGPUs) - request.GPUCount)
			if waste < minWaste {
				minWaste = waste
				bestNode = node
				bestGPUs = availGPUs[:request.GPUCount]
			}
		}
	}

	if bestNode == nil {
		return &models.AllocationResult{
			Success: false,
			Message: "no suitable node found",
		}, utils.ErrInsufficientResources
	}

	// Create allocation
	return a.createAllocation(ctx, request, bestNode, bestGPUs)
}

// gangSchedule allocates all resources atomically
func (a *Allocator) gangSchedule(ctx context.Context, request *models.AllocationRequest, nodes []*models.Node) (*models.AllocationResult, error) {
	// For simplicity, try to allocate on a single node
	// Production version would support multi-node gang scheduling
	
	for _, node := range nodes {
		if node.AvailableGPUs >= request.GPUCount {
			gpus, err := a.storage.ListGPUsByNode(ctx, node.ID)
			if err != nil {
				continue
			}

			var availGPUs []*models.GPU
			for _, gpu := range gpus {
				if gpu.IsAvailable() {
					availGPUs = append(availGPUs, gpu)
					if len(availGPUs) == request.GPUCount {
						break
					}
				}
			}

			if len(availGPUs) == request.GPUCount {
				return a.createAllocation(ctx, request, node, availGPUs)
			}
		}
	}

	return &models.AllocationResult{
		Success: false,
		Message: "gang scheduling failed - insufficient resources on single node",
	}, utils.ErrGangSchedulingFailed
}

// createAllocation creates and persists an allocation
func (a *Allocator) createAllocation(ctx context.Context, request *models.AllocationRequest, node *models.Node, gpus []*models.GPU) (*models.AllocationResult, error) {
	gpuIDs := make([]string, len(gpus))
	for i, gpu := range gpus {
		gpuIDs[i] = gpu.ID
	}

	allocation := &models.Allocation{
		ID:             generateAllocationID(),
		JobID:          request.JobID,
		TenantID:       request.TenantID,
		State:          models.AllocationActive,
		GPUIDs:         gpuIDs,
		NodeID:         node.ID,
		CPUCores:       request.CPUCores,
		MemoryMB:       request.MemoryMB,
		AllocatedAt:    time.Now(),
		PlannedDuration: 1 * time.Hour, // Default
	}

	// Save allocation
	if err := a.storage.CreateAllocation(ctx, allocation); err != nil {
		return nil, fmt.Errorf("failed to create allocation: %w", err)
	}

	// Mark GPUs as allocated
	for _, gpu := range gpus {
		gpu.Allocated = true
		gpu.AllocationID = allocation.ID
		gpu.JobID = request.JobID
		gpu.TenantID = request.TenantID
		
		if err := a.storage.UpdateGPU(ctx, gpu); err != nil {
			utils.Error("Failed to update GPU", zap.String("gpu_id", gpu.ID), zap.Error(err))
		}
	}

	// Update node capacity
	node.AvailableGPUs -= len(gpus)
	node.AvailableCPUCores -= request.CPUCores
	node.AvailableMemoryMB -= request.MemoryMB
	
	if err := a.storage.UpdateNode(ctx, node); err != nil {
		utils.Error("Failed to update node", zap.String("node_id", node.ID), zap.Error(err))
	}

	utils.Info("Allocation created", 
		zap.String("allocation_id", allocation.ID),
		zap.String("job_id", request.JobID),
		zap.String("node_id", node.ID),
		zap.Int("gpus", len(gpus)))

	return &models.AllocationResult{
		Success:      true,
		AllocationID: allocation.ID,
		GPUIDs:       gpuIDs,
		NodeID:       node.ID,
		Timestamp:    time.Now(),
	}, nil
}

// Free releases an allocation
func (a *Allocator) Free(ctx context.Context, allocationID string) error {
	allocation, err := a.storage.GetAllocation(ctx, allocationID)
	if err != nil {
		return err
	}

	allocation.State = models.AllocationCompleted
	now := time.Now()
	allocation.CompletedAt = &now
	allocation.CalculateDuration()

	if err := a.storage.UpdateAllocation(ctx, allocation); err != nil {
		return err
	}

	// Free GPUs
	for _, gpuID := range allocation.GPUIDs {
		gpu, err := a.storage.GetGPU(ctx, gpuID)
		if err != nil {
			continue
		}

		gpu.Allocated = false
		gpu.AllocationID = ""
		gpu.JobID = ""
		gpu.TenantID = ""

		if err := a.storage.UpdateGPU(ctx, gpu); err != nil {
			utils.Error("Failed to free GPU", zap.String("gpu_id", gpuID), zap.Error(err))
		}
	}

	// Update node capacity
	node, err := a.storage.GetNode(ctx, allocation.NodeID)
	if err != nil {
		return err
	}

	node.AvailableGPUs += len(allocation.GPUIDs)
	node.AvailableCPUCores += allocation.CPUCores
	node.AvailableMemoryMB += allocation.MemoryMB

	if err := a.storage.UpdateNode(ctx, node); err != nil {
		return err
	}

	utils.Info("Allocation freed", 
		zap.String("allocation_id", allocationID),
		zap.String("job_id", allocation.JobID))

	return nil
}

func generateAllocationID() string {
	return fmt.Sprintf("alloc-%d", time.Now().UnixNano())
}
