package federation

import (
	"context"
	"log/slog"
)

type (
	loggerKey      struct{}
	celCacheMapKey struct{}
)

type loggerRef struct {
	logger *slog.Logger
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, &loggerRef{logger: logger})
}

func Logger(ctx context.Context) *slog.Logger {
	value := ctx.Value(loggerKey{})
	if value == nil {
		return slog.Default()
	}
	return value.(*loggerRef).logger
}

// SetLogger set logger instance for current context.
// This is intended to be called from a custom resolver and is currently propagated to the current context and its children.
func SetLogger(ctx context.Context, logger *slog.Logger) {
	value := ctx.Value(loggerKey{})
	if value == nil {
		return
	}
	value.(*loggerRef).logger = logger
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
