// Package bolt provides a bolt logger adapter for jira-connect.
//
// This adapter allows you to use the bolt structured logging library
// with jira-connect for zero-allocation logging and OpenTelemetry integration.
//
// Example usage:
//
//	logger := bolt.New(bolt.NewJSONHandler(os.Stdout))
//	client, err := jiraconnect.NewClient(
//		jiraconnect.WithBaseURL("https://your-domain.atlassian.net"),
//		jiraconnect.WithAPIToken("email", "token"),
//		jiraconnect.WithLogger(boltadapter.NewAdapter(logger)),
//	)
package bolt

import (
	"context"
	"time"

	jira "github.com/felixgeelhaar/jira-connect"
	"github.com/felixgeelhaar/bolt"
)

// Adapter adapts a bolt logger to the jira-connect Logger interface
type Adapter struct {
	logger *bolt.Logger
}

// NewAdapter creates a new bolt logger adapter
func NewAdapter(logger *bolt.Logger) *Adapter {
	return &Adapter{logger: logger}
}

// Debug logs a debug-level message with fields
func (a *Adapter) Debug(ctx context.Context, msg string, fields ...jira.Field) {
	event := a.logger.Debug()
	event = a.addFields(event, fields)
	event.Msg(msg)
}

// Info logs an info-level message with fields
func (a *Adapter) Info(ctx context.Context, msg string, fields ...jira.Field) {
	event := a.logger.Info()
	event = a.addFields(event, fields)
	event.Msg(msg)
}

// Warn logs a warning-level message with fields
func (a *Adapter) Warn(ctx context.Context, msg string, fields ...jira.Field) {
	event := a.logger.Warn()
	event = a.addFields(event, fields)
	event.Msg(msg)
}

// Error logs an error-level message with fields
func (a *Adapter) Error(ctx context.Context, msg string, fields ...jira.Field) {
	event := a.logger.Error()
	event = a.addFields(event, fields)
	event.Msg(msg)
}

// With creates a child logger with the given fields
func (a *Adapter) With(fields ...jira.Field) jira.Logger {
	// Build context with fields using With() which returns an Event
	event := a.logger.With()
	event = a.addFields(event, fields)
	childLogger := event.Logger()
	return &Adapter{logger: childLogger}
}

// addFields adds fields to a bolt event
func (a *Adapter) addFields(event *bolt.Event, fields []jira.Field) *bolt.Event {
	for _, f := range fields {
		event = a.addField(event, f)
	}
	return event
}

// addField adds a single field to a bolt event
func (a *Adapter) addField(event *bolt.Event, f jira.Field) *bolt.Event {
	switch v := f.Value.(type) {
	case string:
		return event.Str(f.Key, v)
	case int:
		return event.Int(f.Key, v)
	case int64:
		return event.Int64(f.Key, v)
	case bool:
		return event.Bool(f.Key, v)
	case time.Duration:
		return event.Dur(f.Key, v)
	case error:
		if v != nil {
			return event.Err(v)
		}
		return event
	case nil:
		return event.Str(f.Key, "")
	default:
		return event.Any(f.Key, v)
	}
}
