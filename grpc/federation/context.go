package federation

import (
	"context"
	"log/slog"

	"github.com/mercari/grpc-federation/grpc/federation/log"
)

type (
	celCacheMapKey struct{}
)

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return log.WithLogger(ctx, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	return log.Logger(ctx)
}

// SetLogger set logger instance for current context.
// This is intended to be called from a custom resolver and is currently propagated to the current context and its children.
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
