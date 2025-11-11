package logger

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
	contextKeys   []any
}

func NewGormLogger(level logger.LogLevel) *GormLogger {
	return &GormLogger{
		LogLevel:      level,
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (l *GormLogger) WithSlowThreshold(threshold time.Duration) *GormLogger {
	l.SlowThreshold = threshold
	return l
}

func (l *GormLogger) WithContextKeys(keys ...any) *GormLogger {
	l.contextKeys = keys
	return l
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Info {
		LogWith(ctx, zapcore.InfoLevel, fmt.Sprintf(msg, data...), l.getCommonKV(ctx))
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		LogWith(ctx, zapcore.WarnLevel, fmt.Sprintf(msg, data...), l.getCommonKV(ctx))
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		LogWith(ctx, zapcore.ErrorLevel, fmt.Sprintf(msg, data...), l.getCommonKV(ctx))
	}
}

func (l *GormLogger) Trace(
	ctx context.Context,
	begin time.Time,
	f func() (string, int64),
	err error,
) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := f()

	kv := l.getCommonKV(ctx)
	kv["sql"] = sql
	kv["rows_affected"] = rows
	kv["elapsed_ac"] = elapsed.Milliseconds()

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		kv["error"] = err.Error()
		LogWith(ctx, zapcore.ErrorLevel, "sql operation: failed", kv)

	case elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		kv["slow_threshold_ms"] = l.SlowThreshold.Milliseconds()
		LogWith(ctx, zapcore.WarnLevel, "sql operation: slow query", kv)

	case l.LogLevel >= logger.Info:
		LogWith(ctx, zapcore.InfoLevel, "sql operation: executed", kv)
	}
}

func (l *GormLogger) getCommonKV(ctx context.Context) map[string]any {
	kv := make(map[string]any)
	kv["type"] = "database"

	for _, key := range l.contextKeys {
		if value := ctx.Value(key); value != nil {
			kv[fmt.Sprintf("%v", key)] = value
		}
	}

	return kv
}
