package config

import (
	"fmt"
	"time"
)

// Config represents the complete MiniDB configuration
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Storage       StorageConfig       `mapstructure:"storage"`
	Cache         CacheConfig         `mapstructure:"cache"`
	QueryCache    QueryCacheConfig    `mapstructure:"query_cache"`
	Compaction    CompactionConfig    `mapstructure:"compaction"`
	Execution     ExecutionConfig     `mapstructure:"execution"`
	Statistics    StatisticsConfig    `mapstructure:"statistics"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	Transaction   TransactionConfig   `mapstructure:"transaction"`
	Vectorization VectorizationConfig `mapstructure:"vectorization"`
	CostOptimizer CostOptimizerConfig `mapstructure:"cost_optimizer"`
}

// ServerConfig contains server-related settings
type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	MaxConnections int    `mapstructure:"max_connections"`
}

// StorageConfig contains storage-related settings
type StorageConfig struct {
	DataDir string `mapstructure:"data_dir"`
	Format  string `mapstructure:"format"`  // parquet, orc
	Backend string `mapstructure:"backend"` // local, s3, minio
}

// QueryCacheConfig contains query cache settings
type QueryCacheConfig struct {
	Enabled          bool `mapstructure:"enabled"`
	MaxSizeMB        int  `mapstructure:"max_size_mb"`
	TTLSeconds       int  `mapstructure:"ttl_seconds"`
	MinComputeTimeMS int  `mapstructure:"min_compute_time_ms"`
}

// CompactionConfig contains compaction settings
type CompactionConfig struct {
	Enabled          bool `mapstructure:"enabled"`
	TriggerFileCount int  `mapstructure:"trigger_file_count"`
	MaxFileSizeMB    int  `mapstructure:"max_file_size_mb"`
	IORateLimitMbps  int  `mapstructure:"io_rate_limit_mbps"`
}

// StatisticsConfig contains statistics collection settings
type StatisticsConfig struct {
	AutoCollect      bool    `mapstructure:"auto_collect"`
	SampleRate       float64 `mapstructure:"sample_rate"`
	HistogramBuckets int     `mapstructure:"histogram_buckets"`
	HLLPrecision     int     `mapstructure:"hll_precision"`
}

// TransactionConfig contains transaction settings
type TransactionConfig struct {
	IsolationLevel        string `mapstructure:"isolation_level"` // READ_UNCOMMITTED, READ_COMMITTED, REPEATABLE_READ, SERIALIZABLE
	DefaultTimeoutSeconds int    `mapstructure:"default_timeout_seconds"`
}

// VectorizationConfig contains vectorization settings
type VectorizationConfig struct {
	Enabled                     bool `mapstructure:"enabled"`
	BatchSize                   int  `mapstructure:"batch_size"`
	ForceRegularForComplexWhere bool `mapstructure:"force_regular_for_complex_where"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	MaxSizeMB        int    `mapstructure:"max_size_mb"`
	TTL              string `mapstructure:"ttl"`
	MinComputeTimeMS int    `mapstructure:"min_compute_time_ms"`
}

