package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGPUIsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		gpu      *GPU
		expected bool
	}{
		{
			name: "Available GPU",
			gpu: &GPU{
				Allocated:       false,
				Health:          HealthHealthy,
				ThermalThrottle: false,
				CoolingPeriod:   time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "Allocated GPU",
			gpu: &GPU{
				Allocated:       true,
				Health:          HealthHealthy,
				ThermalThrottle: false,
			},
			expected: false,
		},
		{
			name: "Unhealthy GPU",
			gpu: &GPU{
				Allocated:       false,
				Health:          HealthUnhealthy,
				ThermalThrottle: false,
			},
			expected: false,
		},
		{
			name: "Thermal throttling",
			gpu: &GPU{
				Allocated:       false,
				Health:          HealthHealthy,
				ThermalThrottle: true,
			},
			expected: false,
		},
		{
			name: "Still in cooling period",
			gpu: &GPU{
				Allocated:       false,
				Health:          HealthHealthy,
				ThermalThrottle: false,
				CoolingPeriod:   time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.gpu.IsAvailable())
		})
	}
}

func TestNeedsCooling(t *testing.T) {
	gpu := &GPU{Temperature: 80.0}
	assert.True(t, gpu.NeedsCooling(75.0))
	assert.False(t, gpu.NeedsCooling(85.0))
}

func TestUpdateMetrics(t *testing.T) {
	gpu := &GPU{
		MaxTemperature: 60.0,
		ErrorCount:     0,
	}

	gpu.UpdateMetrics(75.5, 82.0, 250.0)

	assert.Equal(t, 75.5, gpu.Utilization)
	assert.Equal(t, 82.0, gpu.Temperature)
	assert.Equal(t, 250.0, gpu.PowerUsage)
	assert.Equal(t, 82.0, gpu.MaxTemperature) // Should update max
	assert.Equal(t, HealthDegraded, gpu.Health) // Temp > 75
}

func TestUpdateHealth(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		errorCount  int
		expected    GPUHealth
	}{
		{"Healthy", 60.0, 0, HealthHealthy},
		{"Warning temp", 70.0, 0, HealthWarning},
		{"Warning errors", 60.0, 3, HealthWarning},
		{"Degraded temp", 80.0, 0, HealthDegraded},
		{"Degraded errors", 60.0, 7, HealthDegraded},
		{"Unhealthy temp", 90.0, 0, HealthUnhealthy},
		{"Unhealthy errors", 60.0, 15, HealthUnhealthy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpu := &GPU{
				Temperature: tt.temperature,
				ErrorCount:  tt.errorCount,
			}
			gpu.UpdateHealth()
			assert.Equal(t, tt.expected, gpu.Health)
		})
	}
}

func TestNodeHasCapacity(t *testing.T) {
	node := &Node{
		Online:             true,
		Schedulable:        true,
		DrainingMode:       false,
		AvailableGPUs:      8,
		AvailableCPUCores:  64,
		AvailableMemoryMB:  256000,
	}

	tests := []struct {
		name     string
		gpus     int
		cpus     int
		memory   int64
		expected bool
	}{
		{"Within capacity", 4, 32, 128000, true},
		{"At capacity", 8, 64, 256000, true},
		{"Exceeds GPU", 10, 32, 128000, false},
		{"Exceeds CPU", 4, 128, 128000, false},
		{"Exceeds memory", 4, 32, 512000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := node.HasCapacity(tt.gpus, tt.cpus, tt.memory)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNodeNotSchedulable(t *testing.T) {
	tests := []struct {
		name         string
		online       bool
		schedulable  bool
		drainingMode bool
		expected     bool
	}{
		{"Fully available", true, true, false, true},
		{"Offline", false, true, false, false},
		{"Not schedulable", true, false, false, false},
		{"Draining", true, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{
				Online:             tt.online,
				Schedulable:        tt.schedulable,
				DrainingMode:       tt.drainingMode,
				AvailableGPUs:      8,
				AvailableCPUCores:  64,
				AvailableMemoryMB:  256000,
			}
			result := node.HasCapacity(4, 32, 128000)
			assert.Equal(t, tt.expected, result)
		})
	}
}
