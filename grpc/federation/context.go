package federation

import (
	"context"
	"log/slog"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/mercari/grpc-federation/grpc/federation/log"
	"github.com/mercari/grpc-federation/grpc/federation/trace"
)

type (
	celCacheMapKey struct{}
)

func WithLogger(ctx context.Context, logger *slog.Logger, attrs ...slog.Attr) context.Context {
	return log.WithLogger(ctx, logger, attrs...)
}

func Logger(ctx context.Context) *slog.Logger {
	return log.Logger(ctx)
}

func WithTracer(ctx context.Context, tracer oteltrace.Tracer) context.Context {
	return trace.WithTracer(ctx, tracer)
}

func Tracer(ctx context.Context) oteltrace.Tracer {
	return trace.Tracer(ctx)
}

func LogAttrs(ctx context.Context) []slog.Attr {
	return log.Attrs(ctx)
}

func SetLogger(ctx context.Context, logger *slog.Logger) {
	log.SetLogger(ctx, logger)
}

func WithCELCacheMap(ctx context.Context, celCacheMap *CELCacheMap) context.Context {
	return context.WithValue(ctx, celCacheMapKey{}, celCacheMap)
}

func getCELCacheMap(ctx context.Context) *CELCacheMap {
	value := ctx.Value(celCacheMapKey{})
	if value == nil {
		return nil
	}
	return value.(*CELCacheMap)
}
