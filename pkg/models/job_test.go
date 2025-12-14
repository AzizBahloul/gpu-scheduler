package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobIsTerminal(t *testing.T) {
	tests := []struct {
		name     string
		state    JobState
		expected bool
	}{
		{"Completed is terminal", JobStateCompleted, true},
		{"Failed is terminal", JobStateFailed, true},
		{"Cancelled is terminal", JobStateCancelled, true},
		{"Pending is not terminal", JobStatePending, false},
		{"Running is not terminal", JobStateRunning, false},
		{"Preempted is not terminal", JobStatePreempted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{State: tt.state}
			assert.Equal(t, tt.expected, job.IsTerminal())
		})
	}
}

func TestJobIsActive(t *testing.T) {
	tests := []struct {
		name     string
		state    JobState
		expected bool
	}{
		{"Running is active", JobStateRunning, true},
		{"Pending is active", JobStatePending, true},
		{"Completed is not active", JobStateCompleted, false},
		{"Failed is not active", JobStateFailed, false},
		{"Cancelled is not active", JobStateCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{State: tt.state}
			assert.Equal(t, tt.expected, job.IsActive())
		})
	}
}

func TestCalculateActualDuration(t *testing.T) {
	start := time.Now()
	end := start.Add(2 * time.Hour)

	job := &Job{
		StartedAt:   &start,
		CompletedAt: &end,
	}

	job.CalculateActualDuration()
	assert.Equal(t, 2*time.Hour, job.ActualDuration)
}

func TestCalculateActualDurationNilTimes(t *testing.T) {
	job := &Job{}
	job.CalculateActualDuration()
	assert.Equal(t, time.Duration(0), job.ActualDuration)
}
