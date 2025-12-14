package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	Environment      string           `mapstructure:"environment"`
	Scheduler        SchedulerConfig  `mapstructure:"scheduler"`
	Agent            AgentConfig      `mapstructure:"agent"`
	Database         DatabaseConfig   `mapstructure:"database"`
	Redis            RedisConfig      `mapstructure:"redis"`
	Kubernetes       KubernetesConfig `mapstructure:"kubernetes"`
	API              APIConfig        `mapstructure:"api"`
	Metrics          MetricsConfig    `mapstructure:"metrics"`
	Telemetry        TelemetryConfig  `mapstructure:"telemetry"`
}

type SchedulerConfig struct {
	SchedulingInterval   int     `mapstructure:"scheduling_interval_ms"`
	MaxQueueSize         int     `mapstructure:"max_queue_size"`
	EnablePreemption     bool    `mapstructure:"enable_preemption"`
	EnableGangScheduling bool    `mapstructure:"enable_gang_scheduling"`
	EnableThermalAware   bool    `mapstructure:"enable_thermal_aware"`
	ThermalThreshold     float64 `mapstructure:"thermal_threshold"`
	DefaultPriority      int     `mapstructure:"default_priority"`
}

type AgentConfig struct {
	NodeID              string `mapstructure:"node_id"`
	HeartbeatInterval   int    `mapstructure:"heartbeat_interval_ms"`
	MetricsInterval     int    `mapstructure:"metrics_interval_ms"`
	HealthCheckInterval int    `mapstructure:"health_check_interval_ms"`
	DCGMEnabled         bool   `mapstructure:"dcgm_enabled"`
	DCGMHostPort        string `mapstructure:"dcgm_host_port"`
	ContainerRuntime    string `mapstructure:"container_runtime"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime_mins"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type KubernetesConfig struct {
	InCluster      bool   `mapstructure:"in_cluster"`
	KubeConfigPath string `mapstructure:"kubeconfig_path"`
	Namespace      string `mapstructure:"namespace"`
}

type APIConfig struct {
	GRPCPort        int    `mapstructure:"grpc_port"`
	HTTPPort        int    `mapstructure:"http_port"`
	EnableTLS       bool   `mapstructure:"enable_tls"`
	TLSCertPath     string `mapstructure:"tls_cert_path"`
	TLSKeyPath      string `mapstructure:"tls_key_path"`
	CORSEnabled     bool   `mapstructure:"cors_enabled"`
	CORSOrigins     string `mapstructure:"cors_origins"`
	RateLimitRPS    int    `mapstructure:"rate_limit_rps"`
}

type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Port       int    `mapstructure:"port"`
	Path       string `mapstructure:"path"`
	Namespace  string `mapstructure:"namespace"`
}

type TelemetryConfig struct {
	TracingEnabled  bool   `mapstructure:"tracing_enabled"`
	JaegerEndpoint  string `mapstructure:"jaeger_endpoint"`
	ProfilingEnabled bool  `mapstructure:"profiling_enabled"`
	ProfilingPort    int   `mapstructure:"profiling_port"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/gpu-scheduler")
	}

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("GPU_SCHEDULER")

	// Set defaults
	setDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Environment
	v.SetDefault("environment", "development")

	// Scheduler
	v.SetDefault("scheduler.scheduling_interval_ms", 1000)
	v.SetDefault("scheduler.max_queue_size", 10000)
	v.SetDefault("scheduler.enable_preemption", true)
	v.SetDefault("scheduler.enable_gang_scheduling", true)
	v.SetDefault("scheduler.enable_thermal_aware", true)
	v.SetDefault("scheduler.thermal_threshold", 75.0)
	v.SetDefault("scheduler.default_priority", 100)

	// Agent
	v.SetDefault("agent.heartbeat_interval_ms", 5000)
	v.SetDefault("agent.metrics_interval_ms", 10000)
	v.SetDefault("agent.health_check_interval_ms", 30000)
	v.SetDefault("agent.dcgm_enabled", true)
	v.SetDefault("agent.dcgm_host_port", "localhost:5555")
	v.SetDefault("agent.container_runtime", "docker")

	// Database
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.database", "gpu_scheduler")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime_mins", 30)

	// Redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 2)

	// Kubernetes
	v.SetDefault("kubernetes.in_cluster", false)
	v.SetDefault("kubernetes.namespace", "gpu-system")

	// API
	v.SetDefault("api.grpc_port", 9090)
	v.SetDefault("api.http_port", 8080)
	v.SetDefault("api.enable_tls", false)
	v.SetDefault("api.cors_enabled", true)
	v.SetDefault("api.cors_origins", "*")
	v.SetDefault("api.rate_limit_rps", 100)

	// Metrics
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9091)
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.namespace", "gpu_scheduler")

	// Telemetry
	v.SetDefault("telemetry.tracing_enabled", false)
	v.SetDefault("telemetry.profiling_enabled", false)
	v.SetDefault("telemetry.profiling_port", 6060)
}
