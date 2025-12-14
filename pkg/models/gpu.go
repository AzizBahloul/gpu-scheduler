package models

import (
	"time"
)

// GPUModel represents different GPU hardware models
type GPUModel string

const (
	GPUA100   GPUModel = "A100"
	GPUA10    GPUModel = "A10"
	GPUH100   GPUModel = "H100"
	GPUV100   GPUModel = "V100"
	GPUT4     GPUModel = "T4"
	GPUL4     GPUModel = "L4"
	GPURTX4090 GPUModel = "RTX4090"
)

// GPUHealth represents the health status of a GPU
type GPUHealth string

const (
	HealthHealthy   GPUHealth = "healthy"
	HealthWarning   GPUHealth = "warning"
	HealthDegraded  GPUHealth = "degraded"
	HealthUnhealthy GPUHealth = "unhealthy"
)

// GPU represents a physical GPU resource
type GPU struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	NodeID          string    `json:"node_id" gorm:"index"`
	Index           int       `json:"index"`
	Model           GPUModel  `json:"model"`
	MemoryTotalMB   int64     `json:"memory_total_mb"`
	MemoryFreeMB    int64     `json:"memory_free_mb"`
	MemoryUsedMB    int64     `json:"memory_used_mb"`
	
	// Current Allocation
	Allocated       bool      `json:"allocated"`
	AllocationID    string    `json:"allocation_id"`
	JobID           string    `json:"job_id"`
	TenantID        string    `json:"tenant_id"`
	
	// Performance Metrics
	Utilization     float64   `json:"utilization"`
	Temperature     float64   `json:"temperature"`
	PowerUsage      float64   `json:"power_usage"`
	PowerLimitW     float64   `json:"power_limit_w"`
	
	// Thermal Management
	MaxTemperature  float64   `json:"max_temperature"`
	CoolingPeriod   time.Time `json:"cooling_period"`
	ThermalThrottle bool      `json:"thermal_throttle"`
	
	// Health and Status
	Health          GPUHealth `json:"health"`
	ErrorCount      int       `json:"error_count"`
	LastError       string    `json:"last_error"`
	
	// Capabilities
	ComputeCapability string  `json:"compute_capability"`
	CUDACores       int       `json:"cuda_cores"`
	TensorCores     int       `json:"tensor_cores"`
	ClockSpeedMHz   int       `json:"clock_speed_mhz"`
	
	// Lifecycle
	LastHealthCheck time.Time `json:"last_health_check"`
	LastHeartbeat   time.Time `json:"last_heartbeat"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Node represents a physical node with GPUs
type Node struct {
	ID              string            `json:"id" gorm:"primaryKey"`
	Name            string            `json:"name"`
	IPAddress       string            `json:"ip_address"`
	Hostname        string            `json:"hostname"`
	
	// Resources
	TotalGPUs       int               `json:"total_gpus"`
	AvailableGPUs   int               `json:"available_gpus"`
	TotalCPUCores   int               `json:"total_cpu_cores"`
	AvailableCPUCores int             `json:"available_cpu_cores"`
	TotalMemoryMB   int64             `json:"total_memory_mb"`
	AvailableMemoryMB int64           `json:"available_memory_mb"`
	
	// Status
	Online          bool              `json:"online"`
	Schedulable     bool              `json:"schedulable"`
	DrainingMode    bool              `json:"draining_mode"`
	
	// Labels and Taints
	Labels          map[string]string `json:"labels" gorm:"serializer:json"`
	Taints          []string          `json:"taints" gorm:"serializer:json"`
	
	// Metrics
	CPUUtilization  float64           `json:"cpu_utilization"`
	MemoryUtilization float64         `json:"memory_utilization"`
	
	// Lifecycle
	LastHeartbeat   time.Time         `json:"last_heartbeat"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// IsAvailable checks if GPU is available for allocation
func (g *GPU) IsAvailable() bool {
	return !g.Allocated && 
	       g.Health == HealthHealthy && 
	       !g.ThermalThrottle &&
	       time.Since(g.CoolingPeriod) > 0
}

// NeedsCooling checks if GPU needs cooling period
func (g *GPU) NeedsCooling(threshold float64) bool {
	return g.Temperature > threshold
}

// UpdateMetrics updates GPU metrics
func (g *GPU) UpdateMetrics(util, temp, power float64) {
	g.Utilization = util
	g.Temperature = temp
	g.PowerUsage = power
	g.LastHeartbeat = time.Now()
	
	// Update max temperature
	if temp > g.MaxTemperature {
		g.MaxTemperature = temp
	}
	
	// Update health based on metrics
	g.UpdateHealth()
}

// UpdateHealth determines GPU health status
func (g *GPU) UpdateHealth() {
	switch {
	case g.Temperature > 85 || g.ErrorCount > 10:
		g.Health = HealthUnhealthy
	case g.Temperature > 75 || g.ErrorCount > 5:
		g.Health = HealthDegraded
	case g.Temperature > 65 || g.ErrorCount > 0:
		g.Health = HealthWarning
	default:
		g.Health = HealthHealthy
	}
}

// HasCapacity checks if node has capacity for the job
func (n *Node) HasCapacity(gpus, cpus int, memoryMB int64) bool {
	return n.Online &&
		n.Schedulable &&
		!n.DrainingMode &&
		n.AvailableGPUs >= gpus &&
		n.AvailableCPUCores >= cpus &&
		n.AvailableMemoryMB >= memoryMB
}
