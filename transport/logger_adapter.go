package transport

import (
	"context"
	"time"
)

// loggerAdapter adapts a jira-connect Logger to transport.Logger
// This avoids circular dependencies between packages
type loggerAdapter struct {
	logger interface {
		Debug(ctx context.Context, msg string, fields ...interface{})
		Info(ctx context.Context, msg string, fields ...interface{})
		Warn(ctx context.Context, msg string, fields ...interface{})
		Error(ctx context.Context, msg string, fields ...interface{})
	}
}

func newLoggerAdapter(logger interface{}) Logger {
	// Check if logger already implements transport.Logger
	if tl, ok := logger.(Logger); ok {
		return tl
	}

	// Adapt the logger interface
	return &adaptedLogger{logger: logger}
}

// adaptedLogger wraps any logger that has Debug/Info/Warn/Error methods
type adaptedLogger struct {
	logger interface{}
}

func (a *adaptedLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	if debugLogger, ok := a.logger.(interface {
		Debug(ctx context.Context, msg string, fields ...interface{})
	}); ok {
		debugLogger.Debug(ctx, msg, convertToInterface(fields)...)
	}
}

func (a *adaptedLogger) Info(ctx context.Context, msg string, fields ...Field) {
	if infoLogger, ok := a.logger.(interface {
		Info(ctx context.Context, msg string, fields ...interface{})
	}); ok {
		infoLogger.Info(ctx, msg, convertToInterface(fields)...)
	}
}

func (a *adaptedLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	if warnLogger, ok := a.logger.(interface {
		Warn(ctx context.Context, msg string, fields ...interface{})
	}); ok {
		warnLogger.Warn(ctx, msg, convertToInterface(fields)...)
	}
}

func (a *adaptedLogger) Error(ctx context.Context, msg string, fields ...Field) {
	if errorLogger, ok := a.logger.(interface {
		Error(ctx context.Context, msg string, fields ...interface{})
	}); ok {
		errorLogger.Error(ctx, msg, convertToInterface(fields)...)
	}
}

func (a *adaptedLogger) With(fields ...Field) Logger {
	// For simplicity, return the same logger
	// A full implementation would create a child logger with fields
	return a
}

func convertToInterface(fields []Field) []interface{} {
	result := make([]interface{}, len(fields))
	for i, f := range fields {
		result[i] = struct {
			Key   string
			Value interface{}
		}{f.Key, f.Value}
	}
	return result
}

// Helper functions to create fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}
