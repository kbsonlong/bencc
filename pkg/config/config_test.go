package config

import (
	"os"
	"testing"
	"time"
	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
coordinator:
  port: 8080
  grpc_port: 8090
  heartbeat_timeout: 30s

worker:
  port: 8080
  max_concurrent: 100
  heartbeat_interval: 10s

target:
  url: "http://vlselect:8481"
  timeout: 30s
  auth:
    type: "none"

workload:
  total_concurrent_users: 100
  duration: 300s
  ramp_up_time: 60s
  queries:
    - name: "basic_filter"
      template: 'job:"test"'
      weight: 20
      category: "filter"

monitoring:
  enabled: true
  metrics_port: 9090
`

	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Test loading configuration
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify configuration values
	if config.Coordinator.Port != 8080 {
		t.Errorf("Expected coordinator port 8080, got %d", config.Coordinator.Port)
	}
	if config.Worker.MaxConcurrent != 100 {
		t.Errorf("Expected max concurrent 100, got %d", config.Worker.MaxConcurrent)
	}
	if config.Target.URL != "http://vlselect:8481" {
		t.Errorf("Expected target URL 'http://vlselect:8481', got %s", config.Target.URL)
	}
	if len(config.Workload.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(config.Workload.Queries))
	}
}

func TestLoadBenchmarkConfig(t *testing.T) {
	// Create a temporary benchmark config file
	benchmarkConfig := GetDefaultBenchmarkConfig()
	
	data, err := yaml.Marshal(benchmarkConfig)
	if err != nil {
		t.Fatalf("Failed to marshal benchmark config: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "benchmark-config-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		t.Fatalf("Failed to write benchmark config: %v", err)
	}
	tmpFile.Close()

	// Test loading benchmark configuration
	config, err := LoadBenchmarkConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load benchmark config: %v", err)
	}

	// Verify configuration values
	if config.APIVersion != "benchmark.victorialogs.io/v1" {
		t.Errorf("Expected API version 'benchmark.victorialogs.io/v1', got %s", config.APIVersion)
	}
	if config.Kind != "BenchmarkConfig" {
		t.Errorf("Expected kind 'BenchmarkConfig', got %s", config.Kind)
	}
	if config.Spec.Workers.Replicas != 5 {
		t.Errorf("Expected 5 worker replicas, got %d", config.Spec.Workers.Replicas)
	}
	if len(config.Spec.Workload.Queries) != 2 {
		t.Errorf("Expected 2 queries, got %d", len(config.Spec.Workload.Queries))
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Coordinator: CoordinatorConfig{
					Port:             8080,
					GRPCPort:         8090,
					HeartbeatTimeout: 30 * time.Second,
				},
				Worker: WorkerConfig{
					Port:              8080,
					MaxConcurrent:     100,
					HeartbeatInterval: 10 * time.Second,
				},
				Target: TargetConfig{
					URL:     "http://vlselect:8481",
					Timeout: 30 * time.Second,
					Auth:    AuthConfig{Type: "none"},
				},
				Workload: WorkloadConfig{
					TotalConcurrentUsers: 100,
					Duration:             300 * time.Second,
					RampUpTime:           60 * time.Second,
					Queries: []QueryConfig{
						{
							Name:     "test",
							Template: "test query",
							Weight:   10,
							Category: "test",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid coordinator port",
			config: &Config{
				Coordinator: CoordinatorConfig{
					Port:             -1,
					GRPCPort:         8090,
					HeartbeatTimeout: 30 * time.Second,
				},
				Worker: WorkerConfig{
					Port:              8080,
					MaxConcurrent:     100,
					HeartbeatInterval: 10 * time.Second,
				},
				Target: TargetConfig{
					URL:     "http://vlselect:8481",
					Timeout: 30 * time.Second,
					Auth:    AuthConfig{Type: "none"},
				},
				Workload: WorkloadConfig{
					TotalConcurrentUsers: 100,
					Duration:             300 * time.Second,
					Queries: []QueryConfig{
						{
							Name:     "test",
							Template: "test query",
							Weight:   10,
							Category: "test",
						},
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid target URL",
			config: &Config{
				Coordinator: CoordinatorConfig{
					Port:             8080,
					GRPCPort:         8090,
					HeartbeatTimeout: 30 * time.Second,
				},
				Worker: WorkerConfig{
					Port:              8080,
					MaxConcurrent:     100,
					HeartbeatInterval: 10 * time.Second,
				},
				Target: TargetConfig{
					URL:     "invalid-url",
					Timeout: 30 * time.Second,
					Auth:    AuthConfig{Type: "none"},
				},
				Workload: WorkloadConfig{
					TotalConcurrentUsers: 100,
					Duration:             300 * time.Second,
					Queries: []QueryConfig{
						{
							Name:     "test",
							Template: "test query",
							Weight:   10,
							Category: "test",
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestAuthConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		auth        AuthConfig
		expectError bool
	}{
		{
			name:        "none auth",
			auth:        AuthConfig{Type: "none"},
			expectError: false,
		},
		{
			name:        "bearer auth with token",
			auth:        AuthConfig{Type: "bearer", Token: "test-token"},
			expectError: false,
		},
		{
			name:        "bearer auth without token",
			auth:        AuthConfig{Type: "bearer"},
			expectError: true,
		},
		{
			name:        "basic auth with credentials",
			auth:        AuthConfig{Type: "basic", User: "user", Pass: "pass"},
			expectError: false,
		},
		{
			name:        "basic auth without credentials",
			auth:        AuthConfig{Type: "basic"},
			expectError: true,
		},
		{
			name:        "unsupported auth type",
			auth:        AuthConfig{Type: "oauth"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuthConfig(&tt.auth)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestQueryConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		query       QueryConfig
		expectError bool
	}{
		{
			name: "valid query",
			query: QueryConfig{
				Name:     "test",
				Template: "test query",
				Weight:   10,
				Category: "test",
			},
			expectError: false,
		},
		{
			name: "missing name",
			query: QueryConfig{
				Template: "test query",
				Weight:   10,
				Category: "test",
			},
			expectError: true,
		},
		{
			name: "missing template",
			query: QueryConfig{
				Name:     "test",
				Weight:   10,
				Category: "test",
			},
			expectError: true,
		},
		{
			name: "invalid weight",
			query: QueryConfig{
				Name:     "test",
				Template: "test query",
				Weight:   0,
				Category: "test",
			},
			expectError: true,
		},
		{
			name: "missing category",
			query: QueryConfig{
				Name:     "test",
				Template: "test query",
				Weight:   10,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQueryConfig(&tt.query)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestResourceSpecValidation(t *testing.T) {
	tests := []struct {
		name        string
		spec        ResourceSpec
		expectError bool
	}{
		{
			name:        "valid resource spec",
			spec:        ResourceSpec{CPU: "500m", Memory: "1Gi"},
			expectError: false,
		},
		{
			name:        "valid resource spec with plain numbers",
			spec:        ResourceSpec{CPU: "2", Memory: "4096"},
			expectError: false,
		},
		{
			name:        "missing CPU",
			spec:        ResourceSpec{Memory: "1Gi"},
			expectError: true,
		},
		{
			name:        "missing Memory",
			spec:        ResourceSpec{CPU: "500m"},
			expectError: true,
		},
		{
			name:        "invalid CPU format",
			spec:        ResourceSpec{CPU: "invalid", Memory: "1Gi"},
			expectError: true,
		},
		{
			name:        "invalid Memory format",
			spec:        ResourceSpec{CPU: "500m", Memory: "invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceSpec(&tt.spec, "test")
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestIsValidReportingFormat(t *testing.T) {
	validFormats := []string{"json", "csv", "prometheus", "yaml"}
	invalidFormats := []string{"xml", "txt", "binary", ""}

	for _, format := range validFormats {
		if !isValidReportingFormat(format) {
			t.Errorf("Expected %s to be valid reporting format", format)
		}
	}

	for _, format := range invalidFormats {
		if isValidReportingFormat(format) {
			t.Errorf("Expected %s to be invalid reporting format", format)
		}
	}
}

func TestGetDefaultBenchmarkConfig(t *testing.T) {
	config := GetDefaultBenchmarkConfig()

	// Test that default config is valid
	if err := validateBenchmarkConfig(config); err != nil {
		t.Errorf("Default benchmark config should be valid, but got error: %v", err)
	}

	// Test specific default values
	if config.APIVersion != "benchmark.victorialogs.io/v1" {
		t.Errorf("Expected default API version 'benchmark.victorialogs.io/v1', got %s", config.APIVersion)
	}
	if config.Spec.Workers.Replicas != 5 {
		t.Errorf("Expected default worker replicas 5, got %d", config.Spec.Workers.Replicas)
	}
	if len(config.Spec.Workload.Queries) != 2 {
		t.Errorf("Expected 2 default queries, got %d", len(config.Spec.Workload.Queries))
	}
}