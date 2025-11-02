package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/config"
)

// TestConfigurationLoading tests loading configuration from YAML file
func TestConfigurationLoading(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minidb.yaml")

	configContent := `
server:
  host: localhost
  port: 7205
  max_connections: 100

storage:
  data_dir: ./data
  format: parquet
  backend: local

query_cache:
  enabled: true
  max_size_mb: 1024
  ttl_seconds: 300
  min_compute_time_ms: 10

compaction:
  enabled: true
  trigger_file_count: 10
  max_file_size_mb: 128
  io_rate_limit_mbps: 20

statistics:
  auto_collect: true
  sample_rate: 0.01
  histogram_buckets: 20
  hll_precision: 14

transaction:
  isolation_level: snapshot
  default_timeout_seconds: 300

vectorization:
  enabled: true
  batch_size: 8192
  force_regular_for_complex_where: false

cost_optimizer:
  cpu_cost_per_row: 0.0001
  io_cost_per_page: 1.0
  memory_cost_per_mb: 0.01
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load configuration
	loader := config.NewLoader(configPath)
	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify server config
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 7205, cfg.Server.Port)
	assert.Equal(t, 100, cfg.Server.MaxConnections)

	// Verify storage config
	assert.Equal(t, "./data", cfg.Storage.DataDir)
	assert.Equal(t, "parquet", cfg.Storage.Format)
	assert.Equal(t, "local", cfg.Storage.Backend)

	// Verify query cache config
	assert.True(t, cfg.QueryCache.Enabled)
	assert.Equal(t, 1024, cfg.QueryCache.MaxSizeMB)
	assert.Equal(t, 300, cfg.QueryCache.TTLSeconds)
	assert.Equal(t, 10, cfg.QueryCache.MinComputeTimeMS)
	assert.Equal(t, 300*time.Second, cfg.QueryCache.GetTTL())
	assert.Equal(t, 10*time.Millisecond, cfg.QueryCache.GetMinComputeTime())

	// Verify transaction config
	assert.Equal(t, "snapshot", cfg.Transaction.IsolationLevel)
	assert.Equal(t, 300, cfg.Transaction.DefaultTimeoutSeconds)
	assert.Equal(t, 300*time.Second, cfg.Transaction.GetDefaultTimeout())

	// Verify vectorization config
	assert.True(t, cfg.Vectorization.Enabled)
	assert.Equal(t, 8192, cfg.Vectorization.BatchSize)

	// Verify cost optimizer config
	assert.Equal(t, 0.0001, cfg.CostOptimizer.CPUCostPerRow)
	assert.Equal(t, 1.0, cfg.CostOptimizer.IOCostPerPage)
	assert.Equal(t, 0.01, cfg.CostOptimizer.MemoryCostPerMB)
}

// TestConfigurationDefaults tests default values when no config file exists
func TestConfigurationDefaults(t *testing.T) {
	// Load config from non-existent file (should use defaults)
	loader := config.NewLoader("/nonexistent/path/minidb.yaml")
	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify defaults
	defaults := config.DefaultConfig()
	assert.Equal(t, defaults.Server.Host, cfg.Server.Host)
	assert.Equal(t, defaults.Server.Port, cfg.Server.Port)
	assert.Equal(t, defaults.QueryCache.Enabled, cfg.QueryCache.Enabled)
	assert.Equal(t, defaults.Transaction.IsolationLevel, cfg.Transaction.IsolationLevel)
}

// TestConfigurationEnvironmentOverride tests environment variable overrides
func TestConfigurationEnvironmentOverride(t *testing.T) {
	// Create temporary config file with default values
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minidb.yaml")

	configContent := `
server:
  host: localhost
  port: 7205
query_cache:
  enabled: true
  max_size_mb: 512
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override
	os.Setenv("MINIDB_SERVER_PORT", "8080")
	os.Setenv("MINIDB_QUERY_CACHE_ENABLED", "false")
	os.Setenv("MINIDB_QUERY_CACHE_MAX_SIZE_MB", "2048")
	defer func() {
		os.Unsetenv("MINIDB_SERVER_PORT")
		os.Unsetenv("MINIDB_QUERY_CACHE_ENABLED")
		os.Unsetenv("MINIDB_QUERY_CACHE_MAX_SIZE_MB")
	}()

	// Load configuration
	loader := config.NewLoader(configPath)
	cfg, err := loader.Load()
	require.NoError(t, err)

	// Verify environment overrides
	assert.Equal(t, 8080, cfg.Server.Port, "Port should be overridden by env var")
	assert.False(t, cfg.QueryCache.Enabled, "Cache enabled should be overridden by env var")
	assert.Equal(t, 2048, cfg.QueryCache.MaxSizeMB, "Cache size should be overridden by env var")

	// Verify non-overridden values remain unchanged
	assert.Equal(t, "localhost", cfg.Server.Host)
}

