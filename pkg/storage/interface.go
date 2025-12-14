package storage

import (
	"context"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
)

// Repository defines the storage interface
type Repository interface {
	// Job operations
	CreateJob(ctx context.Context, job *models.Job) error
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	UpdateJob(ctx context.Context, job *models.Job) error
	DeleteJob(ctx context.Context, jobID string) error
	ListJobs(ctx context.Context, limit, offset int) ([]*models.Job, error)
	ListJobsByTenant(ctx context.Context, tenantID string) ([]*models.Job, error)
	ListJobsByState(ctx context.Context, state models.JobState) ([]*models.Job, error)

	// Tenant operations
	CreateTenant(ctx context.Context, tenant *models.Tenant) error
	GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *models.Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
	ListTenants(ctx context.Context) ([]*models.Tenant, error)

	// GPU operations
	CreateGPU(ctx context.Context, gpu *models.GPU) error
	GetGPU(ctx context.Context, gpuID string) (*models.GPU, error)
	UpdateGPU(ctx context.Context, gpu *models.GPU) error
	DeleteGPU(ctx context.Context, gpuID string) error
	ListGPUs(ctx context.Context) ([]*models.GPU, error)
	ListGPUsByNode(ctx context.Context, nodeID string) ([]*models.GPU, error)
	ListAvailableGPUs(ctx context.Context) ([]*models.GPU, error)

	// Node operations
	CreateNode(ctx context.Context, node *models.Node) error
	GetNode(ctx context.Context, nodeID string) (*models.Node, error)
	UpdateNode(ctx context.Context, node *models.Node) error
	DeleteNode(ctx context.Context, nodeID string) error
	ListNodes(ctx context.Context) ([]*models.Node, error)

	// Allocation operations
	CreateAllocation(ctx context.Context, allocation *models.Allocation) error
	GetAllocation(ctx context.Context, allocationID string) (*models.Allocation, error)
	UpdateAllocation(ctx context.Context, allocation *models.Allocation) error
	DeleteAllocation(ctx context.Context, allocationID string) error
	GetJobAllocations(ctx context.Context, jobID string) ([]*models.Allocation, error)
	ListActiveAllocations(ctx context.Context) ([]*models.Allocation, error)

	// Health check
	Ping(ctx context.Context) error
	Close() error
}
