package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// ContextKey is the key type for logger context values
type ContextKey string

const (
	// LoggerContextKey is the context key for logger
	LoggerContextKey ContextKey = "logger"
	// RequestIDContextKey is the context key for request ID
	RequestIDContextKey ContextKey = "request_id"
	// SessionIDContextKey is the context key for session ID
	SessionIDContextKey ContextKey = "session_id"
)

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(*Logger); ok {
		return logger
	}
	return GetLogger()
}

// ToContext adds logger to context
func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	baseLogger := FromContext(ctx)
	logger := &Logger{
		Logger: baseLogger.With(zap.String("request_id", requestID)),
		config: baseLogger.config,
	}
	ctx = ToContext(ctx, logger)
	return context.WithValue(ctx, RequestIDContextKey, requestID)
}

// WithSessionID adds session ID to context
func WithSessionID(ctx context.Context, sessionID int64) context.Context {
	baseLogger := FromContext(ctx)
	logger := &Logger{
		Logger: baseLogger.With(zap.Int64("session_id", sessionID)),
		config: baseLogger.config,
	}
	ctx = ToContext(ctx, logger)
	return context.WithValue(ctx, SessionIDContextKey, sessionID)
}

// Timer helps measure operation duration
type Timer struct {
	start  time.Time
	logger *Logger
	name   string
	fields []zap.Field
}

// NewTimer creates a new timer for performance logging
func NewTimer(name string, fields ...zap.Field) *Timer {
	return &Timer{
		start:  time.Now(),
		logger: GetLogger(),
		name:   name,
		fields: fields,
	}
}

// NewTimerWithLogger creates a new timer with specific logger
func NewTimerWithLogger(logger *Logger, name string, fields ...zap.Field) *Timer {
	return &Timer{
		start:  time.Now(),
		logger: logger,
		name:   name,
		fields: fields,
	}
}

// Stop stops the timer and logs the duration
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.start)
	fields := append(t.fields, zap.Duration("duration", duration))
	t.logger.Info(t.name, fields...)
	return duration
}

// StopWithLevel stops the timer and logs with specified level
func (t *Timer) StopWithLevel(level LogLevel) time.Duration {
	duration := time.Since(t.start)
	fields := append(t.fields, zap.Duration("duration", duration))

	switch level {
	case DebugLevel:
		t.logger.Debug(t.name, fields...)
	case InfoLevel:
		t.logger.Info(t.name, fields...)
	case WarnLevel:
		t.logger.Warn(t.name, fields...)
	case ErrorLevel:
		t.logger.Error(t.name, fields...)
	}

	return duration
}

// StopWithError stops the timer and logs with error if provided
func (t *Timer) StopWithError(err error) time.Duration {
	duration := time.Since(t.start)
	fields := append(t.fields, zap.Duration("duration", duration))

	if err != nil {
		fields = append(fields, zap.Error(err))
		t.logger.Error(t.name, fields...)
	} else {
		t.logger.Info(t.name, fields...)
	}

	return duration
}

// OperationLogger provides structured logging for database operations
type OperationLogger struct {
	logger *Logger
	timer  *Timer
}

// NewOperationLogger creates a new operation logger
func NewOperationLogger(operation string, fields ...zap.Field) *OperationLogger {
	baseLogger := GetLogger()
	logger := &Logger{
		Logger: baseLogger.With(fields...),
		config: baseLogger.config,
	}
	timer := NewTimerWithLogger(logger, operation, fields...)

	logger.Debug("Operation started", zap.String("operation", operation))

	return &OperationLogger{
		logger: logger,
		timer:  timer,
	}
}

// Success logs successful completion of the operation
func (ol *OperationLogger) Success(msg string, fields ...zap.Field) {
	duration := ol.timer.Stop()
	allFields := append(fields, zap.Duration("duration", duration), zap.Bool("success", true))
	ol.logger.Info(msg, allFields...)
}

// Error logs error completion of the operation
func (ol *OperationLogger) Error(msg string, err error, fields ...zap.Field) {
	duration := time.Since(ol.timer.start)
	allFields := append(fields,
		zap.Duration("duration", duration),
		zap.Bool("success", false),
		zap.Error(err),
	)
	ol.logger.Error(msg, allFields...)
}

// Warning logs warning during the operation
func (ol *OperationLogger) Warning(msg string, fields ...zap.Field) {
	ol.logger.Warn(msg, fields...)
}

// Debug logs debug information during the operation
func (ol *OperationLogger) Debug(msg string, fields ...zap.Field) {
	ol.logger.Debug(msg, fields...)
}
