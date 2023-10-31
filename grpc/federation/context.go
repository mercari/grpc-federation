package federation

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

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
