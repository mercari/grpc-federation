module example

go 1.21

replace github.com/mercari/grpc-federation => ../../

require (
	github.com/google/cel-go v0.18.2
	github.com/google/go-cmp v0.6.0
	github.com/mercari/grpc-federation v0.0.0-00010101000000-000000000000
	github.com/tetratelabs/wazero v1.6.0
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231211222908-989df2bf70f3
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.31.1-0.20231027082548-f4a6c1f6e5c1
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231212172506-995d672761c0 // indirect
)
