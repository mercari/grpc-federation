module example

go 1.21

replace github.com/mercari/grpc-federation => ../../

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/mercari/grpc-federation v0.0.0-00010101000000-000000000000
	golang.org/x/sync v0.2.0
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.2-0.20220831092852-f930b1dc76e8
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
)
