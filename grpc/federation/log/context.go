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
	attrs  []slog.Attr
}

func WithLogger(ctx context.Context, logger *slog.Logger, attrs ...slog.Attr) context.Context {
	return context.WithValue(ctx, loggerKey{}, &loggerRef{logger: logger, attrs: attrs})
}

func Logger(ctx context.Context) *slog.Logger {
	value := ctx.Value(loggerKey{})
	if value == nil {
		return slog.Default()
	}
	ref := value.(*loggerRef)

	ref.mu.RLock()
	defer ref.mu.RUnlock()
	return ref.logger.With(AttrsToArgs(ref.attrs)...)
}

func Attrs(ctx context.Context) []slog.Attr {
	value := ctx.Value(loggerKey{})
	if value == nil {
		return nil
	}
	ref := value.(*loggerRef)

	ref.mu.RLock()
	defer ref.mu.RUnlock()
	return ref.attrs
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

func AddAttrs(ctx context.Context, attrs []slog.Attr) {
	value := ctx.Value(loggerKey{})
	if value == nil {
		return
	}
	ref := value.(*loggerRef)

	ref.mu.Lock()
	defer ref.mu.Unlock()
	ref.attrs = append(ref.attrs, attrs...)
}

func AttrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return args
}
