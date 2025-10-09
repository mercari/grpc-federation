package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type (
	tracerKey struct{}
)

func WithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerKey{}, tracer)
}

func Tracer(ctx context.Context) trace.Tracer {
	value := ctx.Value(tracerKey{})
	if value == nil {
		return otel.Tracer("default")
	}
	return value.(trace.Tracer)
}

func Trace(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer(ctx).Start(ctx, name)
}
