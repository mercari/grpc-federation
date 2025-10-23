module example

go 1.24.9

replace github.com/mercari/grpc-federation => ../../

require (
	github.com/google/go-cmp v0.7.0
	github.com/mercari/grpc-federation v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/goleak v1.3.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240826202546-f6391c0de4c7
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2
)

require (
	cel.dev/expr v0.19.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/wasi-go v0.3.1 // indirect
	github.com/google/cel-go v0.23.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/stealthrocket/wazergo v0.19.1 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240826202546-f6391c0de4c7 // indirect
)
