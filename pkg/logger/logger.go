package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with service-specific fields
type Logger struct {
	*zap.Logger
	service string
	env     string
	version string
}

// New creates a new Logger instance
func New(service, env, version string) (*Logger, error) {
	var cfg zap.Config
	if env == "development" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		cfg.DisableStacktrace = true
	}
	cfg.DisableCaller = false
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.StacktraceKey = "stacktrace"

	l, err := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger:  l,
		service: service,
		env:     env,
		version: version,
	}, nil
}

type contextKey string

const (
	traceIDKey   contextKey = "trace_id"
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// WithTraceID adds trace_id field from context when available.
func (l *Logger) WithTraceID(ctx context.Context) *Logger {
	traceID, _ := ctx.Value(traceIDKey).(string)
	if traceID == "" {
		return l
	}
	return l.With(zap.String("trace_id", traceID))
}

// WithRequestID adds request_id field from context when available.
func (l *Logger) WithRequestID(ctx context.Context) *Logger {
	requestID, _ := ctx.Value(requestIDKey).(string)
	if requestID == "" {
		return l
	}
	return l.With(zap.String("request_id", requestID))
}

// WithUserID adds user_id field from context when available.
func (l *Logger) WithUserID(ctx context.Context) *Logger {
	userID, _ := ctx.Value(userIDKey).(string)
	if userID == "" {
		return l
	}
	return l.With(zap.String("user_id", userID))
}

// WithService adds service, env, version fields to the logger
func (l *Logger) WithService() *Logger {
	return l.With(
		zap.String("service", l.service),
		zap.String("env", l.env),
		zap.String("version", l.version),
	)
}

// Info logs a message at InfoLevel
func (l *Logger) Info(msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.String("service", l.service),
		zap.String("env", l.env),
		zap.String("version", l.version),
	)
	l.Logger.Info(msg, fields...)
}

// Error logs a message at ErrorLevel
func (l *Logger) Error(msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.String("service", l.service),
		zap.String("env", l.env),
		zap.String("version", l.version),
	)
	l.Logger.Error(msg, fields...)
}

// Warn logs a message at WarnLevel
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.String("service", l.service),
		zap.String("env", l.env),
		zap.String("version", l.version),
	)
	l.Logger.Warn(msg, fields...)
}

// Debug logs a message at DebugLevel
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.String("service", l.service),
		zap.String("env", l.env),
		zap.String("version", l.version),
	)
	l.Logger.Debug(msg, fields...)
}

// With creates a child logger and adds additional fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger:  l.Logger.With(fields...),
		service: l.service,
		env:     l.env,
		version: l.version,
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
