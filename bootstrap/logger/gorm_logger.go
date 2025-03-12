package logger

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/logger"
)

// GormLogger adapts our logging system to GORM's logger interface
type GormLogger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration // Configurable threshold for slow SQL queries
	contextKeys   []any         // Context keys to extract from the context
}

// NewGormLogger creates a new GORM logger instance with custom settings
func NewGormLogger(level logger.LogLevel) *GormLogger {
	return &GormLogger{
		LogLevel:      level,
		SlowThreshold: 200 * time.Millisecond, // Default threshold
	}
}

// WithSlowThreshold sets a custom threshold for slow query warnings
func (l *GormLogger) WithSlowThreshold(threshold time.Duration) *GormLogger {
	l.SlowThreshold = threshold
	return l
}

// WithContextKeys sets context keys to be logged
func (l *GormLogger) WithContextKeys(keys ...any) *GormLogger {
	l.contextKeys = keys
	return l
}

// LogMode implements GORM's logger interface
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l // Create a copy to avoid mutating the original
	newLogger.LogLevel = level
	return &newLogger
}

// Info implements Info logging
func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Info {
		Infof(ctx, msg, data...)
	}
}

// Warn implements Warning logging
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		Warnf(ctx, msg, data...)
	}
}

// Error implements Error logging
func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		Errorf(ctx, msg, data...)
	}
}

// Trace implements SQL execution tracing
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Pre-allocate with appropriate capacity
	kv := make(map[string]any, 5)
	kv["type"] = "sql"
	kv["duration"] = elapsed.String()
	kv["sql"] = sql
	kv["rows"] = rows

	// Extract request_id if present
	for _, key := range l.contextKeys {
		if value := ctx.Value(key); value != nil {
			kv[fmt.Sprintf("%v", key)] = value
		}
	}

	// Log based on result and threshold
	switch {
	case err != nil && l.LogLevel >= logger.Error:
		kv["error"] = err.Error() // Include the error message
		LogWith(LevelError, "gorm_query_error", kv)
	case elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		LogWith(LevelWarn, "gorm_slow_query", kv)
	case l.LogLevel >= logger.Info:
		LogWith(LevelInfo, "gorm_trace", kv)
	}
}
