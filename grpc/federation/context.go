//go:build !tinygo.wasm

package federation

import (
	"context"
	"log/slog"

	"github.com/mercari/grpc-federation/grpc/federation/cel"
)

type (
	loggerKey    struct{}
	celPluginKey struct{}
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

func WithCELPlugins(ctx context.Context, plugins []*cel.CELPlugin) context.Context {
	return context.WithValue(ctx, celPluginKey{}, plugins)
}

func CELPlugins(ctx context.Context) []*cel.CELPlugin {
	value := ctx.Value(celPluginKey{})
	if value == nil {
		return nil
	}
	return value.([]*cel.CELPlugin)
}
