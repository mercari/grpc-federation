module example

go 1.21

replace github.com/mercari/grpc-federation => ../../

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/google/cel-go v0.18.1
	github.com/mercari/grpc-federation v0.0.0-00010101000000-000000000000
	golang.org/x/sync v0.2.0
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
