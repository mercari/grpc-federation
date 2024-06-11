package log

import (
	"context"
	"log/slog"
	"sync"
)

type (
	loggerKey struct{}
)

type loggerRef struct {
	mu     sync.RWMutex
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
	ref := value.(*loggerRef)

	ref.mu.RLock()
	defer ref.mu.RUnlock()
	return ref.logger
}

// SetLogger set logger instance for current context.
// This is intended to be called from a custom resolver and is currently propagated to the current context and its children.
func SetLogger(ctx context.Context, logger *slog.Logger) {
	value := ctx.Value(loggerKey{})
	if value == nil {
		return
	}
	ref := value.(*loggerRef)

	ref.mu.Lock()
	defer ref.mu.Unlock()
	ref.logger = logger
}
