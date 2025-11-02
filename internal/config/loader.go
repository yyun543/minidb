package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	configMutex  sync.RWMutex
	onceLoader   sync.Once
)

// Loader manages configuration loading and reloading
type Loader struct {
	v              *viper.Viper
	configPath     string
	watchForChange bool
	onChange       func(*Config)
	mu             sync.RWMutex
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	return &Loader{
		v:              viper.New(),
		configPath:     configPath,
		watchForChange: false,
	}
}

// Load loads the configuration from file and environment
func (l *Loader) Load() (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Set default values
	l.setDefaults()

	// Configure viper
	if l.configPath != "" {
		// Set config file path
		dir := filepath.Dir(l.configPath)
		filename := filepath.Base(l.configPath)
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)

		l.v.AddConfigPath(dir)
		l.v.SetConfigName(name)
		l.v.SetConfigType(strings.TrimPrefix(ext, "."))

		// Read config file if it exists
		if _, err := os.Stat(l.configPath); err == nil {
			if err := l.v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	// Enable environment variable overrides
	l.v.SetEnvPrefix("MINIDB")
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	l.v.AutomaticEnv()

	// Unmarshal into config struct
	config := &Config{}
	if err := l.v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadGlobal loads the global configuration
func LoadGlobal(configPath string) (*Config, error) {
	var err error
	var config *Config

	onceLoader.Do(func() {
		loader := NewLoader(configPath)
		config, err = loader.Load()
		if err == nil {
			configMutex.Lock()
			globalConfig = config
			configMutex.Unlock()
		}
	})

	if err != nil {
		return nil, err
	}

	return GetGlobal(), nil
}

// GetGlobal returns the global configuration
func GetGlobal() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// WatchForChanges enables watching for configuration file changes
func (l *Loader) WatchForChanges(onChange func(*Config)) error {
	l.mu.Lock()
	l.watchForChange = true
	l.onChange = onChange
	l.mu.Unlock()

	l.v.WatchConfig()
	l.v.OnConfigChange(func(e fsnotify.Event) {
		// Reload configuration
		config, err := l.Load()
		if err != nil {
			// Log error but don't crash
			fmt.Fprintf(os.Stderr, "Error reloading config: %v\n", err)
			return
		}

		// Update global config
		configMutex.Lock()
		globalConfig = config
		configMutex.Unlock()

		// Call onChange callback
		l.mu.RLock()
		callback := l.onChange
		l.mu.RUnlock()

		if callback != nil {
			callback(config)
		}
	})

	return nil
}

// setDefaults sets default configuration values in viper
func (l *Loader) setDefaults() {
	defaults := DefaultConfig()

	// Server defaults
	l.v.SetDefault("server.host", defaults.Server.Host)
	l.v.SetDefault("server.port", defaults.Server.Port)
	l.v.SetDefault("server.max_connections", defaults.Server.MaxConnections)

	// Storage defaults
	l.v.SetDefault("storage.data_dir", defaults.Storage.DataDir)
	l.v.SetDefault("storage.format", defaults.Storage.Format)
	l.v.SetDefault("storage.backend", defaults.Storage.Backend)

	// Query cache defaults
	l.v.SetDefault("query_cache.enabled", defaults.QueryCache.Enabled)
	l.v.SetDefault("query_cache.max_size_mb", defaults.QueryCache.MaxSizeMB)
	l.v.SetDefault("query_cache.ttl_seconds", defaults.QueryCache.TTLSeconds)
	l.v.SetDefault("query_cache.min_compute_time_ms", defaults.QueryCache.MinComputeTimeMS)

	// Compaction defaults
	l.v.SetDefault("compaction.enabled", defaults.Compaction.Enabled)
	l.v.SetDefault("compaction.trigger_file_count", defaults.Compaction.TriggerFileCount)
	l.v.SetDefault("compaction.max_file_size_mb", defaults.Compaction.MaxFileSizeMB)
	l.v.SetDefault("compaction.io_rate_limit_mbps", defaults.Compaction.IORateLimitMbps)

	// Statistics defaults
	l.v.SetDefault("statistics.auto_collect", defaults.Statistics.AutoCollect)
	l.v.SetDefault("statistics.sample_rate", defaults.Statistics.SampleRate)
	l.v.SetDefault("statistics.histogram_buckets", defaults.Statistics.HistogramBuckets)
	l.v.SetDefault("statistics.hll_precision", defaults.Statistics.HLLPrecision)

	// Transaction defaults
	l.v.SetDefault("transaction.isolation_level", defaults.Transaction.IsolationLevel)
	l.v.SetDefault("transaction.default_timeout_seconds", defaults.Transaction.DefaultTimeoutSeconds)

	// Vectorization defaults
	l.v.SetDefault("vectorization.enabled", defaults.Vectorization.Enabled)
	l.v.SetDefault("vectorization.batch_size", defaults.Vectorization.BatchSize)
	l.v.SetDefault("vectorization.force_regular_for_complex_where", defaults.Vectorization.ForceRegularForComplexWhere)

	// Cost optimizer defaults
	l.v.SetDefault("cost_optimizer.cpu_cost_per_row", defaults.CostOptimizer.CPUCostPerRow)
	l.v.SetDefault("cost_optimizer.io_cost_per_page", defaults.CostOptimizer.IOCostPerPage)
	l.v.SetDefault("cost_optimizer.memory_cost_per_mb", defaults.CostOptimizer.MemoryCostPerMB)

	// Execution defaults
	l.v.SetDefault("execution.use_vectorized", defaults.Execution.UseVectorized)
	l.v.SetDefault("execution.worker_pool_size", defaults.Execution.WorkerPoolSize)
	l.v.SetDefault("execution.batch_size", defaults.Execution.BatchSize)

	// Cache defaults
	l.v.SetDefault("cache.enabled", defaults.Cache.Enabled)
	l.v.SetDefault("cache.max_size_mb", defaults.Cache.MaxSizeMB)
	l.v.SetDefault("cache.ttl", defaults.Cache.TTL)
	l.v.SetDefault("cache.min_compute_time_ms", defaults.Cache.MinComputeTimeMS)

	// Logging defaults
	l.v.SetDefault("logging.level", defaults.Logging.Level)
	l.v.SetDefault("logging.format", defaults.Logging.Format)
	l.v.SetDefault("logging.output", defaults.Logging.Output)
	l.v.SetDefault("logging.rotation.enabled", defaults.Logging.Rotation.Enabled)
	l.v.SetDefault("logging.rotation.max_size_mb", defaults.Logging.Rotation.MaxSizeMB)
	l.v.SetDefault("logging.rotation.max_backups", defaults.Logging.Rotation.MaxBackups)
	l.v.SetDefault("logging.rotation.max_age_days", defaults.Logging.Rotation.MaxAgeDays)
}

// LoadFromMap loads configuration from a map (useful for testing)
func LoadFromMap(m map[string]interface{}) (*Config, error) {
	v := viper.New()

	// Set defaults first
	loader := &Loader{v: v}
	loader.setDefaults()

	// Override with map values
	for key, value := range m {
		v.Set(key, value)
	}

	// Unmarshal into config struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}
