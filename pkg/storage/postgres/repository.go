package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/azizbahloul/gpu-scheduler/pkg/storage"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(config *utils.DatabaseConfig) (storage.Repository, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Minute)

	// Auto-migrate schemas
	if err := db.AutoMigrate(
		&models.Job{},
		&models.Tenant{},
		&models.GPU{},
		&models.Node{},
		&models.Allocation{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Job operations
func (r *PostgresRepository) CreateJob(ctx context.Context, job *models.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *PostgresRepository) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	var job models.Job
	if err := r.db.WithContext(ctx).First(&job, "id = ?", jobID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrJobNotFound
		}
		return nil, err
	}
	return &job, nil
}

func (r *PostgresRepository) UpdateJob(ctx context.Context, job *models.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *PostgresRepository) DeleteJob(ctx context.Context, jobID string) error {
	return r.db.WithContext(ctx).Delete(&models.Job{}, "id = ?", jobID).Error
}

func (r *PostgresRepository) ListJobs(ctx context.Context, limit, offset int) ([]*models.Job, error) {
	var jobs []*models.Job
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("submitted_at DESC").Find(&jobs).Error
	return jobs, err
}

func (r *PostgresRepository) ListJobsByTenant(ctx context.Context, tenantID string) ([]*models.Job, error) {
	var jobs []*models.Job
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("submitted_at DESC").Find(&jobs).Error
	return jobs, err
}

func (r *PostgresRepository) ListJobsByState(ctx context.Context, state models.JobState) ([]*models.Job, error) {
	var jobs []*models.Job
	err := r.db.WithContext(ctx).Where("state = ?", state).Find(&jobs).Error
	return jobs, err
}

// Tenant operations
func (r *PostgresRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *PostgresRepository) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, "id = ?", tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *PostgresRepository) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *PostgresRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	return r.db.WithContext(ctx).Delete(&models.Tenant{}, "id = ?", tenantID).Error
}

func (r *PostgresRepository) ListTenants(ctx context.Context) ([]*models.Tenant, error) {
	var tenants []*models.Tenant
	err := r.db.WithContext(ctx).Find(&tenants).Error
	return tenants, err
}

// GPU operations
func (r *PostgresRepository) CreateGPU(ctx context.Context, gpu *models.GPU) error {
	return r.db.WithContext(ctx).Create(gpu).Error
}

func (r *PostgresRepository) GetGPU(ctx context.Context, gpuID string) (*models.GPU, error) {
	var gpu models.GPU
	if err := r.db.WithContext(ctx).First(&gpu, "id = ?", gpuID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrGPUNotFound
		}
		return nil, err
	}
	return &gpu, nil
}

func (r *PostgresRepository) UpdateGPU(ctx context.Context, gpu *models.GPU) error {
	return r.db.WithContext(ctx).Save(gpu).Error
}

func (r *PostgresRepository) DeleteGPU(ctx context.Context, gpuID string) error {
	return r.db.WithContext(ctx).Delete(&models.GPU{}, "id = ?", gpuID).Error
}

func (r *PostgresRepository) ListGPUs(ctx context.Context) ([]*models.GPU, error) {
	var gpus []*models.GPU
	err := r.db.WithContext(ctx).Find(&gpus).Error
	return gpus, err
}

func (r *PostgresRepository) ListGPUsByNode(ctx context.Context, nodeID string) ([]*models.GPU, error) {
	var gpus []*models.GPU
	err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).Find(&gpus).Error
	return gpus, err
}

func (r *PostgresRepository) ListAvailableGPUs(ctx context.Context) ([]*models.GPU, error) {
	var gpus []*models.GPU
	err := r.db.WithContext(ctx).Where("allocated = ?", false).Where("health = ?", models.HealthHealthy).Find(&gpus).Error
	return gpus, err
}

// Node operations
func (r *PostgresRepository) CreateNode(ctx context.Context, node *models.Node) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *PostgresRepository) GetNode(ctx context.Context, nodeID string) (*models.Node, error) {
	var node models.Node
	if err := r.db.WithContext(ctx).First(&node, "id = ?", nodeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrNodeNotFound
		}
		return nil, err
	}
	return &node, nil
}

func (r *PostgresRepository) UpdateNode(ctx context.Context, node *models.Node) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *PostgresRepository) DeleteNode(ctx context.Context, nodeID string) error {
	return r.db.WithContext(ctx).Delete(&models.Node{}, "id = ?", nodeID).Error
}

func (r *PostgresRepository) ListNodes(ctx context.Context) ([]*models.Node, error) {
	var nodes []*models.Node
	err := r.db.WithContext(ctx).Where("online = ?", true).Find(&nodes).Error
	return nodes, err
}

// Allocation operations
func (r *PostgresRepository) CreateAllocation(ctx context.Context, allocation *models.Allocation) error {
	return r.db.WithContext(ctx).Create(allocation).Error
}

func (r *PostgresRepository) GetAllocation(ctx context.Context, allocationID string) (*models.Allocation, error) {
	var allocation models.Allocation
	if err := r.db.WithContext(ctx).First(&allocation, "id = ?", allocationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrAllocationNotFound
		}
		return nil, err
	}
	return &allocation, nil
}

func (r *PostgresRepository) UpdateAllocation(ctx context.Context, allocation *models.Allocation) error {
	return r.db.WithContext(ctx).Save(allocation).Error
}

func (r *PostgresRepository) DeleteAllocation(ctx context.Context, allocationID string) error {
	return r.db.WithContext(ctx).Delete(&models.Allocation{}, "id = ?", allocationID).Error
}

func (r *PostgresRepository) GetJobAllocations(ctx context.Context, jobID string) ([]*models.Allocation, error) {
	var allocations []*models.Allocation
	err := r.db.WithContext(ctx).Where("job_id = ?", jobID).Find(&allocations).Error
	return allocations, err
}

func (r *PostgresRepository) ListActiveAllocations(ctx context.Context) ([]*models.Allocation, error) {
	var allocations []*models.Allocation
	err := r.db.WithContext(ctx).Where("state = ?", models.AllocationActive).Find(&allocations).Error
	return allocations, err
}

// Health check
func (r *PostgresRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (r *PostgresRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