// TestConfigurationValidation tests validation of configuration values
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name      string
		configMap map[string]interface{}
		expectErr bool
		errMsg    string
	}{
		{
			name: "invalid port negative",
			configMap: map[string]interface{}{
				"server.port": -1,
			},
			expectErr: true,
			errMsg:    "invalid port",
		},
		{
			name: "invalid port too high",
			configMap: map[string]interface{}{
				"server.port": 70000,
			},
			expectErr: true,
			errMsg:    "invalid port",
		},
		{
			name: "invalid isolation level",
			configMap: map[string]interface{}{
				"transaction.isolation_level": "invalid_level",
			},
			expectErr: true,
			errMsg:    "invalid isolation level",
		},
		{
			name: "negative cache size",
			configMap: map[string]interface{}{
				"query_cache.max_size_mb": -100,
			},
			expectErr: true,
			errMsg:    "cache size must be positive",
		},
		{
			name: "invalid storage format",
			configMap: map[string]interface{}{
				"storage.format": "unknown",
			},
			expectErr: true,
			errMsg:    "invalid storage format",
		},
		{
			name: "invalid sample rate",
			configMap: map[string]interface{}{
				"statistics.sample_rate": 1.5,
			},
			expectErr: true,
			errMsg:    "sample_rate must be 0-1",
		},
		{
			name: "valid configuration",
			configMap: map[string]interface{}{
				"server.port":             8080,
				"query_cache.max_size_mb": 2048,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromMap(tt.configMap)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

// TestConfigurationReload tests hot reloading of configuration
func TestConfigurationReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minidb.yaml")

	initialConfig := `
query_cache:
  max_size_mb: 512
`

	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Load initial config
	loader := config.NewLoader(configPath)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, 512, cfg.QueryCache.MaxSizeMB)

	// Setup watch for changes
	reloadChan := make(chan *config.Config, 1)
	err = loader.WatchForChanges(func(newCfg *config.Config) {
		reloadChan <- newCfg
	})
	require.NoError(t, err)

	// Modify config file
	time.Sleep(100 * time.Millisecond)
	updatedConfig := `
query_cache:
  max_size_mb: 1024
`
	err = os.WriteFile(configPath, []byte(updatedConfig), 0644)
	require.NoError(t, err)

	// Wait for reload notification
	select {
	case reloadedCfg := <-reloadChan:
		assert.Equal(t, 1024, reloadedCfg.QueryCache.MaxSizeMB, "Config should be reloaded")
	case <-time.After(2 * time.Second):
		t.Fatal("Config reload timeout")
	}
}

// TestServerConfiguration tests server-specific configuration
func TestServerConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		port           int
		maxConnections int
		expectValid    bool
	}{
		{"valid config", "localhost", 7205, 100, true},
		{"valid all interfaces", "0.0.0.0", 8080, 500, true},
		{"invalid port negative", "localhost", -1, 100, false},
		{"invalid port too high", "localhost", 70000, 100, false},
		{"invalid max connections", "localhost", 7205, -10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromMap(map[string]interface{}{
				"server.host":            tt.host,
				"server.port":            tt.port,
				"server.max_connections": tt.maxConnections,
			})

			if tt.expectValid {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, tt.host, cfg.Server.Host)
				assert.Equal(t, tt.port, cfg.Server.Port)
				assert.Equal(t, tt.maxConnections, cfg.Server.MaxConnections)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestStorageConfiguration tests storage-specific configuration
func TestStorageConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		dataDir string
		format  string
		backend string
		valid   bool
	}{
		{"valid local parquet", "./data", "parquet", "local", true},
		{"valid s3 parquet", "s3://bucket/path", "parquet", "s3", true},
		{"valid orc format", "./data", "orc", "local", true},
		{"invalid format", "./data", "unknown", "local", false},
		{"invalid backend", "./data", "parquet", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromMap(map[string]interface{}{
				"storage.data_dir": tt.dataDir,
				"storage.format":   tt.format,
				"storage.backend":  tt.backend,
			})

			if tt.valid {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestTransactionConfiguration tests transaction configuration
func TestTransactionConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		isolationLevel string
		timeout        int
		valid          bool
	}{
		{"snapshot isolation", "snapshot", 300, true},
		{"read committed", "read_committed", 600, true},
		{"serializable", "serializable", 1200, true},
		{"invalid isolation", "invalid", 300, false},
		{"negative timeout", "snapshot", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromMap(map[string]interface{}{
				"transaction.isolation_level":         tt.isolationLevel,
				"transaction.default_timeout_seconds": tt.timeout,
			})

			if tt.valid {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, tt.isolationLevel, cfg.Transaction.IsolationLevel)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestCostOptimizerConfiguration tests cost optimizer parameters
func TestCostOptimizerConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		cpuCostPerRow   float64
		ioCostPerPage   float64
		memoryCostPerMB float64
		valid           bool
	}{
		{"default costs", 0.0001, 1.0, 0.01, true},
		{"custom costs", 0.0002, 2.0, 0.02, true},
		{"negative cpu cost", -0.1, 1.0, 0.01, false},
		{"negative io cost", 0.0001, -1.0, 0.01, false},
		{"negative memory cost", 0.0001, 1.0, -0.01, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadFromMap(map[string]interface{}{
				"cost_optimizer.cpu_cost_per_row":   tt.cpuCostPerRow,
				"cost_optimizer.io_cost_per_page":   tt.ioCostPerPage,
				"cost_optimizer.memory_cost_per_mb": tt.memoryCostPerMB,
			})

			if tt.valid {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Benchmark configuration loading performance
func BenchmarkConfigurationLoad(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "minidb.yaml")

	configContent := `
server:
  host: localhost
  port: 7205
query_cache:
  enabled: true
  max_size_mb: 1024
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := config.NewLoader(configPath)
		_, err := loader.Load()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConfigurationFromMap benchmarks loading from map
func BenchmarkConfigurationFromMap(b *testing.B) {
	configMap := map[string]interface{}{
		"server.port":             7205,
		"query_cache.max_size_mb": 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.LoadFromMap(configMap)
		if err != nil {
			b.Fatal(err)
		}
	}
}