// ExecutionConfig contains execution settings
type ExecutionConfig struct {
	UseVectorized  bool `mapstructure:"use_vectorized"`
	WorkerPoolSize int  `mapstructure:"worker_pool_size"`
	BatchSize      int  `mapstructure:"batch_size"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level    string         `mapstructure:"level"`
	Format   string         `mapstructure:"format"`
	Output   string         `mapstructure:"output"`
	Rotation RotationConfig `mapstructure:"rotation"`
}

// RotationConfig contains log rotation settings
type RotationConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	MaxSizeMB  int  `mapstructure:"max_size_mb"`
	MaxBackups int  `mapstructure:"max_backups"`
	MaxAgeDays int  `mapstructure:"max_age_days"`
}

// CostOptimizerConfig contains cost model parameters
type CostOptimizerConfig struct {
	CPUCostPerRow   float64 `mapstructure:"cpu_cost_per_row"`
	IOCostPerPage   float64 `mapstructure:"io_cost_per_page"`
	MemoryCostPerMB float64 `mapstructure:"memory_cost_per_mb"`
}

// GetTTL returns the cache TTL as a duration
func (c *QueryCacheConfig) GetTTL() time.Duration {
	return time.Duration(c.TTLSeconds) * time.Second
}

// GetMinComputeTime returns the minimum compute time as a duration
func (c *QueryCacheConfig) GetMinComputeTime() time.Duration {
	return time.Duration(c.MinComputeTimeMS) * time.Millisecond
}

// GetDefaultTimeout returns the default transaction timeout as a duration
func (c *TransactionConfig) GetDefaultTimeout() time.Duration {
	return time.Duration(c.DefaultTimeoutSeconds) * time.Second
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config invalid: %w", err)
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("storage config invalid: %w", err)
	}
	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache config invalid: %w", err)
	}
	if err := c.QueryCache.Validate(); err != nil {
		return fmt.Errorf("query cache config invalid: %w", err)
	}
	if err := c.Compaction.Validate(); err != nil {
		return fmt.Errorf("compaction config invalid: %w", err)
	}
	if err := c.Execution.Validate(); err != nil {
		return fmt.Errorf("execution config invalid: %w", err)
	}
	if err := c.Statistics.Validate(); err != nil {
		return fmt.Errorf("statistics config invalid: %w", err)
	}
	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config invalid: %w", err)
	}
	if err := c.Transaction.Validate(); err != nil {
		return fmt.Errorf("transaction config invalid: %w", err)
	}
	if err := c.Vectorization.Validate(); err != nil {
		return fmt.Errorf("vectorization config invalid: %w", err)
	}
	if err := c.CostOptimizer.Validate(); err != nil {
		return fmt.Errorf("cost optimizer config invalid: %w", err)
	}
	return nil
}

// Validate validates server configuration
func (c *ServerConfig) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Port)
	}
	if c.MaxConnections < 1 {
		return fmt.Errorf("max_connections must be positive, got: %d", c.MaxConnections)
	}
	return nil
}

// Validate validates storage configuration
func (c *StorageConfig) Validate() error {
	validFormats := map[string]bool{"parquet": true, "orc": true}
	if !validFormats[c.Format] {
		return fmt.Errorf("invalid storage format: %s (must be: parquet, orc)", c.Format)
	}

	validBackends := map[string]bool{"local": true, "s3": true, "minio": true}
	if !validBackends[c.Backend] {
		return fmt.Errorf("invalid storage backend: %s (must be: local, s3, minio)", c.Backend)
	}

	return nil
}

// Validate validates query cache configuration
func (c *QueryCacheConfig) Validate() error {
	if c.MaxSizeMB < 0 {
		return fmt.Errorf("cache size must be positive, got: %d", c.MaxSizeMB)
	}
	if c.TTLSeconds < 0 {
		return fmt.Errorf("ttl must be non-negative, got: %d", c.TTLSeconds)
	}
	if c.MinComputeTimeMS < 0 {
		return fmt.Errorf("min_compute_time_ms must be non-negative, got: %d", c.MinComputeTimeMS)
	}
	return nil
}

// Validate validates compaction configuration
func (c *CompactionConfig) Validate() error {
	if c.TriggerFileCount < 0 {
		return fmt.Errorf("trigger_file_count must be non-negative, got: %d", c.TriggerFileCount)
	}
	if c.MaxFileSizeMB < 0 {
		return fmt.Errorf("max_file_size_mb must be non-negative, got: %d", c.MaxFileSizeMB)
	}
	if c.IORateLimitMbps < 0 {
		return fmt.Errorf("io_rate_limit_mbps must be non-negative, got: %d", c.IORateLimitMbps)
	}
	return nil
}

// Validate validates statistics configuration
func (c *StatisticsConfig) Validate() error {
	if c.SampleRate < 0 || c.SampleRate > 1 {
		return fmt.Errorf("sample_rate must be 0-1, got: %f", c.SampleRate)
	}
	if c.HistogramBuckets < 1 {
		return fmt.Errorf("histogram_buckets must be positive, got: %d", c.HistogramBuckets)
	}
	if c.HLLPrecision < 4 || c.HLLPrecision > 18 {
		return fmt.Errorf("hll_precision must be 4-18, got: %d", c.HLLPrecision)
	}
	return nil
}

// Validate validates transaction configuration
func (c *TransactionConfig) Validate() error {
	validLevels := map[string]bool{
		"READ_UNCOMMITTED": true,
		"read_uncommitted": true,
		"READ_COMMITTED":   true,
		"read_committed":   true,
		"REPEATABLE_READ":  true,
		"repeatable_read":  true,
		"SERIALIZABLE":     true,
		"serializable":     true,
		// For backward compatibility
		"SNAPSHOT": true,
		"snapshot": true,
	}
	if !validLevels[c.IsolationLevel] {
		return fmt.Errorf("invalid isolation level: %s (must be: READ_UNCOMMITTED, READ_COMMITTED, REPEATABLE_READ, or SERIALIZABLE)", c.IsolationLevel)
	}
	if c.DefaultTimeoutSeconds < 0 {
		return fmt.Errorf("default_timeout_seconds must be non-negative, got: %d", c.DefaultTimeoutSeconds)
	}
	return nil
}

// Validate validates vectorization configuration
func (c *VectorizationConfig) Validate() error {
	if c.BatchSize < 1 {
		return fmt.Errorf("batch_size must be positive, got: %d", c.BatchSize)
	}
	return nil
}

// Validate validates cache configuration
func (c *CacheConfig) Validate() error {
	if c.MaxSizeMB < 0 {
		return fmt.Errorf("max_size_mb must be non-negative, got: %d", c.MaxSizeMB)
	}
	return nil
}

// Validate validates execution configuration
func (c *ExecutionConfig) Validate() error {
	if c.BatchSize < 1 {
		return fmt.Errorf("batch_size must be positive, got: %d", c.BatchSize)
	}
	if c.WorkerPoolSize < 1 {
		return fmt.Errorf("worker_pool_size must be positive, got: %d", c.WorkerPoolSize)
	}
	return nil
}

// Validate validates logging configuration
func (c *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s (must be: debug, info, warn, error)", c.Level)
	}
	return nil
}

// Validate validates cost optimizer configuration
func (c *CostOptimizerConfig) Validate() error {
	if c.CPUCostPerRow < 0 {
		return fmt.Errorf("cpu_cost_per_row must be non-negative, got: %f", c.CPUCostPerRow)
	}
	if c.IOCostPerPage < 0 {
		return fmt.Errorf("io_cost_per_page must be non-negative, got: %f", c.IOCostPerPage)
	}
	if c.MemoryCostPerMB < 0 {
		return fmt.Errorf("memory_cost_per_mb must be non-negative, got: %f", c.MemoryCostPerMB)
	}
	return nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           "localhost",
			Port:           7205,
			MaxConnections: 100,
		},
		Storage: StorageConfig{
			DataDir: "./minidb_data",
			Format:  "parquet",
			Backend: "local",
		},
		Cache: CacheConfig{
			Enabled:          true,
			MaxSizeMB:        512,
			TTL:              "5m",
			MinComputeTimeMS: 10,
		},
		QueryCache: QueryCacheConfig{
			Enabled:          true,
			MaxSizeMB:        1024,
			TTLSeconds:       300,
			MinComputeTimeMS: 10,
		},
		Compaction: CompactionConfig{
			Enabled:          true,
			TriggerFileCount: 10,
			MaxFileSizeMB:    128,
			IORateLimitMbps:  20,
		},
		Execution: ExecutionConfig{
			UseVectorized:  true,
			WorkerPoolSize: 4,
			BatchSize:      1024,
		},
		Statistics: StatisticsConfig{
			AutoCollect:      true,
			SampleRate:       0.1,
			HistogramBuckets: 20,
			HLLPrecision:     14,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
			Rotation: RotationConfig{
				Enabled:    true,
				MaxSizeMB:  100,
				MaxBackups: 7,
				MaxAgeDays: 30,
			},
		},
		Transaction: TransactionConfig{
			IsolationLevel:        "READ_COMMITTED",
			DefaultTimeoutSeconds: 300,
		},
		Vectorization: VectorizationConfig{
			Enabled:                     true,
			BatchSize:                   8192,
			ForceRegularForComplexWhere: false,
		},
		CostOptimizer: CostOptimizerConfig{
			CPUCostPerRow:   0.0001,
			IOCostPerPage:   1.0,
			MemoryCostPerMB: 0.01,
		},
	}
}
