package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Coordinator CoordinatorConfig `mapstructure:"coordinator"`
	Worker      WorkerConfig      `mapstructure:"worker"`
	Target      TargetConfig      `mapstructure:"target"`
	Workload    WorkloadConfig    `mapstructure:"workload"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
}

// CoordinatorConfig contains coordinator-specific settings
type CoordinatorConfig struct {
	Port         int           `mapstructure:"port"`
	GRPCPort     int           `mapstructure:"grpc_port"`
	RedisURL     string        `mapstructure:"redis_url"`
	HeartbeatTimeout time.Duration `mapstructure:"heartbeat_timeout"`
}

// WorkerConfig contains worker-specific settings
type WorkerConfig struct {
	NodeID         string        `mapstructure:"node_id"`
	CoordinatorURL string        `mapstructure:"coordinator_url"`
	Port           int           `mapstructure:"port"`
	MaxConcurrent  int           `mapstructure:"max_concurrent"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
}

// TargetConfig contains target system configuration
type TargetConfig struct {
	URL     string            `mapstructure:"url"`
	Auth    AuthConfig        `mapstructure:"auth"`
	Timeout time.Duration     `mapstructure:"timeout"`
	Headers map[string]string `mapstructure:"headers"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Type   string `mapstructure:"type"`   // "bearer", "basic", "none"
	Token  string `mapstructure:"token"`
	User   string `mapstructure:"user"`
	Pass   string `mapstructure:"pass"`
}

// WorkloadConfig contains benchmark workload settings
type WorkloadConfig struct {
	TotalConcurrentUsers int           `mapstructure:"total_concurrent_users"`
	Duration             time.Duration `mapstructure:"duration"`
	RampUpTime           time.Duration `mapstructure:"ramp_up_time"`
	Queries              []QueryConfig `mapstructure:"queries"`
}

// QueryConfig represents a query template configuration
type QueryConfig struct {
	Name     string `mapstructure:"name"`
	Template string `mapstructure:"template"`
	Weight   int    `mapstructure:"weight"`
	Category string `mapstructure:"category"`
}

// BenchmarkConfig represents the complete benchmark configuration
type BenchmarkConfig struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   ConfigMetadata `yaml:"metadata"`
	Spec       BenchmarkSpec  `yaml:"spec"`
}

// ConfigMetadata contains metadata for the benchmark configuration
type ConfigMetadata struct {
	Name string `yaml:"name"`
}

// BenchmarkSpec contains the benchmark specification
type BenchmarkSpec struct {
	Target    TargetConfig    `yaml:"target"`
	Workload  WorkloadConfig  `yaml:"workload"`
	Workers   WorkersConfig   `yaml:"workers"`
	Reporting ReportingConfig `yaml:"reporting"`
}

// WorkersConfig contains worker node configuration
type WorkersConfig struct {
	Replicas  int               `yaml:"replicas"`
	Resources WorkerResources   `yaml:"resources"`
}

// WorkerResources contains resource specifications for workers
type WorkerResources struct {
	Requests ResourceSpec `yaml:"requests"`
	Limits   ResourceSpec `yaml:"limits"`
}

// ResourceSpec contains CPU and memory specifications
type ResourceSpec struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// ReportingConfig contains result reporting configuration
type ReportingConfig struct {
	Formats []string      `yaml:"formats"`
	Storage StorageConfig `yaml:"storage"`
}

// StorageConfig contains storage configuration for results
type StorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// MonitoringConfig contains monitoring and observability settings
type MonitoringConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	MetricsPort    int    `mapstructure:"metrics_port"`
	PrometheusPath string `mapstructure:"prometheus_path"`
}

// LoadConfig loads configuration from file
func LoadConfig(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	
	// Set default values
	setDefaults()
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	return &config, nil
}

// LoadBenchmarkConfig loads a benchmark configuration from YAML file
func LoadBenchmarkConfig(configFile string) (*BenchmarkConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read benchmark config file: %w", err)
	}
	
	var config BenchmarkConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal benchmark config: %w", err)
	}
	
	// Validate benchmark configuration
	if err := validateBenchmarkConfig(&config); err != nil {
		return nil, fmt.Errorf("benchmark configuration validation failed: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Coordinator defaults
	viper.SetDefault("coordinator.port", 8080)
	viper.SetDefault("coordinator.grpc_port", 8090)
	viper.SetDefault("coordinator.heartbeat_timeout", "30s")
	
	// Worker defaults
	viper.SetDefault("worker.port", 8080)
	viper.SetDefault("worker.max_concurrent", 100)
	viper.SetDefault("worker.heartbeat_interval", "10s")
	
	// Target defaults
	viper.SetDefault("target.timeout", "30s")
	viper.SetDefault("target.auth.type", "none")
	
	// Workload defaults
	viper.SetDefault("workload.total_concurrent_users", 100)
	viper.SetDefault("workload.duration", "300s")
	viper.SetDefault("workload.ramp_up_time", "60s")
	
	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_port", 9090)
	viper.SetDefault("monitoring.prometheus_path", "/metrics")
}/
/ validateConfig validates the application configuration
func validateConfig(config *Config) error {
	// Validate coordinator configuration
	if config.Coordinator.Port <= 0 || config.Coordinator.Port > 65535 {
		return fmt.Errorf("invalid coordinator port: %d", config.Coordinator.Port)
	}
	if config.Coordinator.GRPCPort <= 0 || config.Coordinator.GRPCPort > 65535 {
		return fmt.Errorf("invalid coordinator GRPC port: %d", config.Coordinator.GRPCPort)
	}
	if config.Coordinator.HeartbeatTimeout <= 0 {
		return fmt.Errorf("heartbeat timeout must be positive")
	}

	// Validate worker configuration
	if config.Worker.Port <= 0 || config.Worker.Port > 65535 {
		return fmt.Errorf("invalid worker port: %d", config.Worker.Port)
	}
	if config.Worker.MaxConcurrent <= 0 {
		return fmt.Errorf("max concurrent must be positive")
	}
	if config.Worker.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat interval must be positive")
	}

	// Validate target configuration
	if config.Target.URL == "" {
		return fmt.Errorf("target URL is required")
	}
	if !strings.HasPrefix(config.Target.URL, "http://") && !strings.HasPrefix(config.Target.URL, "https://") {
		return fmt.Errorf("target URL must start with http:// or https://")
	}
	if config.Target.Timeout <= 0 {
		return fmt.Errorf("target timeout must be positive")
	}

	// Validate auth configuration
	if err := validateAuthConfig(&config.Target.Auth); err != nil {
		return fmt.Errorf("invalid auth config: %w", err)
	}

	// Validate workload configuration
	if config.Workload.TotalConcurrentUsers <= 0 {
		return fmt.Errorf("total concurrent users must be positive")
	}
	if config.Workload.Duration <= 0 {
		return fmt.Errorf("workload duration must be positive")
	}
	if config.Workload.RampUpTime < 0 {
		return fmt.Errorf("ramp up time cannot be negative")
	}

	// Validate queries
	if len(config.Workload.Queries) == 0 {
		return fmt.Errorf("at least one query must be configured")
	}
	for i, query := range config.Workload.Queries {
		if err := validateQueryConfig(&query); err != nil {
			return fmt.Errorf("invalid query config at index %d: %w", i, err)
		}
	}

	return nil
}

// validateBenchmarkConfig validates the benchmark configuration
func validateBenchmarkConfig(config *BenchmarkConfig) error {
	// Validate API version and kind
	if config.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if config.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if config.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	// Validate target configuration
	if config.Spec.Target.URL == "" {
		return fmt.Errorf("spec.target.url is required")
	}
	if !strings.HasPrefix(config.Spec.Target.URL, "http://") && !strings.HasPrefix(config.Spec.Target.URL, "https://") {
		return fmt.Errorf("spec.target.url must start with http:// or https://")
	}

	// Validate auth configuration
	if err := validateAuthConfig(&config.Spec.Target.Auth); err != nil {
		return fmt.Errorf("invalid spec.target.auth: %w", err)
	}

	// Validate workload configuration
	if config.Spec.Workload.TotalConcurrentUsers <= 0 {
		return fmt.Errorf("spec.workload.total_concurrent_users must be positive")
	}
	if config.Spec.Workload.Duration <= 0 {
		return fmt.Errorf("spec.workload.duration must be positive")
	}

	// Validate queries
	if len(config.Spec.Workload.Queries) == 0 {
		return fmt.Errorf("at least one query must be configured in spec.workload.queries")
	}
	for i, query := range config.Spec.Workload.Queries {
		if err := validateQueryConfig(&query); err != nil {
			return fmt.Errorf("invalid query config at spec.workload.queries[%d]: %w", i, err)
		}
	}

	// Validate workers configuration
	if config.Spec.Workers.Replicas <= 0 {
		return fmt.Errorf("spec.workers.replicas must be positive")
	}

	// Validate resource specifications
	if err := validateResourceSpec(&config.Spec.Workers.Resources.Requests, "requests"); err != nil {
		return fmt.Errorf("invalid spec.workers.resources.requests: %w", err)
	}
	if err := validateResourceSpec(&config.Spec.Workers.Resources.Limits, "limits"); err != nil {
		return fmt.Errorf("invalid spec.workers.resources.limits: %w", err)
	}

	// Validate reporting configuration
	if len(config.Spec.Reporting.Formats) == 0 {
		return fmt.Errorf("at least one reporting format must be specified")
	}
	for _, format := range config.Spec.Reporting.Formats {
		if !isValidReportingFormat(format) {
			return fmt.Errorf("invalid reporting format: %s", format)
		}
	}

	return nil
}

// validateAuthConfig validates authentication configuration
func validateAuthConfig(auth *AuthConfig) error {
	switch auth.Type {
	case "none":
		// No additional validation needed
	case "bearer":
		if auth.Token == "" {
			return fmt.Errorf("token is required for bearer auth")
		}
	case "basic":
		if auth.User == "" || auth.Pass == "" {
			return fmt.Errorf("user and pass are required for basic auth")
		}
	default:
		return fmt.Errorf("unsupported auth type: %s", auth.Type)
	}
	return nil
}

// validateQueryConfig validates query configuration
func validateQueryConfig(query *QueryConfig) error {
	if query.Name == "" {
		return fmt.Errorf("query name is required")
	}
	if query.Template == "" {
		return fmt.Errorf("query template is required")
	}
	if query.Weight <= 0 {
		return fmt.Errorf("query weight must be positive")
	}
	if query.Category == "" {
		return fmt.Errorf("query category is required")
	}
	return nil
}

// validateResourceSpec validates Kubernetes resource specifications
func validateResourceSpec(spec *ResourceSpec, specType string) error {
	if spec.CPU == "" {
		return fmt.Errorf("CPU %s is required", specType)
	}
	if spec.Memory == "" {
		return fmt.Errorf("Memory %s is required", specType)
	}
	
	// Basic validation for Kubernetes resource format
	if !isValidKubernetesResource(spec.CPU) {
		return fmt.Errorf("invalid CPU %s format: %s", specType, spec.CPU)
	}
	if !isValidKubernetesResource(spec.Memory) {
		return fmt.Errorf("invalid Memory %s format: %s", specType, spec.Memory)
	}
	
	return nil
}

// isValidReportingFormat checks if the reporting format is supported
func isValidReportingFormat(format string) bool {
	validFormats := []string{"json", "csv", "prometheus", "yaml"}
	for _, valid := range validFormats {
		if format == valid {
			return true
		}
	}
	return false
}

// isValidKubernetesResource performs basic validation for Kubernetes resource strings
func isValidKubernetesResource(resource string) bool {
	if resource == "" {
		return false
	}
	
	// Basic check for common patterns like "500m", "1Gi", "2", etc.
	// This is a simplified validation - in production, you might want more comprehensive validation
	validSuffixes := []string{"m", "Mi", "Gi", "Ti", "Ki"}
	
	// Check if it's a plain number
	if _, err := fmt.Sscanf(resource, "%d", new(int)); err == nil {
		return true
	}
	
	// Check if it has valid suffix
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(resource, suffix) {
			return true
		}
	}
	
	return false
}

// GetDefaultBenchmarkConfig returns a default benchmark configuration
func GetDefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		APIVersion: "benchmark.victorialogs.io/v1",
		Kind:       "BenchmarkConfig",
		Metadata: ConfigMetadata{
			Name: "distributed-logsql-benchmark",
		},
		Spec: BenchmarkSpec{
			Target: TargetConfig{
				URL: "http://vlselect.monitoring.svc.cluster.local:8481",
				Auth: AuthConfig{
					Type: "none",
				},
				Timeout: 30 * time.Second,
			},
			Workload: WorkloadConfig{
				TotalConcurrentUsers: 200,
				Duration:             600 * time.Second,
				RampUpTime:           60 * time.Second,
				Queries: []QueryConfig{
					{
						Name:     "basic_filter",
						Template: `job:"loki.source.kubernetes.pods"`,
						Weight:   20,
						Category: "filter",
					},
					{
						Name:     "error_requests",
						Template: `_msg:*"status":5* OR _msg:*"status":4*`,
						Weight:   15,
						Category: "complex",
					},
				},
			},
			Workers: WorkersConfig{
				Replicas: 5,
				Resources: WorkerResources{
					Requests: ResourceSpec{
						CPU:    "500m",
						Memory: "1Gi",
					},
					Limits: ResourceSpec{
						CPU:    "2",
						Memory: "4Gi",
					},
				},
			},
			Reporting: ReportingConfig{
				Formats: []string{"json", "csv", "prometheus"},
				Storage: StorageConfig{
					Type: "persistent_volume",
					Path: "/results",
				},
			},
		},
	}
}