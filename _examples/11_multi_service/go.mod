module example

go 1.21

replace github.com/mercari/grpc-federation => ../../

require (
	github.com/google/cel-go v0.18.1
	github.com/google/go-cmp v0.6.0
	github.com/mercari/grpc-federation v0.0.0-00010101000000-000000000000
	golang.org/x/sync v0.4.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231012201019-e917dd12ba7a
	google.golang.org/grpc v1.57.1
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231002182017-d307bd883b97 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)