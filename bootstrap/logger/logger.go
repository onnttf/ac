package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Directory string `json:"directory"`
	Level     string `json:"level"`
}

var (
	sugaredLogger *zap.SugaredLogger
	once          sync.Once
	initErr       error
)

func InitLogger(config Config) error {
	once.Do(func() {
		fmt.Fprintf(os.Stdout, "INFO: logger: init: started\n")

		// Get working directory
		workingDir, err := os.Getwd()
		if err != nil {
			initErr = fmt.Errorf("failed to get working directory for logger: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: logger: init: failed, reason=get working dir, error=%v\n", initErr)
			return
		}

		// Create log directory
		logDirectory := filepath.Join(workingDir, config.Directory)
		if err := os.MkdirAll(logDirectory, 0o744); err != nil {
			initErr = fmt.Errorf("failed to create log directory '%s': %w", logDirectory, err)
			fmt.Fprintf(os.Stderr, "ERROR: logger: init: failed, reason=create log dir, path=%s, error=%v\n", logDirectory, initErr)
			return
		}

		outputLevel, err := zapcore.ParseLevel(config.Level)
		if err != nil {
			initErr = fmt.Errorf("failed to parse log level '%s': %w", config.Level, err)
			fmt.Fprintf(
				os.Stderr,
				"ERROR: logger: init: failed, reason=parse log level, level=%s, error=%v\n",
				config.Level, initErr,
			)
			return
		}

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.LevelKey = "level"
		encoderConfig.MessageKey = "message"
		encoderConfig.StacktraceKey = "stack_trace"

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= outputLevel && lvl <= zapcore.InfoLevel
		})
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= outputLevel && lvl > zapcore.InfoLevel
		})

		lowLogFilePath := filepath.Join(logDirectory, "low.log")
		lowLogFile, err := os.OpenFile(
			lowLogFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0o644,
		)
		if err != nil {
			initErr = fmt.Errorf(
				"failed to open low priority log file '%s': %w",
				lowLogFilePath, err,
			)
			fmt.Fprintf(
				os.Stderr,
				"ERROR: logger: init: failed, reason=open low log file, path=%s, error=%v\n",
				lowLogFilePath, initErr,
			)
			return
		}

		highLogFilePath := filepath.Join(logDirectory, "high.log")
		highLogFile, err := os.OpenFile(
			highLogFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0o644,
		)
		if err != nil {
			initErr = fmt.Errorf(
				"failed to open high priority log file '%s': %w",
				highLogFilePath, err,
			)
			fmt.Fprintf(
				os.Stderr,
				"ERROR: logger: init: failed, reason=open high log file, path=%s, error=%v\n",
				highLogFilePath, initErr,
			)
			return
		}

		cores := []zapcore.Core{
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), lowPriority),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stderr), highPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(lowLogFile), lowPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(highLogFile), highPriority),
		}

		zapLogger := zap.New(
			zapcore.NewTee(cores...),
			zap.AddCaller(),
			zap.AddCallerSkip(2),
		)
		sugaredLogger = zapLogger.Sugar()

		fmt.Fprintf(
			os.Stdout,
			"INFO: logger: init: succeeded, level=%s, directory=%s, low_log_file=%s, high_log_file=%s\n",
			outputLevel.String(), logDirectory, lowLogFilePath, highLogFilePath,
		)
	})

	return initErr
}

func LogWith(ctx context.Context, level zapcore.Level, msg string, kv map[string]any) {
	baseKV := getBaseKV(ctx)
	for k, v := range kv {
		baseKV[k] = v
	}

	fields := make([]any, 0, len(baseKV)*2)
	for k, v := range baseKV {
		fields = append(fields, k, v)
	}

	if sugaredLogger == nil {
		var outputStream *os.File
		if level >= zapcore.WarnLevel {
			outputStream = os.Stderr
		} else {
			outputStream = os.Stdout
		}
		fmt.Fprintf(
			outputStream,
			"WARN: logger: log uninitialized: message='%s', level=%s, fields=%v\n",
			msg, level.String(), fields,
		)
		return
	}

	switch level {
	case zapcore.DebugLevel:
		sugaredLogger.Debugw(msg, fields...)
	case zapcore.InfoLevel:
		sugaredLogger.Infow(msg, fields...)
	case zapcore.WarnLevel:
		sugaredLogger.Warnw(msg, fields...)
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		sugaredLogger.Errorw(msg, fields...)
	default:
		sugaredLogger.Infow(msg, fields...)
	}
}

func getBaseKV(ctx context.Context) map[string]any {
	kv := make(map[string]any)
	if ctx == nil {
		return kv
	}

	if ginCtx, ok := ctx.(*gin.Context); ok {
		if requestId := requestid.Get(ginCtx); requestId != "" {
			kv["request_id"] = requestId
		}
	}
	return kv
}

func Debugf(ctx context.Context, format string, args ...any) {
	LogWith(ctx, zapcore.DebugLevel, fmt.Sprintf(format, args...), nil)
}

func Infof(ctx context.Context, format string, args ...any) {
	LogWith(ctx, zapcore.InfoLevel, fmt.Sprintf(format, args...), nil)
}

func Warnf(ctx context.Context, format string, args ...any) {
	LogWith(ctx, zapcore.WarnLevel, fmt.Sprintf(format, args...), nil)
}

func Errorf(ctx context.Context, format string, args ...any) {
	LogWith(ctx, zapcore.ErrorLevel, fmt.Sprintf(format, args...), nil)
}
