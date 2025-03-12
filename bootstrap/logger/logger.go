package logger

import (
	"ac/bootstrap/config"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level represents logging severity levels
type Level string

// Constants for log levels
const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

var (
	sugaredLogger *zap.SugaredLogger
	once          sync.Once
)

// LogRotationConfig Configuration for logger rotation
type LogRotationConfig struct {
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int // days
}

// defaultRotationConfig provides sensible defaults for log rotation
var defaultRotationConfig = LogRotationConfig{
	MaxSize:    500,
	MaxBackups: 3,
	MaxAge:     28,
}

// InitLogger initializes the logger with the provided configuration
func InitLogger() error {
	var initErr error
	once.Do(func() {
		// Get working directory
		workingDir, err := os.Getwd()
		if err != nil {
			initErr = fmt.Errorf("failed to get current working directory, err: %w", err)
			return
		}

		// Configure log directory
		logDirectory := filepath.Join(workingDir, config.Config.Log.Directory)
		if err := os.MkdirAll(logDirectory, 0744); err != nil {
			initErr = fmt.Errorf("failed to create log directory, err: %w", err)
			return
		}

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

		// Create encoders
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

		// Set minimum log level (can be made configurable)
		outputLevel := zapcore.InfoLevel

		// Define level enablers
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= outputLevel && lvl <= zapcore.InfoLevel
		})
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= outputLevel && lvl > zapcore.InfoLevel
		})

		// Configure log rotation for files
		lowLogger := &lumberjack.Logger{
			Filename:   filepath.Join(logDirectory, "low.log"),
			MaxSize:    defaultRotationConfig.MaxSize,
			MaxBackups: defaultRotationConfig.MaxBackups,
			MaxAge:     defaultRotationConfig.MaxAge,
			Compress:   true, // Enable compression for backups
		}

		highLogger := &lumberjack.Logger{
			Filename:   filepath.Join(logDirectory, "high.log"),
			MaxSize:    defaultRotationConfig.MaxSize,
			MaxBackups: defaultRotationConfig.MaxBackups,
			MaxAge:     7, // Keep high priority logs for less time
			Compress:   true,
		}

		// Create cores
		cores := []zapcore.Core{
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), lowPriority),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stderr), highPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(lowLogger), lowPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(highLogger), highPriority),
		}

		// Create the logger
		logger := zap.New(
			zapcore.NewTee(cores...),
			zap.AddCaller(),      // Add caller information
			zap.AddCallerSkip(1), // Skip the wrapper function in the stack
		)

		// Schedule cleanup on shutdown
		defer logger.Sync()

		// Store the sugared logger for convenience
		sugaredLogger = logger.Sugar()
	})

	// Return any initialization error
	return initErr
}

// LogWith records a log message with additional context
func LogWith(level Level, msg string, kv map[string]any) {
	if sugaredLogger == nil {
		// Auto-initialize with defaults if not done yet
		if err := InitLogger(); err != nil {
			fmt.Printf("failed to initialize logger, err: %s", err.Error())
			return
		}
	}

	// Extract key-value pairs more efficiently
	fields := make([]any, 0, len(kv)*2)
	for k, v := range kv {
		fields = append(fields, k, v)
	}

	// Log at appropriate level
	switch level {
	case LevelDebug:
		sugaredLogger.Debugw(msg, fields...)
	case LevelInfo:
		sugaredLogger.Infow(msg, fields...)
	case LevelWarn:
		sugaredLogger.Warnw(msg, fields...)
	case LevelError:
		sugaredLogger.Errorw(msg, fields...)
	default:
		sugaredLogger.Infow(msg, fields...)
	}
}

// getBaseKV extracts common context values for logging
func getBaseKV(ctx context.Context) map[string]any {
	if ctx == nil {
		return make(map[string]any)
	}

	kv := make(map[string]any)
	if val := ctx.Value("X-Request-Id"); val != nil {
		kv["request_id"] = val
	}
	return kv
}

func Infof(ctx context.Context, format string, args ...any) {
	LogWith(LevelInfo, fmt.Sprintf(format, args...), getBaseKV(ctx))
}

func Debugf(ctx context.Context, format string, args ...any) {
	LogWith(LevelDebug, fmt.Sprintf(format, args...), getBaseKV(ctx))
}

func Warnf(ctx context.Context, format string, args ...any) {
	LogWith(LevelWarn, fmt.Sprintf(format, args...), getBaseKV(ctx))
}

func Errorf(ctx context.Context, format string, args ...any) {
	LogWith(LevelError, fmt.Sprintf(format, args...), getBaseKV(ctx))
}
