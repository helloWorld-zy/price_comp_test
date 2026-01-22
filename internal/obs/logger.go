package obs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

// LogLevel represents log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogConfig holds logger configuration
type LogConfig struct {
	Level  LogLevel
	Format string // "json" or "text"
	Output io.Writer
}

// Logger wraps slog.Logger with additional context
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new structured logger
func NewLogger(config LogConfig) *Logger {
	var level slog.Level
	switch config.Level {
	case LogLevelDebug:
		level = slog.LevelDebug
	case LogLevelInfo:
		level = slog.LevelInfo
	case LogLevelWarn:
		level = slog.LevelWarn
	case LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return &Logger{Logger: slog.New(handler)}
}

// WithContext returns a logger with context attributes
func (l *Logger) WithContext(ctx context.Context) *Logger {
	attrs := []any{}

	// Extract trace ID from context
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		attrs = append(attrs, "trace_id", traceID)
	}

	// Extract user ID from context
	if userID := GetUserIDFromContext(ctx); userID > 0 {
		attrs = append(attrs, "user_id", userID)
	}

	if len(attrs) == 0 {
		return l
	}

	return &Logger{Logger: l.With(attrs...)}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{Logger: l.With(key, value)}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]any) *Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return &Logger{Logger: l.With(attrs...)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.With("error", err.Error())}
}

// WithDuration adds a duration field to the logger
func (l *Logger) WithDuration(d time.Duration) *Logger {
	return &Logger{Logger: l.With("duration_ms", d.Milliseconds())}
}

// Context keys for logger
type ctxKey string

const (
	ctxKeyTraceID ctxKey = "trace_id"
	ctxKeyUserID  ctxKey = "user_id"
)

// GetTraceIDFromContext retrieves trace ID from context
func GetTraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(ctxKeyTraceID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(ctx context.Context) uint64 {
	if v := ctx.Value(ctxKeyUserID); v != nil {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxKeyTraceID, traceID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

// Default logger instance
var defaultLogger = NewLogger(LogConfig{Level: LogLevelInfo, Format: "json"})

// Default returns the default logger
func Default() *Logger {
	return defaultLogger
}

// SetDefault sets the default logger
func SetDefault(l *Logger) {
	defaultLogger = l
}
