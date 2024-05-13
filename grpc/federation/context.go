package federation

import (
	"context"
	"log/slog"
)

type (
	loggerKey              struct{}
	customResolverValueKey struct{}
	celCacheMapKey         struct{}
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

// SetLogger set logger instance for current context.
// This is intended to be called from a custom resolver and is currently propagated to the current context and its children.
func SetLogger(ctx context.Context, logger *slog.Logger) {
	value := ctx.Value(customResolverValueKey{})
	if value == nil {
		return
	}
	value.(*CustomResolverValue).Logger = logger
}

// CustomResolverValue type for storing references to be modified when calling a custom resolver.
type CustomResolverValue struct {
	Logger *slog.Logger
}

func WithCustomResolverValue(ctx context.Context) context.Context {
	return context.WithValue(ctx, customResolverValueKey{}, &CustomResolverValue{
		Logger: Logger(ctx),
	})
}

func GetCustomResolverValue(ctx context.Context) *CustomResolverValue {
	value := ctx.Value(customResolverValueKey{})
	if value == nil {
		return &CustomResolverValue{Logger: Logger(ctx)}
	}
	return value.(*CustomResolverValue)
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
