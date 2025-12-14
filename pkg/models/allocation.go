package models

import (
	"time"
)

// AllocationState represents the state of a resource allocation
type AllocationState string

const (
	AllocationPending      AllocationState = "pending"
	AllocationActive       AllocationState = "active"
	AllocationPreempted    AllocationState = "preempted"
	AllocationCheckpointed AllocationState = "checkpointed"
	AllocationMigrating    AllocationState = "migrating"
	AllocationCompleted    AllocationState = "completed"
	AllocationFailed       AllocationState = "failed"
)

// Allocation represents a resource allocation for a job
type Allocation struct {
	ID                string           `json:"id" gorm:"primaryKey"`
	JobID             string           `json:"job_id" gorm:"index"`
	TenantID          string           `json:"tenant_id" gorm:"index"`
	State             AllocationState  `json:"state"`
	
	// Resources
	GPUIDs            []string         `json:"gpu_ids" gorm:"serializer:json"`
	NodeID            string           `json:"node_id"`
	CPUCores          int              `json:"cpu_cores"`
	MemoryMB          int64            `json:"memory_mb"`
	
	// Timing
	AllocatedAt       time.Time        `json:"allocated_at"`
	PlannedDuration   time.Duration    `json:"planned_duration"`
	ActualDuration    time.Duration    `json:"actual_duration"`
	ExtendedCount     int              `json:"extended_count"`
	
	// Preemption
	PreemptedAt       *time.Time       `json:"preempted_at"`
	PreemptedBy       string           `json:"preempted_by"`
	PreemptionReason  string           `json:"preemption_reason"`
	CheckpointSize    int64            `json:"checkpoint_size"`
	CheckpointPath    string           `json:"checkpoint_path"`
	
	// Performance
	AvgGPUUtilization float64          `json:"avg_gpu_utilization"`
	PeakGPUUtilization float64         `json:"peak_gpu_utilization"`
	AvgPowerUsage     float64          `json:"avg_power_usage"`
	
	// Cost
	CostPerHour       float64          `json:"cost_per_hour"`
	TotalCost         float64          `json:"total_cost"`
	
	// Metadata
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	CompletedAt       *time.Time       `json:"completed_at"`
}

// AllocationRequest represents a request for resource allocation
type AllocationRequest struct {
	JobID             string           `json:"job_id"`
	TenantID          string           `json:"tenant_id"`
	GPUCount          int              `json:"gpu_count"`
	GPUMemoryMB       int64            `json:"gpu_memory_mb"`
	CPUCores          int              `json:"cpu_cores"`
	MemoryMB          int64            `json:"memory_mb"`
	GangScheduling    bool             `json:"gang_scheduling"`
	PreferredNodes    []string         `json:"preferred_nodes"`
	RequiredLabels    map[string]string `json:"required_labels"`
	Affinity          *Affinity        `json:"affinity"`
}

// Affinity defines scheduling affinity rules
type Affinity struct {
	NodeAffinity      *NodeAffinity    `json:"node_affinity"`
	GPUModel          GPUModel         `json:"gpu_model"`
	ColocateWithJob   string           `json:"colocate_with_job"`
	AntiColocateWith  []string         `json:"anti_colocate_with"`
}

// NodeAffinity defines node affinity rules
type NodeAffinity struct {
	RequiredLabels    map[string]string `json:"required_labels"`
	PreferredLabels   map[string]string `json:"preferred_labels"`
	PreferredNodes    []string          `json:"preferred_nodes"`
}

// AllocationResult represents the result of an allocation attempt
type AllocationResult struct {
	Success           bool             `json:"success"`
	AllocationID      string           `json:"allocation_id"`
	GPUIDs            []string         `json:"gpu_ids"`
	NodeID            string           `json:"node_id"`
	Message           string           `json:"message"`
	Timestamp         time.Time        `json:"timestamp"`
}

// IsActive returns true if allocation is active
func (a *Allocation) IsActive() bool {
	return a.State == AllocationActive
}

// CalculateDuration updates actual duration
func (a *Allocation) CalculateDuration() {
	if a.CompletedAt != nil {
		a.ActualDuration = a.CompletedAt.Sub(a.AllocatedAt)
	}
}

// CalculateCost calculates total cost based on duration
func (a *Allocation) CalculateCost() {
	hours := a.ActualDuration.Hours()
	a.TotalCost = hours * a.CostPerHour * float64(len(a.GPUIDs))
}

// UpdateUtilization updates GPU utilization metrics
func (a *Allocation) UpdateUtilization(current float64) {
	if current > a.PeakGPUUtilization {
		a.PeakGPUUtilization = current
	}
	
	// Simple moving average
	if a.AvgGPUUtilization == 0 {
		a.AvgGPUUtilization = current
	} else {
		a.AvgGPUUtilization = (a.AvgGPUUtilization + current) / 2
	}
}
