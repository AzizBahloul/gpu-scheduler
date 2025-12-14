package models

import (
	"time"
)

// PriorityTier defines the priority level of a tenant
type PriorityTier string

const (
	PriorityLow      PriorityTier = "low"
	PriorityMedium   PriorityTier = "medium"
	PriorityHigh     PriorityTier = "high"
	PriorityCritical PriorityTier = "critical"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID                string        `json:"id" gorm:"primaryKey"`
	Name              string        `json:"name"`
	Email             string        `json:"email"`
	Organization      string        `json:"organization"`
	
	// Quotas
	MaxGPUs           int           `json:"max_gpus"`
	MaxGPUMemoryMB    int64         `json:"max_gpu_memory_mb"`
	MaxCPUCores       int           `json:"max_cpu_cores"`
	MaxMemoryMB       int64         `json:"max_memory_mb"`
	MaxConcurrentJobs int           `json:"max_concurrent_jobs"`
	
	// Current Usage
	CurrentGPUs       int           `json:"current_gpus"`
	CurrentGPUMemory  int64         `json:"current_gpu_memory"`
	CurrentCPUCores   int           `json:"current_cpu_cores"`
	CurrentMemory     int64         `json:"current_memory"`
	CurrentJobs       int           `json:"current_jobs"`
	
	// Historical Usage
	TotalGPUHours     float64       `json:"total_gpu_hours"`
	TotalJobs         int           `json:"total_jobs"`
	SuccessfulJobs    int           `json:"successful_jobs"`
	FailedJobs        int           `json:"failed_jobs"`
	
	// Priority and Fairness
	PriorityTier      PriorityTier  `json:"priority_tier"`
	FairShareWeight   float64       `json:"fair_share_weight"`
	PriorityDecay     float64       `json:"priority_decay"`
	
	// Policies
	AllowPreemption   bool          `json:"allow_preemption"`
	CanPreemptOthers  bool          `json:"can_preempt_others"`
	MaxPreemptions    int           `json:"max_preemptions"`
	
	// Billing
	BillingEnabled    bool          `json:"billing_enabled"`
	CostPerGPUHour    float64       `json:"cost_per_gpu_hour"`
	TotalCost         float64       `json:"total_cost"`
	
	// Notifications
	NotifyOnStart     bool          `json:"notify_on_start"`
	NotifyOnComplete  bool          `json:"notify_on_complete"`
	NotifyOnFailure   bool          `json:"notify_on_failure"`
	NotifyEmail       string        `json:"notify_email"`
	
	// Metadata
	Active            bool          `json:"active"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// HasAvailableQuota checks if tenant has quota for the requested resources
func (t *Tenant) HasAvailableQuota(gpus int, gpuMemory int64, cpus int, memory int64) bool {
	return t.CurrentGPUs+gpus <= t.MaxGPUs &&
		t.CurrentGPUMemory+gpuMemory <= t.MaxGPUMemoryMB &&
		t.CurrentCPUCores+cpus <= t.MaxCPUCores &&
		t.CurrentMemory+memory <= t.MaxMemoryMB &&
		t.CurrentJobs+1 <= t.MaxConcurrentJobs
}

// UpdateUsage updates current resource usage
func (t *Tenant) UpdateUsage(gpuDelta int, gpuMemDelta int64, cpuDelta int, memDelta int64, jobDelta int) {
	t.CurrentGPUs += gpuDelta
	t.CurrentGPUMemory += gpuMemDelta
	t.CurrentCPUCores += cpuDelta
	t.CurrentMemory += memDelta
	t.CurrentJobs += jobDelta
}

// CalculateFairShare calculates fair share ratio based on usage
func (t *Tenant) CalculateFairShare() float64 {
	if t.MaxGPUs == 0 {
		return 0
	}
	return float64(t.CurrentGPUs) / float64(t.MaxGPUs)
}

// GetPriorityScore returns priority score for scheduling
func (t *Tenant) GetPriorityScore() int {
	scores := map[PriorityTier]int{
		PriorityLow:      100,
		PriorityMedium:   500,
		PriorityHigh:     1000,
		PriorityCritical: 5000,
	}
	return scores[t.PriorityTier]
}
