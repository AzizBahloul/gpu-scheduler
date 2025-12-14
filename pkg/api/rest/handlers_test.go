package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/scheduler/core"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage implements storage.Repository for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreateJob(ctx context.Context, job *models.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockStorage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockStorage) UpdateJob(ctx context.Context, job *models.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockStorage) DeleteJob(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockStorage) ListJobs(ctx context.Context, limit, offset int) ([]*models.Job, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Job), args.Error(1)
}

func (m *MockStorage) ListJobsByTenant(ctx context.Context, tenantID string) ([]*models.Job, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*models.Job), args.Error(1)
}

func (m *MockStorage) ListJobsByState(ctx context.Context, state models.JobState) ([]*models.Job, error) {
	args := m.Called(ctx, state)
	return args.Get(0).([]*models.Job), args.Error(1)
}

func (m *MockStorage) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockStorage) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockStorage) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockStorage) DeleteTenant(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockStorage) ListTenants(ctx context.Context) ([]*models.Tenant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Tenant), args.Error(1)
}

func (m *MockStorage) CreateGPU(ctx context.Context, gpu *models.GPU) error {
	return nil
}

func (m *MockStorage) GetGPU(ctx context.Context, gpuID string) (*models.GPU, error) {
	return nil, nil
}

func (m *MockStorage) UpdateGPU(ctx context.Context, gpu *models.GPU) error {
	return nil
}

func (m *MockStorage) DeleteGPU(ctx context.Context, gpuID string) error {
	return nil
}

func (m *MockStorage) ListGPUs(ctx context.Context) ([]*models.GPU, error) {
	return []*models.GPU{}, nil
}

func (m *MockStorage) ListGPUsByNode(ctx context.Context, nodeID string) ([]*models.GPU, error) {
	return []*models.GPU{}, nil
}

func (m *MockStorage) ListAvailableGPUs(ctx context.Context) ([]*models.GPU, error) {
	return []*models.GPU{}, nil
}

func (m *MockStorage) CreateNode(ctx context.Context, node *models.Node) error {
	return nil
}

func (m *MockStorage) GetNode(ctx context.Context, nodeID string) (*models.Node, error) {
	return nil, nil
}

func (m *MockStorage) UpdateNode(ctx context.Context, node *models.Node) error {
	return nil
}

func (m *MockStorage) DeleteNode(ctx context.Context, nodeID string) error {
	return nil
}

func (m *MockStorage) ListNodes(ctx context.Context) ([]*models.Node, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Node), args.Error(1)
}

func (m *MockStorage) CreateAllocation(ctx context.Context, allocation *models.Allocation) error {
	return nil
}

func (m *MockStorage) GetAllocation(ctx context.Context, allocationID string) (*models.Allocation, error) {
	return nil, nil
}

func (m *MockStorage) UpdateAllocation(ctx context.Context, allocation *models.Allocation) error {
	return nil
}

func (m *MockStorage) DeleteAllocation(ctx context.Context, allocationID string) error {
	return nil
}

func (m *MockStorage) GetJobAllocations(ctx context.Context, jobID string) ([]*models.Allocation, error) {
	return []*models.Allocation{}, nil
}

func (m *MockStorage) ListActiveAllocations(ctx context.Context) ([]*models.Allocation, error) {
	return []*models.Allocation{}, nil
}

func (m *MockStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestHealthCheckHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	handlers := NewHandlers(nil, mockStorage)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthCheckHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
}

func TestSubmitJobHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	
	// Create scheduler with mock storage
	config := &utils.SchedulerConfig{
		SchedulingInterval: 1000,
		MaxQueueSize:      100,
	}
	scheduler := core.NewScheduler(config, mockStorage)

	handlers := NewHandlers(scheduler, mockStorage)

	// Setup mock expectations
	tenant := &models.Tenant{
		ID:                "tenant-1",
		MaxGPUs:           10,
		MaxGPUMemoryMB:    160000,
		MaxCPUCores:       64,
		MaxMemoryMB:       256000,
		MaxConcurrentJobs: 20,
		CurrentGPUs:       0,
		CurrentGPUMemory:  0,
		CurrentCPUCores:   0,
		CurrentMemory:     0,
		CurrentJobs:       0,
		Active:            true,
	}
	mockStorage.On("GetTenant", mock.Anything, "tenant-1").Return(tenant, nil)
	mockStorage.On("CreateJob", mock.Anything, mock.AnythingOfType("*models.Job")).Return(nil)

	requestBody := map[string]interface{}{
		"tenant_id":    "tenant-1",
		"name":         "test-job",
		"priority":     100,
		"gpu_count":    1,
		"gpu_memory_mb": 16000,
		"cpu_cores":    4,
		"memory_mb":    32000,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.SubmitJobHandler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["job_id"])
	assert.Equal(t, "submitted", response["status"])
}

func TestGetJobStatusHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	
	config := &utils.SchedulerConfig{
		SchedulingInterval: 1000,
		MaxQueueSize:      100,
	}
	scheduler := core.NewScheduler(config, mockStorage)
	handlers := NewHandlers(scheduler, mockStorage)

	job := &models.Job{
		ID:       "job-123",
		Name:     "test-job",
		State:    models.JobStatePending,
		Priority: 100,
	}

	mockStorage.On("GetJob", mock.Anything, "job-123").Return(job, nil)

	req := httptest.NewRequest("GET", "/api/v1/jobs/job-123", nil)
	w := httptest.NewRecorder()

	// Use chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("jobID", "job-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handlers.GetJobStatusHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListJobsHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	scheduler := core.NewScheduler(&utils.SchedulerConfig{}, mockStorage)
	handlers := NewHandlers(scheduler, mockStorage)

	jobs := []*models.Job{
		{ID: "job-1", Name: "job1", State: models.JobStatePending},
		{ID: "job-2", Name: "job2", State: models.JobStateRunning},
	}

	mockStorage.On("ListJobs", mock.Anything, 50, 0).Return(jobs, nil)

	req := httptest.NewRequest("GET", "/api/v1/jobs", nil)
	w := httptest.NewRecorder()

	handlers.ListJobsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(2), response["total"])
}

func TestGetClusterStatusHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	scheduler := core.NewScheduler(&utils.SchedulerConfig{}, mockStorage)
	handlers := NewHandlers(scheduler, mockStorage)

	nodes := []*models.Node{
		{
			ID:            "node-1",
			TotalGPUs:     8,
			AvailableGPUs: 5,
			Online:        true,
		},
		{
			ID:            "node-2",
			TotalGPUs:     8,
			AvailableGPUs: 3,
			Online:        true,
		},
	}

	mockStorage.On("ListNodes", mock.Anything).Return(nodes, nil)
	mockStorage.On("ListJobs", mock.Anything, 10000, 0).Return([]*models.Job{}, nil)

	req := httptest.NewRequest("GET", "/api/v1/cluster/status", nil)
	w := httptest.NewRecorder()

	handlers.GetClusterStatusHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(16), response["total_gpus"])
	assert.Equal(t, float64(8), response["available_gpus"])
	assert.Equal(t, float64(2), response["total_nodes"])
}

func TestCreateTenantHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	scheduler := core.NewScheduler(&utils.SchedulerConfig{}, mockStorage)
	handlers := NewHandlers(scheduler, mockStorage)

	mockStorage.On("CreateTenant", mock.Anything, mock.AnythingOfType("*models.Tenant")).Return(nil)

	requestBody := map[string]interface{}{
		"name":     "test-tenant",
		"max_gpus": 10,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateTenantHandler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
