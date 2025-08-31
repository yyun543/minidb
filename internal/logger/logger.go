package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents different log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// Environment represents the runtime environment
type Environment string

const (
	DevelopmentEnv Environment = "development"
	ProductionEnv  Environment = "production"
	TestEnv        Environment = "test"
)

// Config represents logger configuration
type Config struct {
	// Level is the minimum log level
	Level LogLevel
	// Environment specifies the runtime environment
	Environment Environment
	// LogDir is the directory where log files will be stored
	LogDir string
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	MaxSize int
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int
	// Compress determines if the rotated log files should be compressed using gzip
	Compress bool
	// EnableConsole determines if logs should be output to console
	EnableConsole bool
	// ServiceName is the name of the service for structured logging
	ServiceName string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Level:         InfoLevel,
		Environment:   DevelopmentEnv,
		LogDir:        "logs",
		MaxSize:       100, // 100MB
		MaxAge:        7,   // 7 days
		MaxBackups:    10,  // 10 files
		Compress:      true,
		EnableConsole: true,
		ServiceName:   "minidb",
	}
}

// Logger wraps zap logger with additional functionality
type Logger struct {
	*zap.Logger
	config *Config
}

// Global logger instance
var globalLogger *Logger

// InitLogger initializes the global logger with the given configuration
func InitLogger(config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	// Ensure log directory exists
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return err
	}

	// Configure log level
	level := zapcore.InfoLevel
	switch config.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	}

	// Create encoder config
	encoderConfig := getEncoderConfig(config.Environment)

	// Create cores
	var cores []zapcore.Core

	// File core with rotation
	if config.LogDir != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   filepath.Join(config.LogDir, "minidb.log"),
			MaxSize:    config.MaxSize,
			MaxAge:     config.MaxAge,
			MaxBackups: config.MaxBackups,
			Compress:   config.Compress,
			LocalTime:  true,
		}

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			level,
		)
		cores = append(cores, fileCore)
	}

	// Console core
	if config.EnableConsole {
		var consoleEncoder zapcore.Encoder
		if config.Environment == DevelopmentEnv {
			consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		} else {
			consoleEncoder = zapcore.NewJSONEncoder(encoderConfig)
		}

		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger with additional options
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", config.ServiceName),
			zap.String("environment", string(config.Environment)),
		),
	)

	globalLogger = &Logger{
		Logger: zapLogger,
		config: config,
	}

	return nil
}

// getEncoderConfig returns encoder configuration based on environment
func getEncoderConfig(env Environment) zapcore.EncoderConfig {
	config := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if env == DevelopmentEnv {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncodeCaller = zapcore.FullCallerEncoder
	}

	return config
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		// Initialize with default config if not already initialized
		_ = InitLogger(DefaultConfig())
	}
	return globalLogger
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: GetLogger().With(fields...),
		config: GetLogger().config,
	}
}

// Named creates a named logger
func Named(name string) *Logger {
	return &Logger{
		Logger: GetLogger().Named(name),
		config: GetLogger().config,
	}
}

// WithComponent creates a logger for a specific component
func WithComponent(component string) *Logger {
	return With(zap.String("component", component))
}

// WithSession creates a logger for a specific session
func WithSession(sessionID int64) *Logger {
	return With(zap.Int64("session_id", sessionID))
}

// WithDatabase creates a logger for database operations
func WithDatabase(dbName string) *Logger {
	return With(zap.String("database", dbName))
}

// WithTable creates a logger for table operations
func WithTable(dbName, tableName string) *Logger {
	return With(
		zap.String("database", dbName),
		zap.String("table", tableName),
	)
}

// WithQuery creates a logger for SQL query operations
func WithQuery(query string) *Logger {
	return With(zap.String("query", query))
}

// WithDuration creates a logger with duration field
func WithDuration(duration time.Duration) *Logger {
	return With(zap.Duration("duration", duration))
}

// WithError creates a logger with error field
func WithError(err error) *Logger {
	if err == nil {
		return GetLogger()
	}
	return With(zap.Error(err))
}

// WithClient creates a logger for client operations
func WithClient(clientAddr string) *Logger {
	return With(zap.String("client", clientAddr))
}

// Performance logging helpers

// LogQueryPerformance logs query execution performance
func LogQueryPerformance(query string, duration time.Duration, rowsAffected int64) {
	Info("Query executed",
		zap.String("query", query),
		zap.Duration("duration", duration),
		zap.Int64("rows_affected", rowsAffected),
	)
}

// LogConnectionEvent logs connection events
func LogConnectionEvent(event string, clientAddr string, sessionID int64) {
	Info("Connection event",
		zap.String("event", event),
		zap.String("client", clientAddr),
		zap.Int64("session_id", sessionID),
	)
}

// LogCatalogOperation logs catalog operations
func LogCatalogOperation(operation, database, table string, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("database", database),
	}
	if table != "" {
		fields = append(fields, zap.String("table", table))
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		Error("Catalog operation failed", fields...)
	} else {
		Info("Catalog operation completed", fields...)
	}
}

// LogStorageOperation logs storage operations
func LogStorageOperation(operation string, key []byte, size int, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", string(key)),
		zap.Int("size", size),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		Error("Storage operation failed", fields...)
	} else {
		Debug("Storage operation completed", fields...)
	}
}

// LogServerEvent logs server lifecycle events
func LogServerEvent(event string, details ...zap.Field) {
	allFields := []zap.Field{zap.String("event", event)}
	allFields = append(allFields, details...)
	Info("Server event", allFields...)
}
