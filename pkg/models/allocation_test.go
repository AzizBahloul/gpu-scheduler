package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAllocationIsActive(t *testing.T) {
	tests := []struct {
		name     string
		state    AllocationState
		expected bool
	}{
		{"Active", AllocationActive, true},
		{"Pending", AllocationPending, false},
		{"Completed", AllocationCompleted, false},
		{"Preempted", AllocationPreempted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alloc := &Allocation{State: tt.state}
			assert.Equal(t, tt.expected, alloc.IsActive())
		})
	}
}

func TestCalculateDuration(t *testing.T) {
	start := time.Now()
	end := start.Add(3 * time.Hour)

	alloc := &Allocation{
		AllocatedAt: start,
		CompletedAt: &end,
	}

	alloc.CalculateDuration()
	assert.Equal(t, 3*time.Hour, alloc.ActualDuration)
}

func TestCalculateCost(t *testing.T) {
	alloc := &Allocation{
		GPUIDs:         []string{"gpu-1", "gpu-2"},
		ActualDuration: 2 * time.Hour,
		CostPerHour:    10.0,
	}

	alloc.CalculateCost()
	expected := 2.0 * 10.0 * 2 // 2 hours * $10/hour * 2 GPUs
	assert.Equal(t, expected, alloc.TotalCost)
}

func TestUpdateUtilization(t *testing.T) {
	alloc := &Allocation{}

	// First update
	alloc.UpdateUtilization(80.0)
	assert.Equal(t, 80.0, alloc.AvgGPUUtilization)
	assert.Equal(t, 80.0, alloc.PeakGPUUtilization)

	// Second update with higher value
	alloc.UpdateUtilization(95.0)
	assert.Equal(t, 87.5, alloc.AvgGPUUtilization) // (80+95)/2
	assert.Equal(t, 95.0, alloc.PeakGPUUtilization)

	// Third update with lower value
	alloc.UpdateUtilization(70.0)
	assert.Equal(t, 78.75, alloc.AvgGPUUtilization) // (87.5+70)/2
	assert.Equal(t, 95.0, alloc.PeakGPUUtilization) // Peak doesn't change
}
