package tracing

import (
	"context"
)

// contextKey is a private type used for context keys
type contextKey string

// tracerKey is the context key for the tracer
const tracerKey = contextKey("tracer")

// WithTracer adds a tracer to the context
func WithTracer(ctx context.Context, tracer Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, tracer)
}

// GetTracer gets the tracer from the context
func GetTracer(ctx context.Context) Tracer {
	if tracer, ok := ctx.Value(tracerKey).(Tracer); ok {
		return tracer
	}
	return GetGlobalTracer()
}

// RecordEventContext records an event to the tracer in the context
func RecordEventContext(ctx context.Context, event Event) {
	GetTracer(ctx).RecordEvent(ctx, event)
}
