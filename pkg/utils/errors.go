package utils

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// Resource errors
	ErrInsufficientResources   = errors.New("insufficient resources available")
	ErrGPUNotFound             = errors.New("GPU not found")
	ErrNodeNotFound            = errors.New("node not found")
	ErrNoAvailableNodes        = errors.New("no available nodes found")
	
	// Job errors
	ErrJobNotFound             = errors.New("job not found")
	ErrJobAlreadyRunning       = errors.New("job is already running")
	ErrJobAlreadyCompleted     = errors.New("job is already completed")
	ErrInvalidJobState         = errors.New("invalid job state")
	ErrJobCancelled            = errors.New("job was cancelled")
	
	// Tenant errors
	ErrTenantNotFound          = errors.New("tenant not found")
	ErrQuotaExceeded           = errors.New("tenant quota exceeded")
	ErrUnauthorized            = errors.New("unauthorized access")
	
	// Allocation errors
	ErrAllocationFailed        = errors.New("resource allocation failed")
	ErrAllocationNotFound      = errors.New("allocation not found")
	ErrGangSchedulingFailed    = errors.New("gang scheduling failed - partial allocation")
	
	// Configuration errors
	ErrInvalidConfig           = errors.New("invalid configuration")
	ErrMissingConfig           = errors.New("missing required configuration")
	
	// Database errors
	ErrDatabaseConnection      = errors.New("database connection failed")
	ErrDatabaseQuery           = errors.New("database query failed")
	
	// Kubernetes errors
	ErrKubernetesClient        = errors.New("kubernetes client error")
	ErrCRDNotFound             = errors.New("CRD not found")
)

// SchedulerError wraps errors with additional context
type SchedulerError struct {
	Op      string // Operation that failed
	Kind    string // Error kind
	Err     error  // Underlying error
	Message string // User-friendly message
}

func (e *SchedulerError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *SchedulerError) Unwrap() error {
	return e.Err
}

// NewSchedulerError creates a new scheduler error
func NewSchedulerError(op, kind string, err error, message string) *SchedulerError {
	return &SchedulerError{
		Op:      op,
		Kind:    kind,
		Err:     err,
		Message: message,
	}
}

// InsufficientResourcesError represents resource shortage
type InsufficientResourcesError struct {
	Requested map[string]int
	Available map[string]int
	Message   string
}

func (e *InsufficientResourcesError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("insufficient resources: requested=%v, available=%v", e.Requested, e.Available)
}

// QuotaExceededError represents quota violation
type QuotaExceededError struct {
	TenantID  string
	Resource  string
	Requested int
	Quota     int
	Current   int
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("quota exceeded for tenant %s: %s requested=%d, quota=%d, current=%d",
		e.TenantID, e.Resource, e.Requested, e.Quota, e.Current)
}

// JobStateError represents invalid job state transition
type JobStateError struct {
	JobID        string
	CurrentState string
	TargetState  string
	Message      string
}

func (e *JobStateError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("invalid state transition for job %s: %s -> %s",
		e.JobID, e.CurrentState, e.TargetState)
}

// IsNotFound checks if error is a not-found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrJobNotFound) ||
		errors.Is(err, ErrTenantNotFound) ||
		errors.Is(err, ErrGPUNotFound) ||
		errors.Is(err, ErrNodeNotFound) ||
		errors.Is(err, ErrAllocationNotFound)
}

// IsQuotaExceeded checks if error is quota-related
func IsQuotaExceeded(err error) bool {
	var qErr *QuotaExceededError
	return errors.As(err, &qErr) || errors.Is(err, ErrQuotaExceeded)
}

// IsResourceError checks if error is resource-related
func IsResourceError(err error) bool {
	var rErr *InsufficientResourcesError
	return errors.As(err, &rErr) || errors.Is(err, ErrInsufficientResources)
}

// WrapError wraps an error with operation context
func WrapError(op string, err error, message string) error {
	if err == nil {
		return nil
	}
	return NewSchedulerError(op, "error", err, message)
}
