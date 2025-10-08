package jiraconnect

import (
	"context"
	"time"
)

// Logger is the interface for structured logging.
// This allows users to provide their own logger implementation or use the provided bolt adapter.
type Logger interface {
	// Debug logs a debug-level message with fields
	Debug(ctx context.Context, msg string, fields ...Field)

	// Info logs an info-level message with fields
	Info(ctx context.Context, msg string, fields ...Field)

	// Warn logs a warning-level message with fields
	Warn(ctx context.Context, msg string, fields ...Field)

	// Error logs an error-level message with fields
	Error(ctx context.Context, msg string, fields ...Field)

	// With creates a child logger with the given fields
	With(fields ...Field) Logger
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Err creates an error field
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// noopLogger is a logger that does nothing
type noopLogger struct{}

func (n *noopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}
func (n *noopLogger) Info(ctx context.Context, msg string, fields ...Field)  {}
func (n *noopLogger) Warn(ctx context.Context, msg string, fields ...Field)  {}
func (n *noopLogger) Error(ctx context.Context, msg string, fields ...Field) {}
func (n *noopLogger) With(fields ...Field) Logger                            { return n }

// NewNoopLogger creates a logger that does nothing.
// This is the default logger used when no logger is provided.
func NewNoopLogger() Logger {
	return &noopLogger{}
}
