package federation

import (
	"context"
	"log/slog"
)

type (
	loggerKey      struct{}
	celCacheMapKey struct{}
)

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	value := ctx.Value(loggerKey{})
	if value == nil {
		return slog.Default()
	}
	return value.(*slog.Logger)
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
