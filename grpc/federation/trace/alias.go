package trace

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	KeyValue = attribute.KeyValue
)

var (
	WithAttributes = trace.WithAttributes
	StringAttr     = attribute.String
)
