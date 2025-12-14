package models

import (
	"time"
)

// JobState represents the current state of a job
type JobState string

const (
	JobStatePending    JobState = "pending"
	JobStateRunning    JobState = "running"
	JobStateCompleted  JobState = "completed"
	JobStateFailed     JobState = "failed"
	JobStatePreempted  JobState = "preempted"
	JobStateCancelled  JobState = "cancelled"
)

// Job represents a GPU job submitted by a tenant
type Job struct {
	ID                string            `json:"id" gorm:"primaryKey"`
	TenantID          string            `json:"tenant_id" gorm:"index"`
	Name              string            `json:"name"`
	State             JobState          `json:"state" gorm:"index"`
	Priority          int               `json:"priority"`
	GPUCount          int               `json:"gpu_count"`
	GPUMemoryMB       int64             `json:"gpu_memory_mb"`
	CPUCores          int               `json:"cpu_cores"`
	MemoryMB          int64             `json:"memory_mb"`
	Script            string            `json:"script" gorm:"type:text"`
	Environment       map[string]string `json:"environment" gorm:"serializer:json"`
	Image             string            `json:"image"`
	Command           []string          `json:"command" gorm:"serializer:json"`
	Args              []string          `json:"args" gorm:"serializer:json"`
	GangScheduling    bool              `json:"gang_scheduling"`
	MaxRuntime        time.Duration     `json:"max_runtime"`
	CheckpointEnabled bool              `json:"checkpoint_enabled"`
	CheckpointPath    string            `json:"checkpoint_path"`
	
	// Timestamps
	SubmittedAt       time.Time         `json:"submitted_at"`
	ScheduledAt       *time.Time        `json:"scheduled_at"`
	StartedAt         *time.Time        `json:"started_at"`
	CompletedAt       *time.Time        `json:"completed_at"`
	
	// Prediction
	EstimatedDuration time.Duration     `json:"estimated_duration"`
	PredictionConf    float64           `json:"prediction_confidence"`
	
	// Metrics
	ActualDuration    time.Duration     `json:"actual_duration"`
	GPUUtilization    float64           `json:"gpu_utilization"`
	PreemptedCount    int               `json:"preempted_count"`
	
	// Metadata
	Labels            map[string]string `json:"labels" gorm:"serializer:json"`
	Annotations       map[string]string `json:"annotations" gorm:"serializer:json"`
	
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// JobMetadata contains extracted features for ML prediction
type JobMetadata struct {
	ModelType       string  `json:"model_type"`
	BatchSize       int     `json:"batch_size"`
	DatasetSize     int64   `json:"dataset_size"`
	Framework       string  `json:"framework"`
	TensorCores     bool    `json:"tensor_cores"`
	MixedPrecision  bool    `json:"mixed_precision"`
}

// JobStatus provides detailed status information
type JobStatus struct {
	JobID           string            `json:"job_id"`
	State           JobState          `json:"state"`
	Message         string            `json:"message"`
	AllocatedGPUs   []string          `json:"allocated_gpus"`
	NodeName        string            `json:"node_name"`
	QueuePosition   int               `json:"queue_position"`
	EstimatedWait   time.Duration     `json:"estimated_wait"`
	Logs            string            `json:"logs"`
	Metrics         map[string]float64 `json:"metrics"`
}

// IsTerminal returns true if the job is in a terminal state
func (j *Job) IsTerminal() bool {
	return j.State == JobStateCompleted || 
	       j.State == JobStateFailed || 
	       j.State == JobStateCancelled
}

// IsActive returns true if the job is active (running or pending)
func (j *Job) IsActive() bool {
	return j.State == JobStateRunning || j.State == JobStatePending
}

// CalculateActualDuration updates the actual duration
func (j *Job) CalculateActualDuration() {
	if j.StartedAt != nil && j.CompletedAt != nil {
		j.ActualDuration = j.CompletedAt.Sub(*j.StartedAt)
	}
}
