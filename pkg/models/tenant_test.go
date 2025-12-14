package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasAvailableQuota(t *testing.T) {
	tenant := &Tenant{
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

	tests := []struct {
		name        string
		gpus        int
		gpuMemory   int64
		cpus        int
		memory      int64
		expected    bool
		description string
	}{
		{
			name:        "Within quota",
			gpus:        2,
			gpuMemory:   32000,
			cpus:        16,
			memory:      64000,
			expected:    true,
			description: "Should allow when within limits",
		},
		{
			name:        "Exceeds GPU quota",
			gpus:        6,
			gpuMemory:   32000,
			cpus:        16,
			memory:      64000,
			expected:    false,
			description: "Should deny when GPU quota exceeded",
		},
		{
			name:        "Exceeds memory quota",
			gpus:        2,
			gpuMemory:   100000,
			cpus:        16,
			memory:      64000,
			expected:    false,
			description: "Should deny when GPU memory exceeded",
		},
		{
			name:        "At exact limit",
			gpus:        5,
			gpuMemory:   80000,
			cpus:        32,
			memory:      128000,
			expected:    true,
			description: "Should allow at exact quota",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tenant.HasAvailableQuota(tt.gpus, tt.gpuMemory, tt.cpus, tt.memory)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestUpdateUsage(t *testing.T) {
	tenant := &Tenant{
		CurrentGPUs:      5,
		CurrentGPUMemory: 80000,
		CurrentCPUCores:  32,
		CurrentMemory:    128000,
		CurrentJobs:      10,
	}

	// Increase usage
	tenant.UpdateUsage(2, 32000, 16, 64000, 1)
	assert.Equal(t, 7, tenant.CurrentGPUs)
	assert.Equal(t, int64(112000), tenant.CurrentGPUMemory)
	assert.Equal(t, 48, tenant.CurrentCPUCores)
	assert.Equal(t, int64(192000), tenant.CurrentMemory)
	assert.Equal(t, 11, tenant.CurrentJobs)

	// Decrease usage
	tenant.UpdateUsage(-3, -48000, -24, -96000, -2)
	assert.Equal(t, 4, tenant.CurrentGPUs)
	assert.Equal(t, int64(64000), tenant.CurrentGPUMemory)
	assert.Equal(t, 24, tenant.CurrentCPUCores)
	assert.Equal(t, int64(96000), tenant.CurrentMemory)
	assert.Equal(t, 9, tenant.CurrentJobs)
}

func TestCalculateFairShare(t *testing.T) {
	tests := []struct {
		name        string
		maxGPUs     int
		currentGPUs int
		expected    float64
	}{
		{"50% usage", 10, 5, 0.5},
		{"100% usage", 10, 10, 1.0},
		{"0% usage", 10, 0, 0.0},
		{"No quota", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{
				MaxGPUs:     tt.maxGPUs,
				CurrentGPUs: tt.currentGPUs,
			}
			assert.Equal(t, tt.expected, tenant.CalculateFairShare())
		})
	}
}

func TestGetPriorityScore(t *testing.T) {
	tests := []struct {
		tier     PriorityTier
		expected int
	}{
		{PriorityLow, 100},
		{PriorityMedium, 500},
		{PriorityHigh, 1000},
		{PriorityCritical, 5000},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			tenant := &Tenant{PriorityTier: tt.tier}
			assert.Equal(t, tt.expected, tenant.GetPriorityScore())
		})
	}
}
