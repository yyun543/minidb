package logger

import (
	"os"
	"strconv"
	"strings"
)

// ConfigFromEnv creates a logger configuration from environment variables
func ConfigFromEnv() *Config {
	config := DefaultConfig()

	// LOG_LEVEL
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = LogLevel(strings.ToLower(level))
	}

	// ENVIRONMENT
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = Environment(strings.ToLower(env))
	}

	// LOG_DIR
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		config.LogDir = logDir
	}

	// LOG_MAX_SIZE
	if maxSize := os.Getenv("LOG_MAX_SIZE"); maxSize != "" {
		if size, err := strconv.Atoi(maxSize); err == nil {
			config.MaxSize = size
		}
	}

	// LOG_MAX_AGE
	if maxAge := os.Getenv("LOG_MAX_AGE"); maxAge != "" {
		if age, err := strconv.Atoi(maxAge); err == nil {
			config.MaxAge = age
		}
	}

	// LOG_MAX_BACKUPS
	if maxBackups := os.Getenv("LOG_MAX_BACKUPS"); maxBackups != "" {
		if backups, err := strconv.Atoi(maxBackups); err == nil {
			config.MaxBackups = backups
		}
	}

	// LOG_COMPRESS
	if compress := os.Getenv("LOG_COMPRESS"); compress != "" {
		config.Compress = strings.ToLower(compress) == "true"
	}

	// LOG_CONSOLE
	if console := os.Getenv("LOG_CONSOLE"); console != "" {
		config.EnableConsole = strings.ToLower(console) == "true"
	}

	// SERVICE_NAME
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	return config
}

// GetConfigForEnvironment returns optimized configuration for specific environments
func GetConfigForEnvironment(env Environment) *Config {
	config := DefaultConfig()
	config.Environment = env

	switch env {
	case DevelopmentEnv:
		config.Level = DebugLevel
		config.EnableConsole = true
		config.LogDir = "logs"
		config.MaxSize = 50   // 50MB in dev
		config.MaxAge = 3     // 3 days in dev
		config.MaxBackups = 5 // 5 files in dev

	case ProductionEnv:
		config.Level = InfoLevel
		config.EnableConsole = false // Only file logging in production
		config.LogDir = "/var/log/minidb"
		config.MaxSize = 200   // 200MB in prod
		config.MaxAge = 30     // 30 days in prod
		config.MaxBackups = 20 // 20 files in prod
		config.Compress = true

	case TestEnv:
		config.Level = WarnLevel
		config.EnableConsole = false
		config.LogDir = "test_logs"
		config.MaxSize = 10   // 10MB in test
		config.MaxAge = 1     // 1 day in test
		config.MaxBackups = 3 // 3 files in test
	}

	return config
}

// ValidateConfig validates the logger configuration
func ValidateConfig(config *Config) error {
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.MaxAge <= 0 {
		config.MaxAge = 7
	}
	if config.MaxBackups < 0 {
		config.MaxBackups = 0
	}
	if config.ServiceName == "" {
		config.ServiceName = "minidb"
	}

	return nil
}
