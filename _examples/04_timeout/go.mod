module example

go 1.21

replace github.com/mercari/grpc-federation => ../../

require (
	github.com/mercari/grpc-federation v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240304212257-790db918fca8
	google.golang.org/grpc v1.62.0
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.20.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/planetscale/vtprotobuf v0.6.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/tetratelabs/wazero v1.6.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240304212257-790db918fca8 // indirect
)
