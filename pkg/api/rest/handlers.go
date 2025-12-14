package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/scheduler/core"
	"github.com/azizbahloul/gpu-scheduler/pkg/storage"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Handlers holds HTTP handlers
type Handlers struct {
	scheduler *core.Scheduler
	storage   storage.Repository
}

// NewHandlers creates new HTTP handlers
func NewHandlers(scheduler *core.Scheduler, storage storage.Repository) *Handlers {
	return &Handlers{
		scheduler: scheduler,
		storage:   storage,
	}
}

// SubmitJobHandler handles job submissions
func (h *Handlers) SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID         string            `json:"tenant_id"`
		Name             string            `json:"name"`
		Priority         int               `json:"priority"`
		GPUCount         int               `json:"gpu_count"`
		GPUMemoryMB      int64             `json:"gpu_memory_mb"`
		CPUCores         int               `json:"cpu_cores"`
		MemoryMB         int64             `json:"memory_mb"`
		Script           string            `json:"script"`
		Environment      map[string]string `json:"environment"`
		Image            string            `json:"image"`
		Command          []string          `json:"command"`
		Args             []string          `json:"args"`
		GangScheduling   bool              `json:"gang_scheduling"`
		MaxRuntimeMinutes int              `json:"max_runtime_minutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job := &models.Job{
		ID:          generateJobID(),
		TenantID:    req.TenantID,
		Name:        req.Name,
		Priority:    req.Priority,
		GPUCount:    req.GPUCount,
		GPUMemoryMB: req.GPUMemoryMB,
		CPUCores:    req.CPUCores,
		MemoryMB:    req.MemoryMB,
		Script:      req.Script,
		Environment: req.Environment,
		Image:       req.Image,
		Command:     req.Command,
		Args:        req.Args,
		GangScheduling: req.GangScheduling,
		MaxRuntime:  time.Duration(req.MaxRuntimeMinutes) * time.Minute,
	}

	if err := h.scheduler.SubmitJob(r.Context(), job); err != nil {
		utils.Error("Failed to submit job", zap.Error(err))
		
		if utils.IsQuotaExceeded(err) {
			respondJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to submit job"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"job_id":  job.ID,
		"status":  "submitted",
		"message": "Job submitted successfully",
	})
}

// GetJobStatusHandler returns job status
func (h *Handlers) GetJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")

	status, err := h.scheduler.GetJobStatus(r.Context(), jobID)
	if err != nil {
		if utils.IsNotFound(err) {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get job status", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, status)
}

// ListJobsHandler lists jobs
func (h *Handlers) ListJobsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	state := r.URL.Query().Get("state")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var jobs []*models.Job
	var err error

	if tenantID != "" {
		jobs, err = h.storage.ListJobsByTenant(r.Context(), tenantID)
	} else if state != "" {
		jobs, err = h.storage.ListJobsByState(r.Context(), models.JobState(state))
	} else {
		jobs, err = h.storage.ListJobs(r.Context(), limit, offset)
	}

	if err != nil {
		http.Error(w, "Failed to list jobs", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

// CancelJobHandler cancels a job
func (h *Handlers) CancelJobHandler(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")

	if err := h.scheduler.CancelJob(r.Context(), jobID); err != nil {
		if utils.IsNotFound(err) {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to cancel job", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Job cancelled successfully"})
}

// GetClusterStatusHandler returns cluster status
func (h *Handlers) GetClusterStatusHandler(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.storage.ListNodes(r.Context())
	if err != nil {
		http.Error(w, "Failed to get cluster status", http.StatusInternalServerError)
		return
	}

	totalGPUs := 0
	availableGPUs := 0
	onlineNodes := 0

	for _, node := range nodes {
		totalGPUs += node.TotalGPUs
		availableGPUs += node.AvailableGPUs
		if node.Online {
			onlineNodes++
		}
	}

	allJobs, _ := h.storage.ListJobs(r.Context(), 10000, 0)
	pendingCount := 0
	runningCount := 0

	for _, job := range allJobs {
		if job.State == models.JobStatePending {
			pendingCount++
		} else if job.State == models.JobStateRunning {
			runningCount++
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_gpus":      totalGPUs,
		"available_gpus":  availableGPUs,
		"total_nodes":     len(nodes),
		"online_nodes":    onlineNodes,
		"total_jobs":      len(allJobs),
		"pending_jobs":    pendingCount,
		"running_jobs":    runningCount,
	})
}

// CreateTenantHandler creates a new tenant
func (h *Handlers) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
	var tenant models.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenant.ID = generateTenantID()
	tenant.Active = true
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	if err := h.storage.CreateTenant(r.Context(), &tenant); err != nil {
		http.Error(w, "Failed to create tenant", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, tenant)
}

// HealthCheckHandler returns health status
func (h *Handlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.storage.Ping(r.Context()); err != nil {
		http.Error(w, "Unhealthy", http.StatusServiceUnavailable)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func generateJobID() string {
	return fmt.Sprintf("job-%d", time.Now().UnixNano())
}

func generateTenantID() string {
	return fmt.Sprintf("tenant-%d", time.Now().UnixNano())
}
