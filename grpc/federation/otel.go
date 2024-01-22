//go:build !tinygo.wasm

package federation

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func RecordErrorToSpan(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err, trace.WithStackTrace(true))
}
